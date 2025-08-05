#!/bin/bash

BASE_URL="http://localhost:8080"
CONTENT_TYPE="Content-Type: application/json"

echo "=== Testing Messaging Service Endpoints ==="
echo "Base URL: $BASE_URL"
echo

# Function to run and validate a test
run_test() {
  local description="$1"
  local method="$2"
  local url="$3"
  local data="$4"
  local expected_status="${5:-200}"

  echo "$description..."

  if [[ "$method" == "GET" || "$method" == "DELETE" ]]; then
    status=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url" -H "$CONTENT_TYPE")
  else
    status=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url" -H "$CONTENT_TYPE" -d "$data")
  fi

  if [ "$status" -eq "$expected_status" ]; then
    echo "✅ Passed ($status)"
  else
    echo "❌ Failed ($status expected $expected_status)"
    exit 1
  fi
}

# SECTION 1: Basic Message Sending Tests
echo "=== SECTION 1: Basic Message Sending Tests ==="
run_test "1. Testing SMS send" POST "$BASE_URL/api/messages/message" '{
  "from": "+12016661234",
  "to": "+18045551234",
  "type": "sms",
  "body": "Hello! This is a test SMS message.",
  "attachments": null,
  "timestamp": "2024-11-01T14:00:00Z"
}'

run_test "2. Testing MMS send" POST "$BASE_URL/api/messages/message" '{
  "from": "+12016661234",
  "to": "+18045551234",
  "type": "mms",
  "body": "Hello! This is a test MMS message with attachment.",
  "attachments": ["https://example.com/image.jpg"],
  "timestamp": "2024-11-01T14:00:00Z"
}'

run_test "3. Testing Email send" POST "$BASE_URL/api/messages/email" '{
  "from": "user@usehatchapp.com",
  "to": "contact@gmail.com",
  "body": "Hello! This is a test email message with <b>HTML</b> formatting.",
  "attachments": ["https://example.com/document.pdf"],
  "timestamp": "2024-11-01T14:00:00Z"
}'

# SECTION 2: Incoming Webhook Tests
echo "=== SECTION 2: Incoming Webhook Tests ==="
run_test "4. Testing incoming SMS webhook" POST "$BASE_URL/api/webhooks/message" '{
  "from": "+18045551234",
  "to": "+12016661234",
  "type": "sms",
  "messaging_provider_id": "message-1",
  "body": "This is an incoming SMS message",
  "attachments": null,
  "timestamp": "2024-11-01T14:00:00Z"
}'

run_test "5. Testing incoming MMS webhook" POST "$BASE_URL/api/webhooks/message" '{
  "from": "+18045551234",
  "to": "+12016661234",
  "type": "mms",
  "messaging_provider_id": "message-2",
  "body": "This is an incoming MMS message",
  "attachments": ["https://example.com/received-image.jpg"],
  "timestamp": "2024-11-01T14:00:00Z"
}'

run_test "6. Testing incoming Email webhook" POST "$BASE_URL/api/webhooks/email" '{
  "from": "contact@gmail.com",
  "to": "user@usehatchapp.com",
  "xillio_id": "message-3",
  "body": "<html><body>This is an incoming email with <b>HTML</b> content</body></html>",
  "attachments": ["https://example.com/received-document.pdf"],
  "timestamp": "2024-11-01T14:00:00Z"
}'

# SECTION 3: Outbound Webhook Status Tests
echo "=== SECTION 3: Outbound Webhook Status Tests ==="
run_test "7. Testing SMS send (creates pending message)" POST "$BASE_URL/api/messages/message" '{
  "from": "+12016661234",
  "to": "+18045551234",
  "type": "sms",
  "body": "Hello! This is a test SMS message for webhook testing.",
  "timestamp": "2024-11-01T14:00:00Z"
}'

run_test "8. Testing Email send (creates pending message)" POST "$BASE_URL/api/messages/email" '{
  "from": "contact@gmail.com",
  "to": "user@usehatchapp.com",
  "body": "Hello! This is a test email message for webhook testing.",
  "timestamp": "2024-11-01T14:03:00Z"
}'

# SECTION 4: Data Retrieval Tests
echo "=== SECTION 4: Data Retrieval Tests ==="
run_test "9. Testing get conversations" GET "$BASE_URL/api/conversations?business_phone=+12016661234" ""

run_test "10. Testing get conversations with messages to see statuses" GET "$BASE_URL/api/conversations?business_phone=+12016661234&include_messages=true" ""

echo "=== ✅ All tests passed successfully ==="
