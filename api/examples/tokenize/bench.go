package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

func main() {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Configuration
	modelName := "llama3.2:3b"
	keepAlive := api.Duration{Duration: 5 * time.Minute}
	
	// Test data - various text samples
	testTexts := []string{
		"Hello, world!",
		"This is a simple sentence for testing tokenization performance.",
		"The quick brown fox jumps over the lazy dog. This pangram contains every letter of the English alphabet at least once.",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
		"Artificial intelligence is transforming the way we live and work. Machine learning algorithms are becoming increasingly sophisticated, enabling computers to perform tasks that were once thought to be the exclusive domain of human intelligence.",
		"Programming is the art of telling another human being what one wants the computer to do. It requires logical thinking, problem-solving skills, and attention to detail.",
		"Climate change represents one of the most significant challenges facing humanity in the 21st century. Rising global temperatures, melting ice caps, and extreme weather events are just some of the consequences we must address.",
		"",
		"Special characters: !@#$%^&*()_+-=[]{}|;':\",./<>?",
		"Unicode: 🚀🌟🎉 你好世界 Привет мир",
	}

	fmt.Printf("🔧 Tokenization Benchmark Tool\n")
	fmt.Printf("Model: %s\n", modelName)
	fmt.Printf("Keep-alive: %v\n", keepAlive.Duration)
	fmt.Printf("Test samples: %d\n\n", len(testTexts))

	// Warm up the model with a simple tokenization
	fmt.Printf("🔥 Warming up model...\n")
	warmupReq := &api.TokenizeRequest{
		Model:     modelName,
		Content:   "warmup",
		MediaType: "text",
		KeepAlive: &keepAlive,
	}

	_, err = client.Tokenize(ctx, warmupReq)
	if err != nil {
		log.Fatal("Warmup failed:", err)
	}
	fmt.Printf("✅ Model warmed up\n\n")

	// Benchmark tokenization
	fmt.Printf("📊 Running tokenization benchmark...\n")
	var totalTokens int
	var totalTokenizeTime time.Duration
	var totalDetokenizeTime time.Duration
	var totalLoadTime time.Duration

	for i, text := range testTexts {
		// Tokenize
		tokenizeReq := &api.TokenizeRequest{
			Model:     modelName,
			Content:   text,
			MediaType: "text",
			KeepAlive: &keepAlive,
		}

		start := time.Now()
		tokenizeResp, err := client.Tokenize(ctx, tokenizeReq)
		if err != nil {
			log.Printf("❌ Tokenization failed for text %d: %v", i, err)
			continue
		}
		tokenizeTime := time.Since(start)

		// Detokenize
		detokenizeReq := &api.DetokenizeRequest{
			Model:     modelName,
			Tokens:    tokenizeResp.Tokens,
			MediaType: "text",
			KeepAlive: &keepAlive,
		}

		start = time.Now()
		detokenizeResp, err := client.Detokenize(ctx, detokenizeReq)
		if err != nil {
			log.Printf("❌ Detokenization failed for text %d: %v", i, err)
			continue
		}
		detokenizeTime := time.Since(start)

		// Verify round-trip
		if detokenizeResp.Content != text {
			fmt.Printf("⚠️  Round-trip mismatch for text %d:\n", i)
			fmt.Printf("   Original: %q\n", text)
			fmt.Printf("   Result:   %q\n", detokenizeResp.Content)
		}

		// Accumulate metrics
		totalTokens += len(tokenizeResp.Tokens)
		totalTokenizeTime += tokenizeTime
		totalDetokenizeTime += detokenizeTime
		totalLoadTime += tokenizeResp.LoadDuration

		// Print individual results
		fmt.Printf("Text %d (%d chars): %d tokens in %v (load: %v)\n", 
			i+1, len(text), len(tokenizeResp.Tokens), tokenizeTime, tokenizeResp.LoadDuration)
	}

	// Print summary statistics
	fmt.Printf("\n📈 Benchmark Results:\n")
	fmt.Printf("Total texts processed: %d\n", len(testTexts))
	fmt.Printf("Total tokens: %d\n", totalTokens)
	fmt.Printf("Average tokens per text: %.1f\n", float64(totalTokens)/float64(len(testTexts)))
	fmt.Printf("Total tokenize time: %v\n", totalTokenizeTime)
	fmt.Printf("Total detokenize time: %v\n", totalDetokenizeTime)
	fmt.Printf("Total load time: %v\n", totalLoadTime)
	fmt.Printf("Average tokenize time per text: %v\n", totalTokenizeTime/time.Duration(len(testTexts)))
	fmt.Printf("Average detokenize time per text: %v\n", totalDetokenizeTime/time.Duration(len(testTexts)))
	fmt.Printf("Average load time per text: %v\n", totalLoadTime/time.Duration(len(testTexts)))
	
	if totalTokens > 0 {
		fmt.Printf("Tokens per second (tokenize): %.0f\n", float64(totalTokens)/totalTokenizeTime.Seconds())
		fmt.Printf("Tokens per second (detokenize): %.0f\n", float64(totalTokens)/totalDetokenizeTime.Seconds())
	}

	// Performance insights
	fmt.Printf("\n💡 Performance Insights:\n")
	if totalLoadTime > totalTokenizeTime {
		fmt.Printf("• Load time dominates - consider using keep_alive for batch operations\n")
	} else {
		fmt.Printf("• Tokenization is the bottleneck - model is well-optimized\n")
	}
	
	avgLoadTime := totalLoadTime / time.Duration(len(testTexts))
	if avgLoadTime < 100*time.Millisecond {
		fmt.Printf("• Keep-alive is working well (avg load time: %v)\n", avgLoadTime)
	} else {
		fmt.Printf("• Consider longer keep_alive duration (avg load time: %v)\n", avgLoadTime)
	}
}