package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	skillFileName    = "SKILL.md"
	readDirBatchSize = 32
	maxWorkers       = 10
)

type skillParseResult struct {
	skillFrontmatter
	Path  string `json:"path"`
	error error
}

func discoverSkills(skillsRoot string) <-chan *skillParseResult {
	subdirs := make(chan string, maxWorkers)
	skills := make(chan *skillParseResult, maxWorkers)

	var workers sync.WaitGroup
	// fire up workers first
	worker := func() {
		for subdir := range subdirs {
			r := &skillParseResult{Path: subdir}
			file, err := os.Open(filepath.Join(skillsRoot, subdir, skillFileName))
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					r.error = fmt.Errorf("open %v: %w", skillFileName, err)
					skills <- r
				}
				continue
			}

			fm, err := parseFrontmatter(file)
			file.Close()
			if err == nil {
				r.skillFrontmatter = *fm
			} else {
				r.error = err
			}
			skills <- r
		}
	}

	for range maxWorkers {
		workers.Go(worker)
	}

	// start producer
	go func() {
		dir, err := openDir(skillsRoot)
		if err != nil {
			close(subdirs)
			workers.Wait()
			skills <- &skillParseResult{error: fmt.Errorf("open skills path: %w", err)}
			close(skills)
			return
		}

		defer func() {
			dir.Close()
			close(subdirs)
			workers.Wait()
			close(skills)
		}()

		for {
			dirEntries, err := dir.ReadDir(readDirBatchSize)
			for _, dirEntry := range dirEntries {
				if dirEntry.IsDir() {
					subdirs <- dirEntry.Name()
				}
			}

			if err != nil {
				if !errors.Is(err, io.EOF) {
					skills <- &skillParseResult{error: fmt.Errorf("read skills path: %w", err)}
				}
				break
			}
		}
	}()

	return skills
}

func openDir(path string) (*os.File, error) {
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	dirInfo, err := dir.Stat()
	if err != nil {
		dir.Close()
		return nil, err
	}
	if !dirInfo.IsDir() {
		dir.Close()
		return nil, errors.New("not a directory")
	}

	return dir, nil
}
