package runtimehost

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// clockTicksPerSecond is the kernel's USER_HZ — the unit /proc/<pid>/stat
// reports CPU time in. It is effectively always 100 on Linux distributions
// IAMXFREE targets; there is no portable way to query it without cgo, so it
// is hardcoded rather than pulled in a C dependency for one constant.
const clockTicksPerSecond = 100

// ProcessResources queries /proc for pid's current resource usage. On a
// platform without /proc (e.g. a developer's Mac), or once pid no longer
// exists, every field is simply left unavailable — this is never reported
// as an error, matching every other "could not determine" outcome in this
// package.
func (h *LinuxHost) ProcessResources(pid int) (ProcessResources, error) {
	var res ProcessResources

	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid)); err == nil {
		if utimeTicks, stimeTicks, startTicks, ok := parseProcStatTimes(data); ok {
			if boot, ok := linuxBootTime(); ok {
				startedAt := boot.Add(time.Duration(startTicks) * time.Second / clockTicksPerSecond)
				if elapsed := time.Since(startedAt).Seconds(); elapsed > 0 {
					cpuSeconds := float64(utimeTicks+stimeTicks) / clockTicksPerSecond
					res.CPUPercent = (cpuSeconds / elapsed) * 100
					res.CPUPercentKnown = true
				}
			}
		}
	}

	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid)); err == nil {
		if kb, ok := parseVmStatusKB(data, "VmRSS:"); ok {
			res.MemoryRSSBytes = kb * 1024
			res.MemoryRSSKnown = true
		}
		if kb, ok := parseVmStatusKB(data, "VmSize:"); ok {
			res.MemoryVSZBytes = kb * 1024
			res.MemoryVSZKnown = true
		}
	}

	return res, nil
}

// linuxBootTime reads the system boot time (seconds since epoch) from
// /proc/stat's "btime" line, needed to turn a process's starttime (ticks
// since boot) into an absolute time.
func linuxBootTime() (time.Time, bool) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}, false
	}
	for _, line := range strings.Split(string(data), "\n") {
		seconds, ok := strings.CutPrefix(line, "btime ")
		if !ok {
			continue
		}
		value, err := strconv.ParseInt(strings.TrimSpace(seconds), 10, 64)
		if err != nil {
			return time.Time{}, false
		}
		return time.Unix(value, 0), true
	}
	return time.Time{}, false
}

// parseProcStatTimes extracts utime, stime (clock ticks spent in
// user/kernel mode) and starttime (clock ticks since boot) from the
// contents of /proc/<pid>/stat. Field 2 (comm) is wrapped in parentheses
// and may itself contain spaces or parentheses, so fields are located from
// the last ")" in the line rather than a naive space split.
func parseProcStatTimes(data []byte) (utime, stime, starttime uint64, ok bool) {
	s := string(data)
	closeParen := strings.LastIndex(s, ")")
	if closeParen == -1 || closeParen+2 > len(s) {
		return 0, 0, 0, false
	}

	// fields[0] is field 3 (state) in `man proc`'s 1-indexed numbering, so
	// field N there is fields[N-3] here.
	fields := strings.Fields(s[closeParen+2:])
	const utimeField, stimeField, starttimeField = 14, 15, 22
	if len(fields) <= starttimeField-3 {
		return 0, 0, 0, false
	}

	u, err1 := strconv.ParseUint(fields[utimeField-3], 10, 64)
	st, err2 := strconv.ParseUint(fields[stimeField-3], 10, 64)
	start, err3 := strconv.ParseUint(fields[starttimeField-3], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return u, st, start, true
}

// parseVmStatusKB finds the line in /proc/<pid>/status starting with key
// (e.g. "VmRSS:") and returns its value in kB.
func parseVmStatusKB(data []byte, key string) (uint64, bool) {
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, key) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, false
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, false
		}
		return kb, true
	}
	return 0, false
}
