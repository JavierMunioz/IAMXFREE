// Package execution defines how a registered application can be installed,
// built, started, stopped, restarted and updated, without committing to any
// particular technology or execution mechanism — no os/exec, no systemd, no
// Docker, no process signals. Those arrive in later iterations. This
// package only defines three things: the Strategy contract every
// technology-specific implementation (Node+npm, Python+uv, Docker Compose,
// systemd, ...) will satisfy, a Registry strategies register themselves
// into, and a Resolver that picks the right Strategy for a given
// Application without a growing if/else chain.
package execution
