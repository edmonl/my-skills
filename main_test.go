package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFrontmatterValidYAML(t *testing.T) {
	t.Parallel()

	path := writeSkillFile(t, "---\nname: example\ndescription: >-\n  Example skill\n---\n# Body\n")
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	frontmatter, err := parseFrontmatter(file)
	if err != nil {
		t.Fatalf("parseFrontmatter returned error: %v", err)
	}

	if got, want := frontmatter.Name, "example"; got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}

	if got, want := frontmatter.Description, "Example skill"; got != want {
		t.Fatalf("description = %q, want %q", got, want)
	}
}

func TestParseFrontmatterRejectsMissingDelimiters(t *testing.T) {
	t.Parallel()

	path := writeSkillFile(t, "name: example\n")
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	_, err = parseFrontmatter(file)
	if err == nil {
		t.Fatal("parseFrontmatter unexpectedly succeeded")
	}
}

func TestLoadSkillManifestRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillDir := filepath.Join(root, "example")
	if err := os.Mkdir(skillDir, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	content := "---\nname: example\ndescription: Example skill\nextra: no\n---\n"
	if err := os.WriteFile(filepath.Join(skillDir, skillFileName), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := loadSkillManifest(filepath.Join(skillDir, skillFileName), "example")
	if err == nil {
		t.Fatal("loadSkillManifest unexpectedly succeeded")
	}
}

func TestLoadSkillManifestSupportsQuotedColons(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillDir := filepath.Join(root, "example")
	if err := os.Mkdir(skillDir, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	content := "---\nname: 'skill:example'\ndescription: \"Handles notes: tasks and ideas\"\n---\n"
	path := filepath.Join(skillDir, skillFileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entry, err := loadSkillManifest(path, "example")
	if err != nil {
		t.Fatalf("loadSkillManifest returned error: %v", err)
	}

	if got, want := entry.Name, "skill:example"; got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}

	if got, want := entry.Description, "Handles notes: tasks and ideas"; got != want {
		t.Fatalf("description = %q, want %q", got, want)
	}
}

func TestDiscoverSkillsOnlyUsesImmediateSubdirectories(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	validDir := filepath.Join(root, "valid")
	if err := os.Mkdir(validDir, 0o755); err != nil {
		t.Fatalf("Mkdir valid: %v", err)
	}

	validSkill := "---\nname: valid\ndescription: Valid skill\n---\n"
	if err := os.WriteFile(filepath.Join(validDir, skillFileName), []byte(validSkill), 0o644); err != nil {
		t.Fatalf("WriteFile valid: %v", err)
	}

	nestedDir := filepath.Join(root, "parent", "nested")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("MkdirAll nested: %v", err)
	}

	nestedSkill := "---\nname: nested\ndescription: Nested skill\n---\n"
	if err := os.WriteFile(filepath.Join(nestedDir, skillFileName), []byte(nestedSkill), 0o644); err != nil {
		t.Fatalf("WriteFile nested: %v", err)
	}

	var entries []skillManifestEntry
	err := discoverSkills(root, func(entry skillManifestEntry) error {
		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		t.Fatalf("discoverSkills returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("discoverSkills returned %d entries, want 1", len(entries))
	}

	if got, want := entries[0].Path, "valid"; got != want {
		t.Fatalf("entry path = %q, want %q", got, want)
	}
}

func TestWriteManifestWritesJSONLines(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manifestPath := filepath.Join(root, manifestFileName)
	skills := map[string]string{
		"alpha": "---\nname: alpha\ndescription: Alpha skill\n---\n",
		"beta":  "---\nname: beta\ndescription: Beta skill\n---\n",
	}

	for name, content := range skills {
		skillDir := filepath.Join(root, name)
		if err := os.Mkdir(skillDir, 0o755); err != nil {
			t.Fatalf("Mkdir %s: %v", name, err)
		}

		if err := os.WriteFile(filepath.Join(skillDir, skillFileName), []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile %s: %v", name, err)
		}
	}

	if err := writeManifest(manifestPath, root); err != nil {
		t.Fatalf("writeManifest returned error: %v", err)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("manifest line count = %d, want 2", len(lines))
	}

	manifest := strings.Join(lines, "\n")

	if !strings.Contains(manifest, `"name":"alpha"`) {
		t.Fatalf("manifest missing alpha entry: %s", manifest)
	}

	if !strings.Contains(manifest, `"path":"beta"`) {
		t.Fatalf("manifest missing beta entry: %s", manifest)
	}
}

func writeSkillFile(t *testing.T, contents string) string {
	t.Helper()

	root := t.TempDir()
	path := filepath.Join(root, skillFileName)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	return path
}
