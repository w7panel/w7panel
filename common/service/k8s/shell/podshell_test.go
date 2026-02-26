package shell

import "testing"

func TestRunLs(t *testing.T) {
	shell := NewPodShell("ls", "/home", "/home", []string{}, "1", "0")
	err := shell.RunLs()
	if err != nil {
		t.Errorf("RunLs() failed: %v", err)
	}
}
