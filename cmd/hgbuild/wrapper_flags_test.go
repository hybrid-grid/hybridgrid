package main

import (
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func resetWrapperGlobals(t *testing.T) {
	t.Helper()
	coordinator = ""
	noFallback = false
	t.Cleanup(func() {
		coordinator = ""
		noFallback = false
	})
}

func TestFilterHgbuildWrapperFlags_Coordinator(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantArgs        []string
		wantCoordinator string
	}{
		{
			name:            "long flag with equals",
			args:            []string{"--coordinator=host:9000", "-j8"},
			wantArgs:        []string{"-j8"},
			wantCoordinator: "host:9000",
		},
		{
			name:            "long flag space separated",
			args:            []string{"--coordinator", "host:9000", "-j8"},
			wantArgs:        []string{"-j8"},
			wantCoordinator: "host:9000",
		},
		{
			name:            "short -C is the tool's change-directory flag",
			args:            []string{"-C", "subdir", "-j8"},
			wantArgs:        []string{"-C", "subdir", "-j8"},
			wantCoordinator: "",
		},
		{
			name:            "short -C with an address-looking value is still the tool's",
			args:            []string{"-C", "host:9000", "-j8"},
			wantArgs:        []string{"-C", "host:9000", "-j8"},
			wantCoordinator: "",
		},
		{
			name:            "coordinator after tool args belongs to the tool",
			args:            []string{"-j8", "--coordinator=late:1"},
			wantArgs:        []string{"-j8", "--coordinator=late:1"},
			wantCoordinator: "",
		},
		{
			name:            "coordinator after no-fallback is still leading",
			args:            []string{"--no-fallback", "--coordinator=host:9000", "all"},
			wantArgs:        []string{"all"},
			wantCoordinator: "host:9000",
		},
		{
			name:            "everything after -- is untouched",
			args:            []string{"--", "--coordinator=x:1", "--no-fallback", "-C", "d"},
			wantArgs:        []string{"--", "--coordinator=x:1", "--no-fallback", "-C", "d"},
			wantCoordinator: "",
		},
		{
			name:            "flag before -- is consumed, flag after is not",
			args:            []string{"--coordinator=a:1", "--", "--coordinator=b:2"},
			wantArgs:        []string{"--", "--coordinator=b:2"},
			wantCoordinator: "a:1",
		},
		{
			name:            "trailing coordinator without a value is passed through",
			args:            []string{"--coordinator"},
			wantArgs:        []string{"--coordinator"},
			wantCoordinator: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetWrapperGlobals(t)

			got := filterHgbuildWrapperFlags(tt.args)
			if !reflect.DeepEqual(got, tt.wantArgs) {
				t.Errorf("filtered args = %q, want %q", got, tt.wantArgs)
			}
			if coordinator != tt.wantCoordinator {
				t.Errorf("coordinator = %q, want %q", coordinator, tt.wantCoordinator)
			}
		})
	}
}

func TestFilterHgbuildWrapperFlags_NoFallbackAfterDoubleDashUntouched(t *testing.T) {
	resetWrapperGlobals(t)

	got := filterHgbuildWrapperFlags([]string{"all", "--", "--no-fallback"})
	want := []string{"all", "--", "--no-fallback"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("filtered args = %q, want %q", got, want)
	}
	if noFallback {
		t.Error("--no-fallback after -- must not be consumed")
	}
}

func TestWrapCommandArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCmd  string
		wantArgs []string
		wantErr  bool
	}{
		{name: "plain", args: []string{"cmake", "--build", "."}, wantCmd: "cmake", wantArgs: []string{"--build", "."}},
		{name: "leading separator dropped", args: []string{"--", "env", "A=1"}, wantCmd: "env", wantArgs: []string{"A=1"}},
		{name: "later separator kept", args: []string{"sh", "--", "x"}, wantCmd: "sh", wantArgs: []string{"--", "x"}},
		{name: "empty", args: nil, wantErr: true},
		{name: "only separator", args: []string{"--"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args, err := wrapCommandArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if cmd != tt.wantCmd {
				t.Errorf("command = %q, want %q", cmd, tt.wantCmd)
			}
			if !tt.wantErr && !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("args = %q, want %q", args, tt.wantArgs)
			}
		})
	}
}

// fakeTools installs make, ninja and faketool shims on PATH that record the
// HG_COORDINATOR they inherit and the arguments they receive.
func fakeTools(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/sh shims")
	}

	dir := t.TempDir()
	out := filepath.Join(dir, "record.txt")
	script := "#!/bin/sh\n" +
		"{ printf 'coord=%s\\n' \"$HG_COORDINATOR\"; for a in \"$@\"; do printf 'arg=%s\\n' \"$a\"; done; } > \"$HG_TEST_RECORD\"\n"
	for _, name := range []string{"make", "ninja", "faketool"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
			t.Fatalf("failed to write shim %s: %v", name, err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HG_TEST_RECORD", out)
	return out
}

func readRecord(t *testing.T, path string) (string, []string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("wrapped tool did not run: %v", err)
	}
	defer f.Close()

	var coord string
	var args []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "coord="):
			coord = strings.TrimPrefix(line, "coord=")
		case strings.HasPrefix(line, "arg="):
			args = append(args, strings.TrimPrefix(line, "arg="))
		}
	}
	return coord, args
}

