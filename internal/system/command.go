package system

import (
	"os"
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

// Read returns the bytes of one file. It is here for the same reason Run is:
// what it finds is whatever the machine happens to hold, so the gate takes it
// as a dependency and its suite answers from a table instead of from a disk.
//
// The alternative was `git cat-file blob` once per file, which is correct and
// unusable: measured on this repository's 70 tracked files on 2026-08-10, on
// windows/amd64, that spelling took 1m53s for one leg because every file costs
// a process, and a gate a contributor will not wait for is a gate nobody runs.
//
//	go build -o gate.exe . && time ./gate.exe ci encoding
//	PASS  encoding
//	      Examined 70 tracked text file(s) for UTF-8 validity.
//	real    1m53,543s
func Read(name string) (string, error) {
	content, err := os.ReadFile(name)
	return string(content), err
}

// Look reports whether a program can be found on the path. It answers the
// question a gate leg asks before it decides that it cannot run rather than
// that it failed, and those are different verdicts.
func Look(name string) bool {
	path, err := exec.LookPath(name)
	return err == nil && strings.TrimSpace(path) != ""
}
