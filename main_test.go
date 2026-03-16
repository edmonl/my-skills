package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveSkillsRootUsesWorkingDirectoryWhenEnvUnset(t *testing.T) {
	t.Setenv("MY_SKILLS_PATH", "")

	tempDir := t.TempDir()
	withWorkingDir(t, tempDir)

	got, err := resolveSkillsRoot()
	if err != nil {
		t.Fatalf("resolve skills root: %v", err)
	}

	if got != tempDir {
		t.Fatalf("expected %q, got %q", tempDir, got)
	}
}

func TestResolveSkillsRootUsesAbsoluteConfiguredPath(t *testing.T) {
	tempDir := t.TempDir()
	configured := filepath.Join(tempDir, "..", filepath.Base(tempDir))
	t.Setenv("MY_SKILLS_PATH", "  "+configured+"  ")

	got, err := resolveSkillsRoot()
	if err != nil {
		t.Fatalf("resolve skills root: %v", err)
	}

	want, err := filepath.Abs(configured)
	if err != nil {
		t.Fatalf("resolve expected absolute path: %v", err)
	}

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRunRejectsArguments(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"my-skills", "unexpected"}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	err := run()
	if err == nil {
		t.Fatal("expected argument error")
	}
	if err.Error() != "this command does not accept arguments" {
		t.Fatalf("expected argument error, got %v", err)
	}
}

func TestRunWritesManifestAndWarnsForInvalidSkill(t *testing.T) {
	skillsRoot := makeTestSkillsRoot(t)
	writeSkillFile(t, skillsRoot, "valid-skill",
		"name: Valid Skill\n"+
			"description: Useful description\n")
	writeSkillFile(t, skillsRoot, "invalid-skill",
		"name: Invalid Skill\n")

	workDir := t.TempDir()
	withWorkingDir(t, workDir)
	t.Setenv("MY_SKILLS_PATH", skillsRoot)

	originalArgs := os.Args
	os.Args = []string{"my-skills"}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	restoreStderr := captureStderr(t)
	err := run()
	stderrOutput := restoreStderr()
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	manifestPath := filepath.Join(workDir, manifestFileName)
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(manifestBytes)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 manifest line, got %d", len(lines))
	}

	var entry skillParseResult
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("unmarshal manifest entry: %v", err)
	}
	if entry.Path != "valid-skill" {
		t.Fatalf("expected manifest path %q, got %q", "valid-skill", entry.Path)
	}
	if entry.Name != "Valid Skill" {
		t.Fatalf("expected manifest name %q, got %q", "Valid Skill", entry.Name)
	}
	if entry.Description != "Useful description" {
		t.Fatalf("expected manifest description %q, got %q", "Useful description", entry.Description)
	}

	if !strings.Contains(stderrOutput, "Warning: parse skill invalid-skill: parse frontmatter: missing description\n") {
		t.Fatalf("expected invalid skill warning, got %q", stderrOutput)
	}
}

func TestRunReturnsErrorWhenSkillsRootCannotBeOpened(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)
	t.Setenv("MY_SKILLS_PATH", filepath.Join(workDir, "missing-root"))

	originalArgs := os.Args
	os.Args = []string{"my-skills"}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	restoreStderr := captureStderr(t)
	err := run()
	stderrOutput := restoreStderr()
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stderrOutput, "Error: open skills path: ") {
		t.Fatalf("expected root error on stderr, got %q", stderrOutput)
	}

	manifestPath := filepath.Join(workDir, manifestFileName)
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if len(manifestBytes) != 0 {
		t.Fatalf("expected empty manifest for unreadable root, got %q", string(manifestBytes))
	}
}
