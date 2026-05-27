package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_HelpListsReplay(t *testing.T) {
	root := NewRootCmd("test")
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out.String(), "replay") {
		t.Errorf("--help output missing 'replay' subcommand:\n%s", out.String())
	}
}

func TestRootCmd_VersionFlag(t *testing.T) {
	root := NewRootCmd("9.9.9")
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out.String(), "9.9.9") {
		t.Errorf("--version output missing 9.9.9:\n%s", out.String())
	}
}

func TestReplay_HelpLists_CommitsFlag(t *testing.T) {
	root := NewRootCmd("test")
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"replay", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	body := out.String()
	for _, want := range []string{"--commits", "--branch", "-n"} {
		if !strings.Contains(body, want) {
			t.Errorf("replay --help missing %q:\n%s", want, body)
		}
	}
}
