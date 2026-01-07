#!/bin/bash

API="http://localhost:8080"

echo "=========================================="
echo "  Research Compute Queue API Test Suite"
echo "=========================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 1. Health check
echo -e "\n${GREEN}[1/7] Health Check${NC}"
curl -s $API/health | jq
sleep 1

# 2. Register user
echo -e "\n${GREEN}[2/7] Register User${NC}"
curl -s -X POST $API/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com","password":"testpass123","group_id":1}' | jq
sleep 1

# 3. Login and get token
echo -e "\n${GREEN}[3/7] Login${NC}"
TOKEN=$(curl -s -X POST $API/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com","password":"testpass123"}' \
  | jq -r '.token')

if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
    echo "✓ Token received: ${TOKEN:0:50}..."
else
    echo -e "${RED}✗ Failed to get token${NC}"
    exit 1
fi
sleep 1

# 4. Submit job
echo -e "\n${GREEN}[4/7] Submit Job${NC}"
JOB_ID=$(curl -s -X POST $API/api/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"script":"python test.py","cpu_cores":4,"memory_gb":16,"priority":3}' \
  | jq -r '.job_id')

if [ "$JOB_ID" != "null" ] && [ -n "$JOB_ID" ]; then
    echo "✓ Job created with ID: $JOB_ID"
else
    echo -e "${RED}✗ Failed to create job${NC}"
    exit 1
fi
sleep 1

# 5. Get job status
echo -e "\n${GREEN}[5/7] Get Job Status${NC}"
curl -s $API/api/jobs/$JOB_ID \
  -H "Authorization: Bearer $TOKEN" | jq
sleep 1

# 6. List all jobs
echo -e "\n${GREEN}[6/7] List All Jobs${NC}"
curl -s "$API/api/jobs" \
  -H "Authorization: Bearer $TOKEN" | jq
sleep 1

# 7. Wait and check if job completed
echo -e "\n${GREEN}[7/7] Wait for Job Completion${NC}"
echo "Waiting 60 seconds for scheduler to pick up and complete job..."
sleep 60

echo -e "\nFinal job status:"
curl -s $API/api/jobs/$JOB_ID \
  -H "Authorization: Bearer $TOKEN" | jq

echo -e "\n=========================================="
echo -e "${GREEN}  Test Suite Complete!${NC}"
echo "=========================================="