package anthropic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	sdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/keystonedb/sdk-go/keystone"
)

// toolUseResponse builds a mock Anthropic Messages API response containing a tool_use block.
func toolUseResponse(input any) map[string]any {
	inputJSON, _ := json.Marshal(input)
	return map[string]any{
		"id":    "msg_test",
		"type":  "message",
		"role":  "assistant",
		"model": "claude-haiku-4-5",
		"content": []map[string]any{
			{
				"type":  "tool_use",
				"id":    "toolu_test",
				"name":  "translate",
				"input": json.RawMessage(inputJSON),
			},
		},
		"stop_reason": "tool_use",
		"usage":       map[string]any{"input_tokens": 10, "output_tokens": 20},
	}
}

// textResponse builds a mock Anthropic Messages API response containing only a text block (no tool_use).
func textResponse() map[string]any {
	return map[string]any{
		"id":    "msg_test",
		"type":  "message",
		"role":  "assistant",
		"model": "claude-haiku-4-5",
		"content": []map[string]any{
			{
				"type": "text",
				"text": "I cannot translate that.",
			},
		},
		"stop_reason": "end_turn",
		"usage":       map[string]any{"input_tokens": 10, "output_tokens": 5},
	}
}

func newTestTranslator(serverURL string, opts ...Option) *Translator {
	return newTranslator(opts, option.WithAPIKey("test-key"), option.WithBaseURL(serverURL))
}

func TestTranslate_SingularOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)

		// Verify singular-only prompt doesn't mention "Plural"
		msgs := req["messages"].([]any)
		userContent := msgs[0].(map[string]any)["content"].([]any)
		text := userContent[0].(map[string]any)["text"].(string)
		if contains(text, "Plural") {
			t.Error("Singular-only request should not contain 'Plural' in the prompt")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toolUseResponse(map[string]any{
			"fr": map[string]any{"s": "Bonjour"},
		}))
	}))
	defer server.Close()

	tr := newTestTranslator(server.URL)
	resp, err := tr.Translate(context.Background(), keystone.TranslateRequest{
		SourceLanguage:  "en",
		Singular:        "Hello",
		TargetLanguages: []string{"fr"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Translations["fr"] == nil || resp.Translations["fr"].Singular != "Bonjour" {
		t.Errorf("expected fr.Singular='Bonjour', got %v", resp.Translations["fr"])
	}
	if resp.Translations["fr"].Plural != "" {
		t.Errorf("expected empty plural, got %q", resp.Translations["fr"].Plural)
	}
}

func TestTranslate_SingularAndPlural(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)

		// Verify plural is in the prompt
		msgs := req["messages"].([]any)
		userContent := msgs[0].(map[string]any)["content"].([]any)
		text := userContent[0].(map[string]any)["text"].(string)
		if !contains(text, "Plural: cats") {
			t.Errorf("Expected prompt to contain 'Plural: cats', got: %s", text)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toolUseResponse(map[string]any{
			"fr": map[string]any{"s": "chat", "p": "chats"},
		}))
	}))
	defer server.Close()

	tr := newTestTranslator(server.URL)
	resp, err := tr.Translate(context.Background(), keystone.TranslateRequest{
		SourceLanguage:  "en",
		Singular:        "cat",
		Plural:          "cats",
		TargetLanguages: []string{"fr"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fr := resp.Translations["fr"]
	if fr == nil {
		t.Fatal("expected fr translation")
	}
	if fr.Singular != "chat" {
		t.Errorf("expected fr.Singular='chat', got %q", fr.Singular)
	}
	if fr.Plural != "chats" {
		t.Errorf("expected fr.Plural='chats', got %q", fr.Plural)
	}
}

func TestTranslate_MultipleTargetLanguages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toolUseResponse(map[string]any{
			"fr": map[string]any{"s": "Bonjour"},
			"de": map[string]any{"s": "Hallo"},
			"es": map[string]any{"s": "Hola"},
		}))
	}))
	defer server.Close()

	tr := newTestTranslator(server.URL)
	resp, err := tr.Translate(context.Background(), keystone.TranslateRequest{
		SourceLanguage:  "en",
		Singular:        "Hello",
		TargetLanguages: []string{"fr", "de", "es"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Translations) != 3 {
		t.Fatalf("expected 3 translations, got %d", len(resp.Translations))
	}
	expected := map[string]string{"fr": "Bonjour", "de": "Hallo", "es": "Hola"}
	for lang, want := range expected {
		got := resp.Translations[lang]
		if got == nil || got.Singular != want {
			t.Errorf("expected %s.Singular=%q, got %v", lang, want, got)
		}
	}
}

func TestTranslate_EmptyTargetLanguages(t *testing.T) {
	apiCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tr := newTestTranslator(server.URL)
	resp, err := tr.Translate(context.Background(), keystone.TranslateRequest{
		SourceLanguage:  "en",
		Singular:        "Hello",
		TargetLanguages: []string{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiCalled {
		t.Error("API should not be called for empty target languages")
	}
	if len(resp.Translations) != 0 {
		t.Errorf("expected 0 translations, got %d", len(resp.Translations))
	}
}

func TestTranslate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"type": "error",
			"error": map[string]any{
				"type":    "api_error",
				"message": "internal server error",
			},
		})
	}))
	defer server.Close()

	tr := newTestTranslator(server.URL)
	_, err := tr.Translate(context.Background(), keystone.TranslateRequest{
		SourceLanguage:  "en",
		Singular:        "Hello",
		TargetLanguages: []string{"fr"},
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
	if !contains(err.Error(), "anthropic translate") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

func TestTranslate_NoToolUseBlock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(textResponse())
	}))
	defer server.Close()

	tr := newTestTranslator(server.URL)
	resp, err := tr.Translate(context.Background(), keystone.TranslateRequest{
		SourceLanguage:  "en",
		Singular:        "Hello",
		TargetLanguages: []string{"fr"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Translations) != 0 {
		t.Errorf("expected 0 translations when no tool_use block, got %d", len(resp.Translations))
	}
}

func TestTranslate_PartialResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toolUseResponse(map[string]any{
			"fr": map[string]any{"s": "Bonjour"},
		}))
	}))
	defer server.Close()

	tr := newTestTranslator(server.URL)
	resp, err := tr.Translate(context.Background(), keystone.TranslateRequest{
		SourceLanguage:  "en",
		Singular:        "Hello",
		TargetLanguages: []string{"fr", "de", "es"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Translations) != 1 {
		t.Fatalf("expected 1 translation (partial), got %d", len(resp.Translations))
	}
	if resp.Translations["fr"] == nil || resp.Translations["fr"].Singular != "Bonjour" {
		t.Errorf("expected fr='Bonjour', got %v", resp.Translations["fr"])
	}
}

func TestTranslate_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toolUseResponse(map[string]any{
			"fr": map[string]any{"s": "Bonjour"},
		}))
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tr := newTestTranslator(server.URL)
	_, err := tr.Translate(ctx, keystone.TranslateRequest{
		SourceLanguage:  "en",
		Singular:        "Hello",
		TargetLanguages: []string{"fr"},
	})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestWithModel(t *testing.T) {
	tr := newTranslator([]Option{WithModel(sdk.ModelClaudeHaiku4_5_20251001)}, option.WithAPIKey("test-key"))
	if tr.model != sdk.ModelClaudeHaiku4_5_20251001 {
		t.Errorf("expected model %q, got %q", sdk.ModelClaudeHaiku4_5_20251001, tr.model)
	}
}

