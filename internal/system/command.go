package system

import (
	"os/exec"
	"strings"
)

// Run starts one program, waits for it, and returns everything it wrote on
// both streams together with whether it exited successfully.
//
// It is here rather than beside its caller because starting a process is
// reaching the runtime, in the same sense as reading the clock: what it finds
// is whatever the machine happens to hold. harness_test.go refuses os/exec
// outside this package for that reason, and the gate takes this function as a
// dependency so its own suite can decide what a command returned instead of
// running one.
//
// Both streams are combined rather than kept apart. What a caller does with
// the output is print it under the leg that failed, and a tool that writes its
// findings on one stream and its summary on the other reads wrong when the two
// are separated.
func Run(name string, args ...string) (string, error) {
	command := exec.Command(name, args...)
	output, err := command.CombinedOutput()
	return string(output), err
}

// Look reports whether a program can be found on the path. It answers the
// question a gate leg asks before it decides that it cannot run rather than
// that it failed, and those are different verdicts.
func Look(name string) bool {
	path, err := exec.LookPath(name)
	return err == nil && strings.TrimSpace(path) != ""
}
