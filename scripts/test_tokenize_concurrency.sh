#!/bin/bash

# Tokenization Concurrency Test Script
# This script tests the /tokenize and /detokenize endpoints under concurrent load
# to ensure the scheduler handles multiple requests safely.

set -e

# Configuration
OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
MODEL_NAME="${MODEL_NAME:-orca-mini:latest}"
NUM_CONCURRENT="${NUM_CONCURRENT:-5}"
NUM_REQUESTS_PER_WORKER="${NUM_REQUESTS_PER_WORKER:-10}"
KEEP_ALIVE="${KEEP_ALIVE:-30s}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Tokenization Concurrency Test${NC}"
echo "Host: $OLLAMA_HOST"
echo "Model: $MODEL_NAME"
echo "Concurrent workers: $NUM_CONCURRENT"
echo "Requests per worker: $NUM_REQUESTS_PER_WORKER"
echo "Keep-alive: $KEEP_ALIVE"
echo ""

# Check if Ollama is running
if ! curl -s "$OLLAMA_HOST/api/tags" > /dev/null; then
    echo -e "${RED}❌ Ollama is not running at $OLLAMA_HOST${NC}"
    exit 1
fi

# Check if model is available
if ! curl -s "$OLLAMA_HOST/api/tags" | grep -q "\"$MODEL_NAME\""; then
    echo -e "${YELLOW}⚠️  Model $MODEL_NAME not found, pulling...${NC}"
    curl -X POST "$OLLAMA_HOST/api/pull" \
        -H "Content-Type: application/json" \
        -d "{\"name\": \"$MODEL_NAME\"}" > /dev/null
    echo -e "${GREEN}✅ Model pulled successfully${NC}"
fi

# Test data
TEST_TEXTS=(
    "Hello, world!"
    "This is a test of concurrent tokenization."
    "The quick brown fox jumps over the lazy dog."
    "Artificial intelligence is transforming our world."
    "Programming requires logical thinking and problem-solving skills."
    "Climate change is a significant global challenge."
    "Machine learning algorithms are becoming more sophisticated."
    "Data science combines statistics, programming, and domain expertise."
    "Cloud computing enables scalable and flexible infrastructure."
    "Cybersecurity is essential in our digital age."
)

# Function to make a tokenize request
make_tokenize_request() {
    local worker_id=$1
    local request_id=$2
    local text="${TEST_TEXTS[$((request_id % ${#TEST_TEXTS[@]}))]}"
    
    local start_time=$(date +%s%N)
    local response=$(curl -s -w "\n%{http_code}" "$OLLAMA_HOST/api/tokenize" \
        -H "Content-Type: application/json" \
        -d "{
            \"model\": \"$MODEL_NAME\",
            \"content\": \"$text\",
            \"media_type\": \"text\",
            \"keep_alive\": \"$KEEP_ALIVE\"
        }")
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    
    # Extract HTTP status code (last line)
    local http_code=$(echo "$response" | tail -n1)
    local json_response=$(echo "$response" | head -n -1)
    
    echo "$worker_id,$request_id,$http_code,$duration,$json_response"
}

# Function to make a detokenize request
make_detokenize_request() {
    local worker_id=$1
    local request_id=$2
    local tokens=$3
    
    local start_time=$(date +%s%N)
    local response=$(curl -s -w "\n%{http_code}" "$OLLAMA_HOST/api/detokenize" \
        -H "Content-Type: application/json" \
        -d "{
            \"model\": \"$MODEL_NAME\",
            \"tokens\": $tokens,
            \"media_type\": \"text\",
            \"keep_alive\": \"$KEEP_ALIVE\"
        }")
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    
    # Extract HTTP status code (last line)
    local http_code=$(echo "$response" | tail -n1)
    local json_response=$(echo "$response" | head -n -1)
    
    echo "$worker_id,$request_id,$http_code,$duration,$json_response"
}

