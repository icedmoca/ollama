#!/bin/bash

MODEL="mistral:latest"
ENDPOINT="http://localhost:11434"

function test_tokenize_detokenize() {
  local text="$1"

  echo "🔹 Input: $text"

  # Build JSON safely using jq
  tokenize_payload=$(jq -n \
    --arg model "$MODEL" \
    --arg content "$text" \
    --arg media_type "text" \
    '{model: $model, content: $content, media_type: $media_type}')

  # Tokenize
  response=$(curl -s -X POST "$ENDPOINT/api/tokenize" \
    -H "Content-Type: application/json" \
    -d "$tokenize_payload")

  echo "🧬 Tokenize response: $response"
  tokens=$(echo "$response" | jq .tokens)

  if [[ "$tokens" == "null" || "$tokens" == "" ]]; then
    echo "❌ Tokenization failed."
    echo
    return
  fi

  # Build detokenize payload
  detokenize_payload=$(jq -n \
    --arg model "$MODEL" \
    --argjson tokens "$tokens" \
    --arg media_type "text" \
    '{model: $model, tokens: $tokens, media_type: $media_type}')

  # Detokenize
  response_detokenize=$(curl -s -X POST "$ENDPOINT/api/detokenize" \
    -H "Content-Type: application/json" \
    -d "$detokenize_payload")

  echo "🔁 Detokenize response: $response_detokenize"
  output=$(echo "$response_detokenize" | jq -r .content)
  echo "✅ Round-trip result: \"$output\""
  echo
}

echo "⚙️ Starting comprehensive tokenizer test for model: $MODEL"
echo "-------------------------------------------------------------"

test_tokenize_detokenize "Hello, world!"
test_tokenize_detokenize "🚀 Let's test emoji, Unicode — and punctuation."
test_tokenize_detokenize ""
test_tokenize_detokenize "     "
test_tokenize_detokenize "This is a longer sentence designed to test the ability of the tokenizer to handle complex phrasing, contractions like don't, and symbols like #, @, %, &, and others."
test_tokenize_detokenize "你好，世界！"
test_tokenize_detokenize "function add(a, b) { return a + b; } // JavaScript code"
test_tokenize_detokenize "“Curly quotes” and — em-dashes — with ellipses…"

echo "✅ All tests completed."
