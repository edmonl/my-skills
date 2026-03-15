package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	skillFileName    = "SKILL.md"
	manifestFileName = "skills.manifest"
	readDirBatchSize = 32
	maxWorkers       = 10
)

type skillManifestEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

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
	if err := writeManifest(manifestPath, skillsRoot); err != nil {
		return err
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

func discoverSkills(skillsRoot string, visit func(skillManifestEntry) error) error {
	dir, err := os.Open(skillsRoot)
	if err != nil {
		return fmt.Errorf("open skills directory: %w", err)
	}
	defer dir.Close()

	type result struct {
		entry     skillManifestEntry
		skillPath string
		err       error
		missing   bool
	}

	jobs := make(chan string)
	results := make(chan result, maxWorkers)
	var workers sync.WaitGroup
	workers.Add(maxWorkers)

	for range maxWorkers {
		go func() {
			defer workers.Done()

			for job := range jobs {
				skillPath := filepath.Join(skillsRoot, job, skillFileName)
				entry, err := loadSkillManifest(skillPath, job)
				if err != nil {
					results <- result{
						skillPath: skillPath,
						err:       err,
						missing:   errors.Is(err, os.ErrNotExist),
					}
					continue
				}

				results <- result{entry: entry}
			}
		}()
	}

	readErrs := make(chan error, 1)
	go func() {
		defer close(jobs)

		for {
			dirEntries, err := dir.ReadDir(readDirBatchSize)
			if err != nil && !errors.Is(err, io.EOF) {
				readErrs <- fmt.Errorf("read skills directory: %w", err)
				return
			}

			for _, dirEntry := range dirEntries {
				if !dirEntry.IsDir() {
					continue
				}

				jobs <- dirEntry.Name()
			}

			if errors.Is(err, io.EOF) {
				readErrs <- nil
				return
			}
		}
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	var visitErr error
	for result := range results {
		if result.err != nil {
			if !result.missing {
				fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", result.skillPath, result.err)
			}
			continue
		}

		if visitErr != nil {
			continue
		}

		if err := visit(result.entry); err != nil {
			visitErr = err
		}
	}

	if err := <-readErrs; err != nil {
		return err
	}

	return visitErr
}

func loadSkillManifest(skillPath string, folderName string) (skillManifestEntry, error) {
	file, err := os.Open(skillPath)
	if err != nil {
		return skillManifestEntry{}, err
	}
	defer file.Close()

	frontmatter, err := parseFrontmatter(file)
	if err != nil {
		return skillManifestEntry{}, err
	}

	frontmatter.Path = folderName
	return frontmatter, nil
}

func parseFrontmatter(reader io.Reader) (skillManifestEntry, error) {
	block, err := extractFrontmatterBlock(reader)
	if err != nil {
		return skillManifestEntry{}, err
	}

	var raw map[string]any
	if err := yaml.Unmarshal(block, &raw); err != nil {
		return skillManifestEntry{}, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}

	if len(raw) == 0 {
		return skillManifestEntry{}, errors.New("frontmatter is empty")
	}

	if len(raw) != 2 {
		return skillManifestEntry{}, errors.New("frontmatter must contain only name and description")
	}

	name, err := requireStringField(raw, "name")
	if err != nil {
		return skillManifestEntry{}, err
	}

	description, err := requireStringField(raw, "description")
	if err != nil {
		return skillManifestEntry{}, err
	}

	return skillManifestEntry{
		Name:        name,
		Description: description,
	}, nil
}

func extractFrontmatterBlock(reader io.Reader) ([]byte, error) {
	scanner := bufio.NewScanner(reader)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)

	if !scanner.Scan() {
		return nil, errors.New("skill file is empty")
	}

	if strings.TrimPrefix(scanner.Text(), "\ufeff") != "---" {
		return nil, errors.New(`frontmatter must start with "---"`)
	}

	var block strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("scan skill file: %w", err)
			}
			return []byte(block.String()), nil
		}
		block.WriteString(line)
		block.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan skill file: %w", err)
	}

	return nil, errors.New(`frontmatter must end with "---"`)
}

func requireStringField(raw map[string]any, key string) (string, error) {
	value, ok := raw[key]
	if !ok {
		return "", fmt.Errorf(`missing required frontmatter field %q`, key)
	}

	asString, ok := value.(string)
	if !ok {
		return "", fmt.Errorf(`frontmatter field %q must be a string`, key)
	}

	asString = strings.TrimSpace(asString)
	if asString == "" {
		return "", fmt.Errorf(`frontmatter field %q must not be empty`, key)
	}

	return asString, nil
}

func writeManifest(manifestPath string, skillsRoot string) error {
	file, err := os.Create(manifestPath)
	if err != nil {
		return fmt.Errorf("create manifest: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)

	if err := discoverSkills(skillsRoot, func(entry skillManifestEntry) error {
		if err := encoder.Encode(entry); err != nil {
			return fmt.Errorf("write manifest entry: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
