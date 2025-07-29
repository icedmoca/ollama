package llm

import (
	"testing"
)

func TestMockTokenizerAdapter(t *testing.T) {
	adapter := NewMockTokenizerAdapter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple text",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "complex text",
			input:    "This is a test of the mock tokenizer adapter",
			expected: "This is a test of the mock tokenizer adapter",
		},
		{
			name:     "special characters",
			input:    "Hello, world! How are you?",
			expected: "Hello, world! How are you?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test tokenization
			tokens, err := adapter.Tokenize(tt.input)
			if err != nil {
				t.Fatalf("Tokenize failed: %v", err)
			}

			// Verify we got some tokens (except for empty input)
			if tt.input != "" && len(tokens) == 0 {
				t.Errorf("Expected non-empty tokens for non-empty input")
			}

			// Test detokenization
			content, err := adapter.Detokenize(tokens)
			if err != nil {
				t.Fatalf("Detokenize failed: %v", err)
			}

			// Verify round-trip consistency
			if content != tt.expected {
				t.Errorf("Round-trip failed: expected %q, got %q", tt.expected, content)
			}

			// Test that repeated calls return the same results
			tokens2, err := adapter.Tokenize(tt.input)
			if err != nil {
				t.Fatalf("Second Tokenize failed: %v", err)
			}

			if len(tokens) != len(tokens2) {
				t.Errorf("Token count mismatch: first=%d, second=%d", len(tokens), len(tokens2))
			}

			for i, token := range tokens {
				if i >= len(tokens2) || token != tokens2[i] {
					t.Errorf("Token mismatch at index %d: first=%d, second=%d", i, token, tokens2[i])
				}
			}
		})
	}
}

func TestMockTokenizerAdapterDeterministic(t *testing.T) {
	adapter1 := NewMockTokenizerAdapter()
	adapter2 := NewMockTokenizerAdapter()

	testText := "Hello world test"

	// Test that different adapter instances produce the same results
	tokens1, err := adapter1.Tokenize(testText)
	if err != nil {
		t.Fatalf("Adapter1 Tokenize failed: %v", err)
	}

	tokens2, err := adapter2.Tokenize(testText)
	if err != nil {
		t.Fatalf("Adapter2 Tokenize failed: %v", err)
	}

	if len(tokens1) != len(tokens2) {
		t.Errorf("Token count mismatch between adapters: %d vs %d", len(tokens1), len(tokens2))
	}

	for i, token := range tokens1 {
		if i >= len(tokens2) || token != tokens2[i] {
			t.Errorf("Token mismatch at index %d: adapter1=%d, adapter2=%d", i, token, tokens2[i])
		}
	}
}

func TestMockTokenizerAdapterInterface(t *testing.T) {
	// Test that MockTokenizerAdapter properly implements TokenizerAdapter interface
	var _ TokenizerAdapter = NewMockTokenizerAdapter()
	
	// Test that llamaTokenizerAdapter properly implements TokenizerAdapter interface
	// (This would require a mock LlamaServer, so we'll just verify the interface is implemented)
	var _ TokenizerAdapter = (*llamaTokenizerAdapter)(nil)
}
