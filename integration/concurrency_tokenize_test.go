package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ollama/ollama/api"
)

func TestConcurrentTokenizeRequests(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	client, _, cleanup := InitServerConnection(ctx, t)
	defer cleanup()

	modelName := "orca-mini:latest"
	if err := PullIfMissing(ctx, client, modelName); err != nil {
		t.Fatalf("pull failed %s", err)
	}

	// Test data for concurrent requests
	testTexts := []string{
		"Hello, world!",
		"This is a test of concurrent tokenization.",
		"The quick brown fox jumps over the lazy dog.",
		"Artificial intelligence is transforming our world.",
		"Programming requires logical thinking and problem-solving skills.",
		"Climate change is a significant global challenge.",
		"Machine learning algorithms are becoming more sophisticated.",
		"Data science combines statistics, programming, and domain expertise.",
		"Cloud computing enables scalable and flexible infrastructure.",
		"Cybersecurity is essential in our digital age.",
	}

	// Configuration for concurrent testing
	numConcurrent := 5
	keepAlive := api.Duration{Duration: 30 * time.Second}

	t.Run("concurrent_tokenize", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan struct {
			index int
			tokens []int
			duration time.Duration
			err     error
		}, len(testTexts))

		// Launch concurrent tokenize requests
		for i := 0; i < numConcurrent; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for j, text := range testTexts {
					req := &api.TokenizeRequest{
						Model:     modelName,
						Content:   text,
						MediaType: "text",
						KeepAlive: &keepAlive,
					}

					start := time.Now()
					resp, err := client.Tokenize(ctx, req)
					duration := time.Since(start)

					results <- struct {
						index int
						tokens []int
						duration time.Duration
						err     error
					}{
						index:    workerID*len(testTexts) + j,
						tokens:   resp.Tokens,
						duration: duration,
						err:     err,
					}
				}
			}(i)
		}

		// Close results channel when all workers complete
		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect and verify results
		var totalTokens int
		var totalDuration time.Duration
		var errorCount int
		var successCount int

		for result := range results {
			if result.err != nil {
				t.Errorf("Worker %d failed: %v", result.index, result.err)
				errorCount++
			} else {
				totalTokens += len(result.tokens)
				totalDuration += result.duration
				successCount++
				
				t.Logf("Worker %d: %d tokens in %v", result.index, len(result.tokens), result.duration)
			}
		}

		t.Logf("Concurrent test results:")
		t.Logf("  Total requests: %d", numConcurrent*len(testTexts))
		t.Logf("  Successful: %d", successCount)
		t.Logf("  Failed: %d", errorCount)
		t.Logf("  Total tokens: %d", totalTokens)
		t.Logf("  Average duration: %v", totalDuration/time.Duration(successCount))
		t.Logf("  Tokens per second: %.0f", float64(totalTokens)/totalDuration.Seconds())

		if errorCount > 0 {
			t.Errorf("Expected no errors, got %d", errorCount)
		}
	})

	t.Run("concurrent_detokenize", func(t *testing.T) {
		// First, tokenize all texts to get tokens for detokenization
		tokensMap := make(map[string][]int)
		for _, text := range testTexts {
			req := &api.TokenizeRequest{
				Model:     modelName,
				Content:   text,
				MediaType: "text",
				KeepAlive: &keepAlive,
			}
			resp, err := client.Tokenize(ctx, req)
			if err != nil {
				t.Fatalf("Failed to tokenize text for detokenize test: %v", err)
			}
			tokensMap[text] = resp.Tokens
		}

		var wg sync.WaitGroup
		results := make(chan struct {
			index int
			content string
			duration time.Duration
			err     error
		}, len(testTexts))

		// Launch concurrent detokenize requests
		for i := 0; i < numConcurrent; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for j, text := range testTexts {
					tokens := tokensMap[text]
					req := &api.DetokenizeRequest{
						Model:     modelName,
						Tokens:    tokens,
						MediaType: "text",
						KeepAlive: &keepAlive,
					}

					start := time.Now()
					resp, err := client.Detokenize(ctx, req)
					duration := time.Since(start)

					results <- struct {
						index int
						content string
						duration time.Duration
						err     error
					}{
						index:    workerID*len(testTexts) + j,
						content:  resp.Content,
						duration: duration,
						err:     err,
					}
				}
			}(i)
		}

		// Close results channel when all workers complete
		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect and verify results
		var totalDuration time.Duration
		var errorCount int
		var successCount int
		var roundTripErrors int

		for result := range results {
			if result.err != nil {
				t.Errorf("Worker %d detokenize failed: %v", result.index, result.err)
				errorCount++
			} else {
				totalDuration += result.duration
				successCount++
				
				// Verify round-trip consistency
				originalText := testTexts[result.index%len(testTexts)]
				if result.content != originalText {
					t.Errorf("Worker %d round-trip failed: expected %q, got %q", 
						result.index, originalText, result.content)
					roundTripErrors++
				}
				
				t.Logf("Worker %d detokenize: %v", result.index, result.duration)
			}
		}

		t.Logf("Concurrent detokenize test results:")
		t.Logf("  Total requests: %d", numConcurrent*len(testTexts))
		t.Logf("  Successful: %d", successCount)
		t.Logf("  Failed: %d", errorCount)
		t.Logf("  Round-trip errors: %d", roundTripErrors)
		t.Logf("  Average duration: %v", totalDuration/time.Duration(successCount))

		if errorCount > 0 {
			t.Errorf("Expected no detokenize errors, got %d", errorCount)
		}
		if roundTripErrors > 0 {
			t.Errorf("Expected no round-trip errors, got %d", roundTripErrors)
		}
	})
}

