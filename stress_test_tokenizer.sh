#!/bin/bash

MODEL="mistral:latest"
ENDPOINT="http://localhost:11434"

function stress_test_tokenizer() {
  local text="$1"
  local label="$2"

  echo "🔹 [$label] Input: (truncated) ${text:0:100}..."

  tokenize_payload=$(jq -n \
    --arg model "$MODEL" \
    --arg content "$text" \
    --arg media_type "text" \
    '{model: $model, content: $content, media_type: $media_type}')

  response=$(curl -s -X POST "$ENDPOINT/api/tokenize" \
    -H "Content-Type: application/json" \
    -d "$tokenize_payload")

  tokens=$(echo "$response" | jq .tokens)
  echo "🧬 Tokens: $(echo "$tokens" | jq length) tokens"

  if [[ "$tokens" == "null" || "$tokens" == "" ]]; then
    echo "❌ Tokenization failed."
    echo
    return
  fi

  detokenize_payload=$(jq -n \
    --arg model "$MODEL" \
    --argjson tokens "$tokens" \
    --arg media_type "text" \
    '{model: $model, tokens: $tokens, media_type: $media_type}')

  response_detokenize=$(curl -s -X POST "$ENDPOINT/api/detokenize" \
    -H "Content-Type: application/json" \
    -d "$detokenize_payload")

  output=$(echo "$response_detokenize" | jq -r .content)
  echo "✅ Detokenized (truncated): \"${output:0:100}...\""

  if [[ "${text:0:250}" == "${output:1:250}" ]]; then
    echo "✅ Round-trip MATCH (prefix)"
  else
    echo "⚠️  Round-trip MISMATCH"
  fi

  echo
}

echo "🚀 Running ultra-hard tokenizer stress tests for model: $MODEL"
echo "==============================================================="

# Test 1: Long sentence 3k+ tokens
long_sentence=$(yes "The quick brown fox jumps over the lazy dog." | head -n 300 | tr '\n' ' ')
stress_test_tokenizer "$long_sentence" "300 sentences"

# Test 2: Complex multilingual + emoji + math + symbols + smart quotes
mixed_input="🚀✨🤖 Hello 世界! 这是一个测试。💬👩‍💻 def add(x, y): return x + y # 🧠 🏳️‍🌈\n\n🧮∑x=1^n π≈3.14159\n\n¡Hola! ¿Cómo estás? Voilà! déjà vu. \"Theorem ∀x∃y\""
stress_test_tokenizer "$mixed_input" "Multimodal Fusion"

# Test 3: 1,000 printable random Unicode characters
unicode_input=$(cat /dev/urandom | tr -dc '[:print:]' | head -c 1000)
stress_test_tokenizer "$unicode_input" "Random Printable Blob"

# Test 4: Repeated token load test
repeated=$(printf "test%.0s " {1..1024})
stress_test_tokenizer "$repeated" "Repetitive Token Compression"

# Test 5: Whitespace edge case
sparse=$'\n\n\n    \t\t  \n\n     text      \n\n\n\n   \tend'
stress_test_tokenizer "$sparse" "Sparse Whitespace"

# Test 6: Near context limit (~3900 tokens)
near_limit=$(yes "token" | head -n 3900 | tr '\n' ' ')
stress_test_tokenizer "$near_limit" "Max Context Push"

# Test 7: Fuzzy injected smart characters and invisible junk
fuzz="﻿\u200B\uFEFF\u00A0\u200E\u202A Here’s a “quote”—and\u202Fsome non-breaking junk…"
stress_test_tokenizer "$fuzz" "Invisible Char Fuzz"

# Test 8: Bracket/brace-heavy code and unbalanced nesting
code="function fn(x) { if (x > 0) { while(true) { doSomething(x); } } return x; // Missing closing brace"
stress_test_tokenizer "$code" "Unbalanced Code Nesting"

# Test 9: Structured markdown + LaTeX hybrid
doc=$'## Heading\n\nHere is a list:\n- Item 1\n- Item 2\n\n$$E=mc^2$$\n\n```python\ndef foo():\n    return True\n```'
stress_test_tokenizer "$doc" "Markdown + Math"

# Test 10: Long Arabic + RTL characters
rtl="الذكاء الاصطناعي هو المستقبل. 🧠🚀 هل يمكنك فهم هذا النص؟ هذا اختبار للنصوص من اليمين إلى اليسار."
stress_test_tokenizer "$rtl" "RTL Language"

echo "✅ Ultra-hard stress test complete."
