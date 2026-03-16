package main

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var unsafePathChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func makeTestSkillsRoot(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	projectDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get project dir: %v", err)
	}

	rootName := unsafePathChars.ReplaceAllString(projectDir, "_")
	testName := unsafePathChars.ReplaceAllString(t.Name(), "_")
	root := filepath.Join(tempDir, "skills-root-"+rootName+"-"+testName)
	if err := os.RemoveAll(root); err != nil {
		t.Fatalf("clear skills root dir: %v", err)
	}
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("create skills root dir: %v", err)
	}
	t.Cleanup(func() {
		if !t.Failed() {
			_ = os.RemoveAll(root)
		}
	})

	return root
}

func writeSkillFile(t *testing.T, root string, skillName string, contents string) {
	t.Helper()

	skillDir := filepath.Join(root, skillName)
	if err := os.Mkdir(skillDir, 0o755); err != nil {
		t.Fatalf("create skill dir %q: %v", skillName, err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, skillFileName), []byte(contents), 0o644); err != nil {
		t.Fatalf("write skill file %q: %v", skillName, err)
	}
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}

func captureStderr(t *testing.T) func() string {
	t.Helper()

	original := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stderr pipe: %v", err)
	}
	os.Stderr = writer

	return func() string {
		_ = writer.Close()
		os.Stderr = original

		output, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("read stderr: %v", err)
		}
		_ = reader.Close()
		return string(output)
	}
}