// TestWrapperSubcommands_CoordinatorRouting runs the real cobra tree against
// shim tools to pin down where a coordinator may be given and what the
// wrapped tool finally receives.
func TestWrapperSubcommands_CoordinatorRouting(t *testing.T) {
	tests := []struct {
		name            string
		env             string // HG_COORDINATOR in the caller's environment
		args            []string
		wantCoordinator string
		wantArgs        []string
	}{
		// make
		{
			name:            "make: --coordinator= before subcommand",
			args:            []string{"--coordinator=host:9000", "make", "-j8"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "make: --coordinator space-separated before subcommand",
			args:            []string{"--coordinator", "host:9000", "make", "-j8"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "make: -C before subcommand is the coordinator",
			args:            []string{"-C", "host:9000", "make", "-j8"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "make: --coordinator= after subcommand",
			args:            []string{"make", "--coordinator=host:9000", "-j8"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "make: --coordinator space-separated after subcommand",
			args:            []string{"make", "--coordinator", "host:9000", "-j8"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "make: -C dir after subcommand is passed to make",
			args:            []string{"make", "-C", "subdir", "-j8"},
			wantCoordinator: "",
			wantArgs:        []string{"-C", "subdir", "-j8"},
		},
		{
			name:            "make: coordinator before and -C dir after",
			args:            []string{"-C", "host:9000", "make", "-C", "subdir", "-j8"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"-C", "subdir", "-j8"},
		},
		{
			name:            "make: no flag inherits environment",
			env:             "env:1",
			args:            []string{"make", "-j8"},
			wantCoordinator: "env:1",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "make: flag before subcommand beats environment",
			env:             "env:1",
			args:            []string{"--coordinator=flag:2", "make", "-j8"},
			wantCoordinator: "flag:2",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "make: flag after subcommand beats environment",
			env:             "env:1",
			args:            []string{"make", "--coordinator=flag:2", "-j8"},
			wantCoordinator: "flag:2",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "make: -- passes everything after it to make",
			args:            []string{"make", "--", "--coordinator=x:1", "all"},
			wantCoordinator: "",
			wantArgs:        []string{"--", "--coordinator=x:1", "all"},
		},
		// ninja
		{
			name:            "ninja: --coordinator= before subcommand",
			args:            []string{"--coordinator=host:9000", "ninja", "-j8"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "ninja: -C before subcommand is the coordinator",
			args:            []string{"-C", "host:9000", "ninja", "-j8"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "ninja: --coordinator after subcommand",
			args:            []string{"ninja", "--coordinator", "host:9000", "-j8"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"-j8"},
		},
		{
			name:            "ninja: -C dir after subcommand is passed to ninja",
			args:            []string{"ninja", "-C", "build", "all"},
			wantCoordinator: "",
			wantArgs:        []string{"-C", "build", "all"},
		},
		// wrap
		{
			name:            "wrap: --coordinator= before subcommand",
			args:            []string{"--coordinator=host:9000", "wrap", "faketool", "a", "b"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"a", "b"},
		},
		{
			name:            "wrap: -C before subcommand is the coordinator",
			args:            []string{"-C", "host:9000", "wrap", "faketool", "a"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"a"},
		},
		{
			name:            "wrap: --coordinator= after subcommand",
			args:            []string{"wrap", "--coordinator=host:9000", "faketool", "a"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"a"},
		},
		{
			name:            "wrap: --coordinator space-separated after subcommand",
			args:            []string{"wrap", "--coordinator", "host:9000", "faketool", "a"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"a"},
		},
		{
			name:            "wrap: -- separator before the command",
			args:            []string{"wrap", "--coordinator=host:9000", "--", "faketool", "a"},
			wantCoordinator: "host:9000",
			wantArgs:        []string{"a"},
		},
		{
			name:            "wrap: -C and --coordinator after the command belong to it",
			args:            []string{"wrap", "faketool", "-C", "dir", "--coordinator=x:1"},
			wantCoordinator: "",
			wantArgs:        []string{"-C", "dir", "--coordinator=x:1"},
		},
		{
			name:            "wrap: everything after -- following the command is untouched",
			args:            []string{"wrap", "faketool", "--", "-C", "dir", "--coordinator=x:1"},
			wantCoordinator: "",
			wantArgs:        []string{"--", "-C", "dir", "--coordinator=x:1"},
		},
		{
			name:            "wrap: flag beats environment",
			env:             "env:1",
			args:            []string{"--coordinator=flag:2", "wrap", "faketool"},
			wantCoordinator: "flag:2",
			wantArgs:        nil,
		},
		{
			name:            "wrap: no flag inherits environment",
			env:             "env:1",
			args:            []string{"wrap", "faketool"},
			wantCoordinator: "env:1",
			wantArgs:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := fakeTools(t)
			t.Setenv("HG_COORDINATOR", tt.env)
			t.Setenv(noFallbackEnv, "")
			resetWrapperGlobals(t)

			root := newRootCmd()
			root.SetArgs(tt.args)
			if err := root.Execute(); err != nil {
				t.Fatalf("Execute(%q) failed: %v", tt.args, err)
			}

			gotCoord, gotArgs := readRecord(t, record)
			if gotCoord != tt.wantCoordinator {
				t.Errorf("wrapped tool saw HG_COORDINATOR=%q, want %q", gotCoord, tt.wantCoordinator)
			}

			// make receives CC=/CXX= overrides ahead of the user's arguments.
			if len(tt.args) > 0 && containsArg(tt.args, "make") {
				if len(gotArgs) < 2 || !strings.HasPrefix(gotArgs[0], "CC=") || !strings.HasPrefix(gotArgs[1], "CXX=") {
					t.Fatalf("make args missing CC/CXX overrides: %q", gotArgs)
				}
				gotArgs = gotArgs[2:]
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("wrapped tool args = %q, want %q", gotArgs, tt.wantArgs)
			}
		})
	}
}
