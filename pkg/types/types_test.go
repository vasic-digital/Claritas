package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemPromptValidateValid(t *testing.T) {
	opts := SystemPrompt{
		Model:       "gpt-4",
		ExtractedAt: "test",
		Category:    "test",
		Company:     "OpenAI",
		ID:          "test-id-123",
		Confidence:  0.95,
		PromptText:  "test prompttext",
		Product:     "ChatGPT",
		Source:      "test",
		Tags:        []string{"test"},
	}
	assert.NoError(t, opts.Validate())
}

func TestSystemPromptValidateEmpty(t *testing.T) {
	opts := SystemPrompt{}
	err := opts.Validate()
	assert.Error(t, err)
}

func TestPromptEntryValidateValid(t *testing.T) {
	opts := PromptEntry{
		Category:   "test",
		Company:    "OpenAI",
		ID:         "test-id-123",
		Prompt:     "test prompt",
		Confidence: 0.95,
		Product:    "ChatGPT",
		Date:       "test",
		Source:     "test",
		Tags:       []string{"test"},
	}
	assert.NoError(t, opts.Validate())
}

func TestPromptEntryValidateEmpty(t *testing.T) {
	opts := PromptEntry{}
	err := opts.Validate()
	assert.Error(t, err)
}

func TestSearchOptionsValidateValid(t *testing.T) {
	opts := SearchOptions{
		Limit:         10,
		Categories:    []string{"test"},
		Query:         "test query",
		Tags:          []string{"test"},
		Companies:     []string{"test"},
		MinConfidence: 0.95,
	}
	assert.NoError(t, opts.Validate())
}

func TestSearchOptionsValidateEmpty(t *testing.T) {
	opts := SearchOptions{}
	err := opts.Validate()
	assert.Error(t, err)
}

func TestSearchOptionsDefaults(t *testing.T) {
	opts := SearchOptions{}
	opts.Query = "test"
	opts.Defaults()
	assert.Equal(t, 50, opts.Limit)
}

func TestSystemPromptValidateConfidenceRange(t *testing.T) {
	opts := SystemPrompt{Model: "gpt-4", Company: "O", ID: "test",
		Product: "P", Confidence: 1.5}
	assert.Error(t, opts.Validate())
	opts.Confidence = -0.1
	assert.Error(t, opts.Validate())
}

func TestPromptEntryValidateConfidenceRange(t *testing.T) {
	opts := PromptEntry{Company: "O", ID: "test", Prompt: "p",
		Product: "P", Confidence: 1.5}
	assert.Error(t, opts.Validate())
	opts.Confidence = -0.1
	assert.Error(t, opts.Validate())
}

func TestSearchOptionsValidateLimitNegative(t *testing.T) {
	opts := SearchOptions{Query: "test", Limit: -1}
	assert.Error(t, opts.Validate())
}
