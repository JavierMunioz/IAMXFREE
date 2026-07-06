// Package runtimehost is the only place in IAMXFREE allowed to talk to the
// operating system: running commands, checking whether a tool is on PATH,
// reading its version, and checking whether paths exist. Execution
// strategies (and anything else that needs the OS) depend on the Host
// interface this package defines — never the other way around, and nothing
// outside this package should import os/exec or similar directly.
//
// LinuxHost is the concrete implementation used in production, on the
// Linux VPS IAMXFREE manages. Tests should use runtimehosttest.FakeHost
// instead, so they never depend on what tools happen to be installed on
// the machine running them.
package runtimehost
