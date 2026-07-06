// Package runtimehosttest provides FakeHost, a deterministic test double
// for runtimehost.Host. It is a regular (non-_test.go) package so it can be
// imported by any other package's tests — most notably future execution
// strategies, which will need to test their behavior against a Host
// without depending on what tools happen to be installed on the machine
// running the tests.
package runtimehosttest
