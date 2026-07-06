package runtimehost

import "testing"

func TestParseProcStatTimes(t *testing.T) {
	// A realistic /proc/<pid>/stat line: pid, comm (parenthesized, here
	// containing a space to exercise the ")"-based split), state, then
	// numeric fields. utime=field14, stime=field15, starttime=field22.
	line := "4242 (my app) S 1 4242 4242 0 -1 4194304 100 0 0 0 " +
		"111 222 0 0 20 0 1 0 355000 0 0 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 17 1 0 0 0 0 0"

	utime, stime, starttime, ok := parseProcStatTimes([]byte(line))
	if !ok {
		t.Fatal("expected parseProcStatTimes to succeed")
	}
	if utime != 111 {
		t.Errorf("utime = %d, want 111", utime)
	}
	if stime != 222 {
		t.Errorf("stime = %d, want 222", stime)
	}
	if starttime != 355000 {
		t.Errorf("starttime = %d, want 355000", starttime)
	}
}

func TestParseProcStatTimesMalformed(t *testing.T) {
	if _, _, _, ok := parseProcStatTimes([]byte("garbage without parens")); ok {
		t.Fatal("expected ok=false for a line with no comm parentheses")
	}
	if _, _, _, ok := parseProcStatTimes([]byte("1 (sh) S 1 1")); ok {
		t.Fatal("expected ok=false when too few fields follow the comm")
	}
}

func TestParseVmStatusKB(t *testing.T) {
	status := "Name:\tmy-app\nVmRSS:\t   12345 kB\nVmSize:\t   67890 kB\n"

	rss, ok := parseVmStatusKB([]byte(status), "VmRSS:")
	if !ok || rss != 12345 {
		t.Errorf("VmRSS = %d, %v, want 12345, true", rss, ok)
	}

	vsz, ok := parseVmStatusKB([]byte(status), "VmSize:")
	if !ok || vsz != 67890 {
		t.Errorf("VmSize = %d, %v, want 67890, true", vsz, ok)
	}

	if _, ok := parseVmStatusKB([]byte(status), "VmHWM:"); ok {
		t.Error("expected ok=false for a key not present in the status text")
	}
}