# Function to run a worker
run_worker() {
    local worker_id=$1
    local results_file=$2
    
    for ((i=0; i<NUM_REQUESTS_PER_WORKER; i++)); do
        local result=$(make_tokenize_request $worker_id $i)
        echo "$result" >> "$results_file"
        
        # Extract tokens from response for detokenize test
        local tokens=$(echo "$result" | cut -d',' -f5 | jq -r '.tokens | @json' 2>/dev/null || echo "[]")
        
        # Make detokenize request
        local detokenize_result=$(make_detokenize_request $worker_id $i "$tokens")
        echo "$detokenize_result" >> "${results_file}.detokenize"
        
        # Small delay to avoid overwhelming the server
        sleep 0.1
    done
}

# Create temporary files for results
TOKENIZE_RESULTS=$(mktemp)
DETOKENIZE_RESULTS=$(mktemp)

echo -e "${BLUE}🚀 Starting concurrent tokenization test...${NC}"

# Start workers in background
for ((worker=0; worker<NUM_CONCURRENT; worker++)); do
    run_worker $worker "$TOKENIZE_RESULTS" &
    WORKER_PIDS[$worker]=$!
done

# Wait for all workers to complete
for pid in "${WORKER_PIDS[@]}"; do
    wait $pid
done

echo -e "${GREEN}✅ All workers completed${NC}"

# Analyze results
echo -e "${BLUE}📊 Analyzing results...${NC}"

# Tokenize results analysis
total_requests=$(wc -l < "$TOKENIZE_RESULTS")
successful_requests=$(grep -c "^[0-9]*,[0-9]*,200," "$TOKENIZE_RESULTS" || echo "0")
failed_requests=$((total_requests - successful_requests))

echo "Tokenize Results:"
echo "  Total requests: $total_requests"
echo "  Successful: $successful_requests"
echo "  Failed: $failed_requests"

if [ $failed_requests -gt 0 ]; then
    echo -e "${RED}  ❌ Some requests failed:${NC}"
    grep -v "^[0-9]*,[0-9]*,200," "$TOKENIZE_RESULTS" | head -5
fi

# Calculate average duration
if [ $successful_requests -gt 0 ]; then
    avg_duration=$(awk -F',' '$3 == "200" {sum += $4; count++} END {if (count > 0) print sum/count; else print 0}' "$TOKENIZE_RESULTS")
    echo "  Average duration: ${avg_duration}ms"
fi

# Detokenize results analysis
total_detokenize=$(wc -l < "$DETOKENIZE_RESULTS")
successful_detokenize=$(grep -c "^[0-9]*,[0-9]*,200," "$DETOKENIZE_RESULTS" || echo "0")
failed_detokenize=$((total_detokenize - successful_detokenize))

echo ""
echo "Detokenize Results:"
echo "  Total requests: $total_detokenize"
echo "  Successful: $successful_detokenize"
echo "  Failed: $failed_detokenize"

if [ $failed_detokenize -gt 0 ]; then
    echo -e "${RED}  ❌ Some detokenize requests failed:${NC}"
    grep -v "^[0-9]*,[0-9]*,200," "$DETOKENIZE_RESULTS" | head -5
fi

# Calculate average detokenize duration
if [ $successful_detokenize -gt 0 ]; then
    avg_detokenize_duration=$(awk -F',' '$3 == "200" {sum += $4; count++} END {if (count > 0) print sum/count; else print 0}' "$DETOKENIZE_RESULTS")
    echo "  Average duration: ${avg_detokenize_duration}ms"
fi

# Performance insights
echo ""
echo -e "${BLUE}💡 Performance Insights:${NC}"

if [ $failed_requests -eq 0 ] && [ $failed_detokenize -eq 0 ]; then
    echo -e "${GREEN}  ✅ All requests succeeded - scheduler is handling concurrent load well${NC}"
else
    echo -e "${YELLOW}  ⚠️  Some requests failed - may need to reduce concurrency or increase keep_alive${NC}"
fi

if [ $successful_requests -gt 0 ]; then
    if (( $(echo "$avg_duration < 1000" | bc -l) )); then
        echo -e "${GREEN}  ✅ Good performance: average response time ${avg_duration}ms${NC}"
    else
        echo -e "${YELLOW}  ⚠️  Slow performance: average response time ${avg_duration}ms${NC}"
    fi
fi

# Clean up
rm -f "$TOKENIZE_RESULTS" "$DETOKENIZE_RESULTS"

echo ""
echo -e "${GREEN}🎉 Concurrency test completed!${NC}"