package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const manifestFileName = "skills.manifest"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) > 1 {
		return errors.New("this command does not accept arguments")
	}

	skillsRoot, err := resolveSkillsRoot()
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(skillsRoot, manifestFileName)
	manifestFile, err := os.OpenFile(manifestPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open manifest: %w", err)
	}
	defer manifestFile.Close()
	writer := bufio.NewWriter(manifestFile)

	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)

	for skill := range discoverSkills(skillsRoot) {
		if skill.error != nil {
			if skill.Path == "" {
				fmt.Fprintf(os.Stderr, "Error: %v\n", skill.error)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: parse skill %v: %v\n", skill.Path, skill.error)
			}
			continue
		}

		if err := encoder.Encode(skill); err != nil {
			return fmt.Errorf("write skill %v to manifest: %w", skill.Path, err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush manifest: %w", err)
	}

	fmt.Println(manifestPath)
	return nil
}

func resolveSkillsRoot() (string, error) {
	configured := strings.TrimSpace(os.Getenv("MY_SKILLS_PATH"))
	if configured == "" {
		return os.Getwd()
	}

	return filepath.Abs(configured)
}