func TestDefaultModel(t *testing.T) {
	tr := newTranslator(nil, option.WithAPIKey("test-key"))
	if tr.model != sdk.ModelClaudeHaiku4_5 {
		t.Errorf("expected default model %q, got %q", sdk.ModelClaudeHaiku4_5, tr.model)
	}
}

func TestTranslate_RequestStructure(t *testing.T) {
	var captured map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &captured)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toolUseResponse(map[string]any{
			"fr": map[string]any{"s": "Bonjour"},
		}))
	}))
	defer server.Close()

	tr := newTestTranslator(server.URL)
	_, err := tr.Translate(context.Background(), keystone.TranslateRequest{
		SourceLanguage:  "en",
		Singular:        "Hello",
		TargetLanguages: []string{"fr"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify model
	if captured["model"] != "claude-haiku-4-5" {
		t.Errorf("expected model 'claude-haiku-4-5', got %v", captured["model"])
	}

	// Verify tool_choice is "any"
	toolChoice, ok := captured["tool_choice"].(map[string]any)
	if !ok {
		t.Fatalf("expected tool_choice to be a map, got %T", captured["tool_choice"])
	}
	if toolChoice["type"] != "any" {
		t.Errorf("expected tool_choice.type='any', got %v", toolChoice["type"])
	}

	// Verify system prompt exists
	system, ok := captured["system"].([]any)
	if !ok || len(system) == 0 {
		t.Fatal("expected system prompt")
	}

	// Verify tools
	tools, ok := captured["tools"].([]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %v", captured["tools"])
	}
	tool := tools[0].(map[string]any)
	if tool["name"] != "translate" {
		t.Errorf("expected tool name 'translate', got %v", tool["name"])
	}
}

func TestTranslate_Integration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	tr := NewTranslator(apiKey)
	resp, err := tr.Translate(context.Background(), keystone.TranslateRequest{
		SourceLanguage:  "en",
		Singular:        "Hello",
		Plural:          "Hellos",
		TargetLanguages: []string{"fr", "de"},
	})
	if err != nil {
		t.Fatalf("integration test error: %v", err)
	}

	for _, lang := range []string{"fr", "de"} {
		translation := resp.Translations[lang]
		if translation == nil {
			t.Errorf("missing translation for %s", lang)
			continue
		}
		if translation.Singular == "" {
			t.Errorf("empty singular for %s", lang)
		}
		t.Logf("%s: singular=%q plural=%q", lang, translation.Singular, translation.Plural)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
