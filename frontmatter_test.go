package main

import (
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        *skillFrontmatter
		wantErr     string
	}{
		{
			name: "valid frontmatter with extra fields",
			input: "name: Example Skill\n" +
				"description: Useful description\n" +
				"path: ignored\n" +
				"extra:\n" +
				"  nested: value\n",
			want: &skillFrontmatter{
				Name:        "Example Skill",
				Description: "Useful description",
			},
		},
		{
			name: "missing description",
			input: "name: Example Skill\n",
			wantErr: "parse frontmatter: missing description",
		},
		{
			name: "non mapping document",
			input: "- name: Example Skill\n- description: Useful description\n",
			wantErr: "parse frontmatter: not a map",
		},
		{
			name: "non string value",
			input: "name:\n  nested: value\n" +
				"description: Useful description\n",
			wantErr: "parse frontmatter: invalid value for name",
		},
		{
			name: "empty value after trimming",
			input: "name: '   '\n" +
				"description: Useful description\n",
			wantErr: "parse frontmatter: empty value for name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFrontmatter(strings.NewReader(tt.input))
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got == nil {
				t.Fatal("expected metadata, got nil")
			}

			if got.Name != tt.want.Name {
				t.Fatalf("expected Name %q, got %q", tt.want.Name, got.Name)
			}

			if got.Description != tt.want.Description {
				t.Fatalf("expected Description %q, got %q", tt.want.Description, got.Description)
			}
		})
	}
}
