package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ollama/ollama/api"
)

func main() {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Example text to tokenize
	text := "Hello, world! This is a test of the tokenization API."

	// Tokenize the text
	tokenizeReq := &api.TokenizeRequest{
		Model:   "llama3.2:3b", // Use a model that's available
		Content: text,
	}

	fmt.Printf("Tokenizing: %q\n", text)
	tokenizeResp, err := client.Tokenize(ctx, tokenizeReq)
	if err != nil {
		log.Fatal("Tokenize error:", err)
	}

	fmt.Printf("Tokens: %v\n", tokenizeResp.Tokens)
	fmt.Printf("Token count: %d\n", len(tokenizeResp.Tokens))
	fmt.Printf("Total duration: %v\n", tokenizeResp.TotalDuration)
	fmt.Printf("Load duration: %v\n", tokenizeResp.LoadDuration)

	// Detokenize the tokens back to text
	detokenizeReq := &api.DetokenizeRequest{
		Model:  "llama3.2:3b",
		Tokens: tokenizeResp.Tokens,
	}

	fmt.Printf("\nDetokenizing tokens: %v\n", tokenizeResp.Tokens)
	detokenizeResp, err := client.Detokenize(ctx, detokenizeReq)
	if err != nil {
		log.Fatal("Detokenize error:", err)
	}

	fmt.Printf("Detokenized text: %q\n", detokenizeResp.Content)
	fmt.Printf("Total duration: %v\n", detokenizeResp.TotalDuration)
	fmt.Printf("Load duration: %v\n", detokenizeResp.LoadDuration)

	// Verify round-trip
	if text == detokenizeResp.Content {
		fmt.Println("\n✅ Round-trip tokenization successful!")
	} else {
		fmt.Printf("\n❌ Round-trip tokenization failed!\nOriginal: %q\nResult:   %q\n", text, detokenizeResp.Content)
	}
}