// Voice Transcriber
// Copyright (c) 2025 Ihor Dvoretskyi
//
// Licensed under MIT License

// Package e2e contains end-to-end tests for the voice-transcriber binary.
// Tests build the binary once in TestMain and exercise it as a subprocess,
// covering CLI help/version output, argument validation, exit codes, and
// GCP error-wrapping paths — all without requiring real GCP credentials.
package e2e_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// CLI tokens used across multiple test cases.
const (
	flagHelp      = "--help"
	cmdTranscribe = "transcribe"
)

// binaryPath holds the path to the compiled binary built by TestMain.
var binaryPath string

// TestMain builds the binary once for the whole suite, then runs all tests.
func TestMain(m *testing.M) {
	bin, cleanup, err := buildBinary()
	if err != nil {
		// Print directly — t.Fatal is unavailable in TestMain.
		_, _ = os.Stderr.WriteString("e2e: failed to build binary: " + err.Error() + "\n")

		os.Exit(1)
	}

	binaryPath = bin

	code := m.Run()

	cleanup()
	os.Exit(code)
}

// buildBinary compiles the voice-transcriber binary into a temp directory and
// returns the full path to the binary, a cleanup function, and any error.
func buildBinary() (string, func(), error) {
	dir, err := os.MkdirTemp("", "voice-transcriber-e2e-*")
	if err != nil {
		return "", func() {}, fmt.Errorf("creating temp dir: %w", err)
	}

	cleanup := func() { _ = os.RemoveAll(dir) }

	bin := filepath.Join(dir, "voice-transcriber")

	// Resolve the repo root relative to this source file so the build works
	// regardless of the directory from which `go test` is invoked.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", func() {}, fmt.Errorf("runtime.Caller failed")
	}

	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	//nolint:gosec // arguments are hard-coded, not user-controlled
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/voice-transcriber")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		cleanup()

		return "", func() {}, fmt.Errorf("building binary: %w", err)
	}

	return bin, cleanup, nil
}

// run executes the binary with the given arguments and optional env overrides,
// and returns the combined stdout, stderr, and exit code.
func run(t *testing.T, env []string, args ...string) (string, string, int) {
	t.Helper()

	//nolint:gosec // binaryPath is built by buildBinary above, args are test-controlled
	cmd := exec.Command(binaryPath, args...)

	if env != nil {
		cmd.Env = env
	}

	var outBuf, errBuf strings.Builder

	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()

	var exitCode int

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running binary: %v", err)
		}
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// TestVersion verifies the version subcommand outputs expected build metadata.
func TestVersion(t *testing.T) {
	t.Parallel()

	stdout, _, exitCode := run(t, nil, "version")

	if exitCode != 0 {
		t.Errorf("version: want exit 0, got %d", exitCode)
	}

	for _, want := range []string{"Voice Transcriber", "Build Date:", "Git Commit:"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("version stdout missing %q\ngot: %s", want, stdout)
		}
	}
}

// TestHelp verifies help output for the root command and subcommands.
func TestHelp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantOut  []string
		exitCode int
	}{
		{
			name:     "root no args",
			args:     nil,
			wantOut:  []string{"Multilingual media-to-text transcription", "--language", "--model", "--location"},
			exitCode: 0,
		},
		{
			name:     "root --help",
			args:     []string{flagHelp},
			wantOut:  []string{"Multilingual media-to-text transcription", "--language", "--model", "--location"},
			exitCode: 0,
		},
		{
			name:     "transcribe --help",
			args:     []string{cmdTranscribe, flagHelp},
			wantOut:  []string{"Transcribe a video or audio file", "--output"},
			exitCode: 0,
		},
		{
			name:     "version --help",
			args:     []string{"version", flagHelp},
			wantOut:  []string{"Show version information"},
			exitCode: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stdout, _, exitCode := run(t, nil, tc.args...)

			if exitCode != tc.exitCode {
				t.Errorf("exit code: want %d, got %d", tc.exitCode, exitCode)
			}

			for _, want := range tc.wantOut {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout missing %q\ngot: %s", want, stdout)
				}
			}
		})
	}
}

