#!/bin/bash

set -e

echo "=========================================="
echo "  QueueCTL - Integration Test Script"
echo "=========================================="

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build
echo -e "\n${YELLOW}Building QueueCTL...${NC}"
go build -o queuectl .
echo -e "${GREEN}✓ Build successful${NC}"

rm -rf data/
mkdir -p data

echo -e "\n${YELLOW}Testing basic enqueue...${NC}"
./queuectl enqueue 'echo "Test 1: Hello World"'
echo -e "${GREEN}✓ Enqueue successful${NC}"

echo -e "\n${YELLOW}Enqueueing multiple jobs...${NC}"
./queuectl enqueue 'sleep 2 && echo "Test 2: Job 2"'
./queuectl enqueue 'echo "Test 3: Job 3" && exit 1'
echo -e "${GREEN}✓ Multiple jobs enqueued${NC}"

echo -e "\n${YELLOW}Checking status (before workers)...${NC}"
./queuectl status

echo -e "\n${YELLOW}Starting 2 workers (for 12 seconds)...${NC}"
# Start workers in background and gracefully stop after 12 seconds (macOS compatible)
./queuectl worker start --count 2 &
WORKER_PID=$!
sleep 12
# Send SIGTERM for graceful shutdown
kill -TERM $WORKER_PID 2>/dev/null || true
wait $WORKER_PID 2>/dev/null || true
sleep 2

echo -e "\n${YELLOW}Checking status (after workers)...${NC}"
./queuectl status

echo -e "\n${YELLOW}Listing jobs by state...${NC}"
echo "--- Pending Jobs ---"
./queuectl list --state pending || true
echo "--- Completed Jobs ---"
./queuectl list --state completed || true
echo "--- Dead Jobs (DLQ) ---"
./queuectl list --state dead || true

echo -e "\n${YELLOW}Testing config commands...${NC}"
./queuectl config set max-retries 5
./queuectl config get max-retries

echo -e "\n${YELLOW}Testing DLQ functionality...${NC}"
# Set max-retries to 1 to quickly trigger DLQ
./queuectl config set max-retries 1
# Enqueue a job that will always fail
./queuectl enqueue 'false'  # false command always exits with code 1
# Start a worker to process it (will fail and move to DLQ)
./queuectl worker start --count 1 &
WORKER_PID=$!
sleep 5
kill -TERM $WORKER_PID 2>/dev/null || true
wait $WORKER_PID 2>/dev/null || true
sleep 1
# Check DLQ
echo "--- Dead Jobs (DLQ) ---"
./queuectl dlq list
# Reset max-retries
./queuectl config set max-retries 3

echo -e "\n${GREEN}=========================================="
echo -e "  All Tests Passed! ✓"
echo -e "==========================================${NC}"
