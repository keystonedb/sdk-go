package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/keystonedb/sdk-go/keystone"
)

// Translator implements keystone.Translator using the Anthropic Claude API.
type Translator struct {
	client anthropic.Client
	model  anthropic.Model
}

// Option configures a Translator.
type Option func(*Translator)

// WithModel sets the Claude model to use for translations.
func WithModel(model anthropic.Model) Option {
	return func(t *Translator) {
		t.model = model
	}
}

// NewTranslator creates a new Anthropic-backed Translator.
// The default model is Claude Haiku 4.5, which is ideal for structured translation tasks.
func NewTranslator(apiKey string, opts ...Option) *Translator {
	return newTranslator(append(opts, WithModel("")), option.WithAPIKey(apiKey))
}

func newTranslator(opts []Option, reqOpts ...option.RequestOption) *Translator {
	t := &Translator{
		client: anthropic.NewClient(reqOpts...),
		model:  anthropic.ModelClaudeHaiku4_5,
	}
	for _, o := range opts {
		o(t)
	}
	if t.model == "" {
		t.model = anthropic.ModelClaudeHaiku4_5
	}
	return t
}

// translationResult maps the JSON output from Claude's tool call.
type translationResult struct {
	S string `json:"s"`
	P string `json:"p,omitempty"`
}

// Translate implements keystone.Translator by calling the Anthropic Messages API
// with tool_use to guarantee structured JSON output for all target languages in a single request.
func (t *Translator) Translate(ctx context.Context, req keystone.TranslateRequest) (*keystone.TranslateResponse, error) {
	if len(req.TargetLanguages) == 0 {
		return &keystone.TranslateResponse{Translations: make(map[string]*keystone.Translation)}, nil
	}

	// Build the tool schema: each target language is a required property
	properties := make(map[string]any, len(req.TargetLanguages))
	for _, lang := range req.TargetLanguages {
		langSchema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"s": map[string]any{"type": "string", "description": "Singular form"},
				"p": map[string]any{"type": "string", "description": "Plural form"},
			},
			"required": []string{"s"},
		}
		if req.Plural == "" {
			langSchema = map[string]any{
				"type": "object",
				"properties": map[string]any{
					"s": map[string]any{"type": "string", "description": "Singular form"},
				},
				"required": []string{"s"},
			}
		}
		properties[lang] = langSchema
	}

	tool := anthropic.ToolUnionParam{
		OfTool: &anthropic.ToolParam{
			Name:        "translate",
			Description: anthropic.String("Provide translations for the requested text"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: properties,
				Required:   req.TargetLanguages,
			},
		},
	}

	// Build user message
	var userMsg string
	if req.Plural != "" {
		userMsg = fmt.Sprintf("Translate from %s:\nSingular: %s\nPlural: %s\nTarget languages: %s",
			req.SourceLanguage, req.Singular, req.Plural, strings.Join(req.TargetLanguages, ", "))
	} else {
		userMsg = fmt.Sprintf("Translate from %s:\nText: %s\nTarget languages: %s",
			req.SourceLanguage, req.Singular, strings.Join(req.TargetLanguages, ", "))
	}

	message, err := t.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:    t.model,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: "You are a professional translator. Translate text accurately, preserving meaning, tone, and register. Use natural phrasing in each target language."},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg)),
		},
		Tools: []anthropic.ToolUnionParam{tool},
		ToolChoice: anthropic.ToolChoiceUnionParam{
			OfAny: &anthropic.ToolChoiceAnyParam{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic translate: %w", err)
	}

	// Find the tool_use block in the response
	for _, block := range message.Content {
		if block.Type != "tool_use" {
			continue
		}
		toolUse := block.AsToolUse()

		var results map[string]translationResult
		if err := json.Unmarshal(toolUse.Input, &results); err != nil {
			return nil, fmt.Errorf("anthropic translate: parse tool response: %w", err)
		}

		resp := &keystone.TranslateResponse{Translations: make(map[string]*keystone.Translation, len(results))}
		for lang, result := range results {
			resp.Translations[lang] = &keystone.Translation{
				Singular: result.S,
				Plural:   result.P,
			}
		}
		return resp, nil
	}

	// No tool_use block found — return empty response
	return &keystone.TranslateResponse{Translations: make(map[string]*keystone.Translation)}, nil
}