// TestArgValidation verifies cobra reports errors correctly for bad invocations.
func TestArgValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "unknown subcommand",
			args:    []string{"foo"},
			wantErr: `unknown command "foo"`,
		},
		{
			name:    "unknown root flag",
			args:    []string{"--badarg"},
			wantErr: "unknown flag: --badarg",
		},
		{
			name:    "transcribe missing argument",
			args:    []string{cmdTranscribe},
			wantErr: "accepts 1 arg(s), received 0",
		},
		{
			name:    "transcribe too many arguments",
			args:    []string{cmdTranscribe, "a", "b"},
			wantErr: "accepts 1 arg(s), received 2",
		},
		{
			name:    "transcribe unknown flag",
			args:    []string{cmdTranscribe, "--badarg"},
			wantErr: "unknown flag: --badarg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, stderr, exitCode := run(t, nil, tc.args...)

			if exitCode != 1 {
				t.Errorf("exit code: want 1, got %d", exitCode)
			}

			if !strings.Contains(stderr, tc.wantErr) {
				t.Errorf("stderr missing %q\ngot: %s", tc.wantErr, stderr)
			}
		})
	}
}

// TestFlagDefaults verifies that default flag values are baked into the binary.
func TestFlagDefaults(t *testing.T) {
	t.Parallel()

	stdout, _, _ := run(t, nil, cmdTranscribe, flagHelp)

	tests := []struct {
		name string
		want string
	}{
		{
			name: "language default is auto",
			want: `(default "auto")`,
		},
		{
			name: "model default is gemini-3.1-flash-lite-preview",
			want: "gemini-3.1-flash-lite-preview",
		},
		{
			name: "location default is global",
			want: `(default "global")`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if !strings.Contains(stdout, tc.want) {
				t.Errorf("transcribe --help missing %q\ngot: %s", tc.want, stdout)
			}
		})
	}
}

// TestGCPErrorPaths verifies the error-wrapping chain through the full stack
// without requiring real GCP credentials.
func TestGCPErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("no gcloud no project env", func(t *testing.T) {
		t.Parallel()

		// Strip PATH entirely so gcloud cannot be found, and ensure
		// GOOGLE_CLOUD_PROJECT is not set.
		env := []string{
			"PATH=/usr/bin:/bin",
			"HOME=" + os.Getenv("HOME"),
		}

		_, stderr, exitCode := run(t, env, cmdTranscribe, "nonexistent.mp4")

		if exitCode != 1 {
			t.Errorf("exit code: want 1, got %d", exitCode)
		}

		if !strings.Contains(stderr, "failed to resolve GCP project ID") {
			t.Errorf("stderr missing 'failed to resolve GCP project ID'\ngot: %s", stderr)
		}
	})

	t.Run("project set but no credentials", func(t *testing.T) {
		t.Parallel()

		// Provide a project ID so the gcloud lookup is skipped, but point
		// GOOGLE_APPLICATION_CREDENTIALS at a nonexistent file so the
		// Vertex AI client creation fails.
		env := []string{
			"PATH=" + os.Getenv("PATH"),
			"HOME=" + os.Getenv("HOME"),
			"GOOGLE_CLOUD_PROJECT=fake-project-for-e2e-test",
			"GOOGLE_APPLICATION_CREDENTIALS=/nonexistent/credentials.json",
		}

		_, stderr, exitCode := run(t, env, cmdTranscribe, "nonexistent.mp4")

		if exitCode != 1 {
			t.Errorf("exit code: want 1, got %d", exitCode)
		}

		if !strings.Contains(stderr, "initialization failed") {
			t.Errorf("stderr missing 'initialization failed'\ngot: %s", stderr)
		}
	})
}
