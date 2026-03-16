package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

type skillFrontmatter struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

const maxFrontmatterBytes = 1 * 1024 * 1024

var targetKeys = map[string]struct{}{
	"name":        {},
	"description": {},
}

func parseFrontmatter(reader io.Reader) (*skillFrontmatter, error) {
	limitedReader := io.LimitReader(reader, maxFrontmatterBytes)
	decoder := yaml.NewDecoder(limitedReader)
	data := &skillFrontmatter{}

	err := parseFrontmatterDoc(decoder, data)
	if err == nil {
		return data, nil
	}

	return nil, fmt.Errorf("parse frontmatter: %w", err)
}

func parseFrontmatterDoc(decoder *yaml.Decoder, data *skillFrontmatter) error {
	var root yaml.Node
	if err := decoder.Decode(&root); err != nil {
		return err
	}

	if len(root.Content) <= 0 {
		return errors.New("empty YAML document")
	}

	if root.Content[0].Kind != yaml.MappingNode {
		return errors.New("not a map")
	}

	mapping := root.Content[0]
	results := map[string]string{}

	// Iterate through Content in pairs: [keyNode, valueNode, keyNode, valueNode]
	for i := 0; i < len(mapping.Content); i += 2 {
		key := mapping.Content[i].Value
		if _, ok := targetKeys[key]; !ok {
			continue
		}

		valueNode := mapping.Content[i+1]
		if valueNode.Kind != yaml.ScalarNode || valueNode.Tag != "!!str" {
			return fmt.Errorf("invalid value for %v", key)
		}

		value := strings.TrimSpace(valueNode.Value)
		if value == "" {
			return fmt.Errorf("empty value for %v", key)
		}

		results[key] = value
		if len(results) == len(targetKeys) {
			data.Name = results["name"]
			data.Description = results["description"]
			return nil
		}
	}

	var err error
	for key := range targetKeys {
		if _, ok := results[key]; !ok {
			err = fmt.Errorf("missing %v", key)
			break
		}
	}

	return err
}
