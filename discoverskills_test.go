package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSkillsCommonPaths(t *testing.T) {
	root := makeTestSkillsRoot(t)

	writeSkillFile(t, root, "valid-skill",
		"name: Valid Skill\n"+
			"description: Useful description\n")

	writeSkillFile(t, root, "invalid-skill",
		"name: Invalid Skill\n",
	)

	missingFileDir := filepath.Join(root, "missing-file")
	if err := os.Mkdir(missingFileDir, 0o755); err != nil {
		t.Fatalf("create missing-file dir: %v", err)
	}

	results := map[string]*skillParseResult{}
	for result := range discoverSkills(root) {
		results[result.Path] = result
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 discoverable results, got %d", len(results))
	}

	valid := results["valid-skill"]
	if valid == nil {
		t.Fatal("expected valid-skill result")
	}
	if valid.error != nil {
		t.Fatalf("expected valid-skill without error, got %v", valid.error)
	}
	if valid.Name != "Valid Skill" {
		t.Fatalf("expected valid name %q, got %q", "Valid Skill", valid.Name)
	}
	if valid.Description != "Useful description" {
		t.Fatalf("expected valid description %q, got %q", "Useful description", valid.Description)
	}

	invalid := results["invalid-skill"]
	if invalid == nil {
		t.Fatal("expected invalid-skill result")
	}
	if invalid.error == nil {
		t.Fatal("expected invalid-skill error")
	}
	if invalid.error.Error() != "parse frontmatter: missing description" {
		t.Fatalf("expected invalid-skill error %q, got %q", "parse frontmatter: missing description", invalid.error.Error())
	}
}

func TestDiscoverSkillsReportsRootOpenError(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing-skills-root")

	results := collectDiscoverSkillsResults(discoverSkills(root))
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Path != "" {
		t.Fatalf("expected empty path, got %q", result.Path)
	}
	if result.error == nil {
		t.Fatal("expected root open error")
	}
	if !errors.Is(result.error, os.ErrNotExist) {
		t.Fatalf("expected wrapped not-exist error, got %v", result.error)
	}
}

func TestOpenDir(t *testing.T) {
	t.Run("opens directory", func(t *testing.T) {
		root := makeTestSkillsRoot(t)

		dir, err := openDir(root)
		if err != nil {
			t.Fatalf("open directory: %v", err)
		}
		if dir == nil {
			t.Fatal("expected directory handle")
		}
		dir.Close()
	})

	t.Run("rejects non directory", func(t *testing.T) {
		root := makeTestSkillsRoot(t)
		filePath := filepath.Join(root, "file.txt")
		if err := os.WriteFile(filePath, []byte("content"), 0o644); err != nil {
			t.Fatalf("write test file: %v", err)
		}

		dir, err := openDir(filePath)
		if err == nil {
			if dir != nil {
				dir.Close()
			}
			t.Fatal("expected non-directory error")
		}
		if err.Error() != "not a directory" {
			t.Fatalf("expected non-directory error, got %v", err)
		}
	})

	t.Run("missing path", func(t *testing.T) {
		missingPath := filepath.Join(t.TempDir(), "missing")

		dir, err := openDir(missingPath)
		if err == nil {
			if dir != nil {
				dir.Close()
			}
			t.Fatal("expected missing-path error")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected wrapped not-exist error, got %v", err)
		}
	})
}

func collectDiscoverSkillsResults(results <-chan *skillParseResult) []*skillParseResult {
	collected := make([]*skillParseResult, 0)
	for result := range results {
		collected = append(collected, result)
	}

	return collected
}