func TestConcurrentMixedRequests(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	client, _, cleanup := InitServerConnection(ctx, t)
	defer cleanup()

	modelName := "orca-mini:latest"
	if err := PullIfMissing(ctx, client, modelName); err != nil {
		t.Fatalf("pull failed %s", err)
	}

	// Test mixing tokenize and detokenize requests
	testText := "Hello, world! This is a mixed concurrent test."
	keepAlive := api.Duration{Duration: 30 * time.Second}

	// First, get tokens for detokenize requests
	tokenizeReq := &api.TokenizeRequest{
		Model:     modelName,
		Content:   testText,
		MediaType: "text",
		KeepAlive: &keepAlive,
	}
	tokenizeResp, err := client.Tokenize(ctx, tokenizeReq)
	if err != nil {
		t.Fatalf("Failed to tokenize text for mixed test: %v", err)
	}

	t.Run("mixed_tokenize_detokenize", func(t *testing.T) {
		numRequests := 10
		var wg sync.WaitGroup
		results := make(chan struct {
			index    int
			opType   string
			duration time.Duration
			err      error
		}, numRequests)

		// Launch mixed tokenize and detokenize requests
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				var duration time.Duration
				var err error

				if index%2 == 0 {
					// Tokenize request
					req := &api.TokenizeRequest{
						Model:     modelName,
						Content:   testText,
						MediaType: "text",
						KeepAlive: &keepAlive,
					}

					start := time.Now()
					_, err = client.Tokenize(ctx, req)
					duration = time.Since(start)

					results <- struct {
						index    int
						opType   string
						duration time.Duration
						err      error
					}{
						index:    index,
						opType:   "tokenize",
						duration: duration,
						err:     err,
					}
				} else {
					// Detokenize request
					req := &api.DetokenizeRequest{
						Model:     modelName,
						Tokens:    tokenizeResp.Tokens,
						MediaType: "text",
						KeepAlive: &keepAlive,
					}

					start := time.Now()
					_, err = client.Detokenize(ctx, req)
					duration = time.Since(start)

					results <- struct {
						index    int
						opType   string
						duration time.Duration
						err      error
					}{
						index:    index,
						opType:   "detokenize",
						duration: duration,
						err:     err,
					}
				}
			}(i)
		}

		// Close results channel when all workers complete
		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect and verify results
		var totalDuration time.Duration
		var errorCount int
		var successCount int
		var tokenizeCount int
		var detokenizeCount int

		for result := range results {
			if result.err != nil {
				t.Errorf("Request %d (%s) failed: %v", result.index, result.opType, result.err)
				errorCount++
			} else {
				totalDuration += result.duration
				successCount++
				
				if result.opType == "tokenize" {
					tokenizeCount++
				} else {
					detokenizeCount++
				}
				
				t.Logf("Request %d (%s): %v", result.index, result.opType, result.duration)
			}
		}

		t.Logf("Mixed concurrent test results:")
		t.Logf("  Total requests: %d", numRequests)
		t.Logf("  Successful: %d", successCount)
		t.Logf("  Failed: %d", errorCount)
		t.Logf("  Tokenize operations: %d", tokenizeCount)
		t.Logf("  Detokenize operations: %d", detokenizeCount)
		t.Logf("  Average duration: %v", totalDuration/time.Duration(successCount))

		if errorCount > 0 {
			t.Errorf("Expected no errors, got %d", errorCount)
		}
	})
}