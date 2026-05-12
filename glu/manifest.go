package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/boxesandglue/glu/internal/errkind"
	"github.com/boxesandglue/glu/markdown"
)

// manifest is the JSON shape written when --manifest is passed.
// Stable, versioned schema — bump SchemaVersion on breaking changes.
type manifest struct {
	SchemaVersion string         `json:"schema_version"`
	GluVersion    string         `json:"glu_version"`
	Input         string         `json:"input"`
	Output        string         `json:"output"`
	Pages         int            `json:"pages"`
	Passes        int            `json:"passes"`
	DurationMs    int64          `json:"duration_ms"`
	GeneratedAt   string         `json:"generated_at"`
	Headings      []manifestHead `json:"headings"`
}

type manifestHead struct {
	Level string `json:"level"`
	Text  string `json:"text"`
	Page  int    `json:"page"`
}

// writeManifest serialises a manifest JSON file. result may be nil if
// the pipeline didn't produce statistics (e.g. .lua entry point), in
// which case pages/passes/headings stay at zero values.
func writeManifest(path, input, output string, elapsed time.Duration, result *markdown.Result) error {
	m := manifest{
		SchemaVersion: "1",
		GluVersion:    Version,
		Input:         input,
		Output:        output,
		DurationMs:    elapsed.Milliseconds(),
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if result != nil {
		m.Pages = result.Pages
		m.Passes = result.Passes
		m.Headings = make([]manifestHead, len(result.Headings))
		for i, h := range result.Headings {
			m.Headings[i] = manifestHead{Level: h.Level, Text: h.Text, Page: h.Page}
		}
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: marshalling manifest: %s", errkind.IO, err.Error())
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("%w: writing manifest: %s", errkind.IO, err.Error())
	}
	return nil
}
