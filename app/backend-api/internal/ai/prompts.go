package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// PromptTemplate holds a single prompt configuration loaded from YAML.
type PromptTemplate struct {
	Task           string  `yaml:"task"`
	Version        string  `yaml:"version"`
	Temperature    float64 `yaml:"temperature"`
	MaxTokens      int     `yaml:"max_tokens"`
	SystemPrompt   string  `yaml:"system_prompt"`
	UserPromptTmpl string  `yaml:"user_prompt_template"`
	OutputSchema   string  `yaml:"output_schema"`
}

// PromptRegistry holds all loaded prompt templates keyed by task name.
type PromptRegistry struct {
	templates map[string]*PromptTemplate
}

// LoadPromptRegistry loads all YAML files from the given directory.
// If the directory does not exist, an empty registry is returned (non-fatal).
func LoadPromptRegistry(dir string) (*PromptRegistry, error) {
	reg := &PromptRegistry{templates: make(map[string]*PromptTemplate)}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return reg, nil // directory absent — degrade gracefully
		}
		return nil, fmt.Errorf("read prompt dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read prompt file %s: %w", path, err)
		}
		var tmpl PromptTemplate
		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			return nil, fmt.Errorf("parse prompt file %s: %w", path, err)
		}
		reg.templates[tmpl.Task] = &tmpl
	}
	return reg, nil
}

// Get returns the prompt template for the given task name.
// Returns nil if not found.
func (r *PromptRegistry) Get(task string) *PromptTemplate {
	return r.templates[task]
}
