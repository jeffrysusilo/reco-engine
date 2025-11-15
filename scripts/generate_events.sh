#!/bin/bash

# Generate sample events script
# This script generates realistic event data for testing

BASE_URL=${1:-"http://localhost:8080"}
NUM_EVENTS=${2:-1000}

echo "Generating $NUM_EVENTS sample events to $BASE_URL"

EVENT_TYPES=("VIEW" "CLICK" "CART" "PURCHASE")
SESSIONS=()

# Generate some session IDs
for i in {1..50}; do
  SESSIONS+=("session_$(uuidgen)")
done

for i in $(seq 1 $NUM_EVENTS); do
  USER_ID=$((RANDOM % 100 + 1))
  ITEM_ID=$((RANDOM % 50 + 1))
  EVENT_TYPE=${EVENT_TYPES[$((RANDOM % 4))]}
  SESSION_ID=${SESSIONS[$((RANDOM % 50))]}
  
  PAYLOAD=$(cat <<EOF
{
  "user_id": $USER_ID,
  "item_id": $ITEM_ID,
  "event_type": "$EVENT_TYPE",
  "session_id": "$SESSION_ID"
}
EOF
)

  curl -s -X POST "$BASE_URL/events" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD" > /dev/null

  if [ $((i % 100)) -eq 0 ]; then
    echo "Generated $i events..."
  fi
  
  # Small delay to avoid overwhelming the system
  sleep 0.01
done

echo "Done! Generated $NUM_EVENTS events"
