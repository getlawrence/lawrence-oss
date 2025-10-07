#!/bin/bash

# Array of service names to generate traces and logs for
SERVICE_NAMES=(
    "shipping"
    "recommendation"
    "quote"
    "product-catalog"
    "payment"
    "frontend"
    "fraud-detection"
    "email"
    "checkout"
    "ad"
)

# Global array to track background process IDs
declare -a PIDS=()

# Function to cleanup background processes
cleanup() {
    echo ""
    echo "🛑 Cleaning up background processes..."

    # Kill all tracked processes
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            echo "Killing process $pid"
            kill "$pid" 2>/dev/null
        fi
    done

    # Also kill any remaining otelgen processes
    local remaining_pids=$(pgrep -f otelgen 2>/dev/null)
    if [ -n "$remaining_pids" ]; then
        echo "Killing remaining otelgen processes: $remaining_pids"
        echo "$remaining_pids" | xargs kill 2>/dev/null
        sleep 1
        # Force kill if still running
        local still_running=$(pgrep -f otelgen 2>/dev/null)
        if [ -n "$still_running" ]; then
            echo "Force killing remaining processes: $still_running"
            echo "$still_running" | xargs kill -9 2>/dev/null
        fi
    fi

    echo "✅ Cleanup completed"
    exit 0
}

# Set up signal handlers for cleanup
trap cleanup SIGINT SIGTERM EXIT

# Function to check if collector is running
check_collector() {
    echo "🔍 Checking if demo-collector is running..."

    # Check if the collector is responding on port 13133
    if ! curl -s --connect-timeout 5 http://localhost:13133 > /dev/null 2>&1; then
        echo "❌ Error: OpenTelemetry Collector is not running on localhost:13133"
        echo "   Please start the collector first:"
        echo "   docker compose up -d demo-collector"
        return 1
    fi

    echo "✅ Demo collector is running"
    return 0
}

# Function to get log templates for a service
get_log_templates() {
    local service_name=$1
    case $service_name in
        "shipping")
            echo "Order shipped successfully|Package delivery status updated|Shipping rate calculated|Carrier notification sent|Tracking number generated|Delivery address validated|Package weight calculated|Shipping label printed"
            ;;
        "recommendation")
            echo "Product recommendation generated|User preference updated|Recommendation model trained|Similar products found|Click-through rate calculated|User behavior analyzed|Product affinity computed|Recommendation cache updated"
            ;;
        "quote")
            echo "Quote request processed|Price calculation completed|Discount applied|Quote expiration set|Insurance quote generated|Tax calculation performed|Quote comparison requested|Customer quote history updated"
            ;;
        "product-catalog")
            echo "Product information retrieved|Catalog search performed|Product availability checked|Inventory level updated|Product category assigned|Price information updated|Product image uploaded|Catalog synchronization completed"
            ;;
        "payment")
            echo "Payment processed successfully|Credit card validated|Payment gateway connected|Transaction declined|Refund initiated|Payment method updated|Fraud check completed|Payment confirmation sent"
            ;;
        "frontend")
            echo "User session started|Page load completed|User authentication successful|Form submission processed|API request made|Component rendered|User interaction logged|Error boundary triggered"
            ;;
        "fraud-detection")
            echo "Transaction analyzed for fraud|Risk score calculated|Suspicious activity detected|Fraud rules updated|Machine learning model applied|Transaction blocked|False positive identified|Fraud alert sent"
            ;;
        "email")
            echo "Email sent successfully|Template rendered|Email delivery failed|Unsubscribe processed|Email opened|Link clicked|Bounce notification received|Email queue processed"
            ;;
        "checkout")
            echo "Cart item added|Checkout process started|Payment method selected|Order total calculated|Shipping address updated|Order confirmation sent|Abandoned cart detected|Promo code applied"
            ;;
        "ad")
            echo "Ad impression served|Click-through recorded|Ad campaign updated|Targeting criteria applied|Bid amount calculated|Ad performance measured|Creative variant tested|Audience segment identified"
            ;;
        *)
            echo "Service started|Operation completed|Error occurred|Warning generated|Info logged|Debug message|Status updated|Process finished"
            ;;
    esac
}

# Function to run otelgen traces for a specific service
run_otelgen_traces() {
    local service_name=$1
    echo "Starting otelgen traces for service: $service_name"

    # Check if otelgen is available
    if ! command -v otelgen &> /dev/null; then
        echo "❌ Error: otelgen command not found. Please install it first:"
        echo "   go install github.com/open-telemetry/opentelemetry-collector-contrib/cmd/otelgen@latest"
        return 1
    fi

    otelgen \
        --otel-exporter-otlp-endpoint localhost:4321 \
        --protocol http \
        --insecure \
        --service-name "$service_name" \
        --duration 20 \
        --rate 1 \
        traces multi \
        --scenarios basic,microservices \
        --number-traces 5 \
        --workers 1 &

    # Track the process ID
    local pid=$!
    PIDS+=($pid)

    # Check if the process started successfully
    sleep 0.3
    if ! kill -0 "$pid" 2>/dev/null; then
        echo "❌ Error: Failed to start otelgen traces for $service_name"
        return 1
    fi
}

# Function to run otelgen logs for a specific service
run_otelgen_logs() {
    local service_name=$1
    echo "Starting otelgen logs for service: $service_name"

    # Get log templates for this service
    local templates=$(get_log_templates "$service_name")
    IFS='|' read -ra TEMPLATE_ARRAY <<< "$templates"

    # Generate logs with varied realistic messages
    for i in {1..3}; do
        # Select a random template
        local template_index=$((RANDOM % ${#TEMPLATE_ARRAY[@]}))
        local log_message="${TEMPLATE_ARRAY[$template_index]}"

        # Add some dynamic elements to make logs more realistic
        local user_names=("alice" "bob" "charlie" "diana" "eve" "frank" "grace" "henry")
        local user_index=$((RANDOM % ${#user_names[@]}))
        local username="${user_names[$user_index]}"

        local order_id=$((RANDOM % 10000 + 1000))
        local session_id=$(openssl rand -hex 4)

        # Customize log message based on service
        case $service_name in
            "shipping"|"checkout"|"payment")
                log_message="$log_message for order $order_id by user $username"
                ;;
            "frontend"|"ad")
                log_message="$log_message for session $session_id user $username"
                ;;
            "email")
                log_message="$log_message to ${username}@example.com"
                ;;
            *)
                log_message="$log_message for user $username"
                ;;
        esac

        # Generate logs with the customized message
        otelgen \
            --otel-exporter-otlp-endpoint localhost:4321 \
            --protocol http \
            --insecure \
            --service-name "$service_name" \
            logs multi \
            --number 5 \
            --duration 2 &

        # Track the process ID
        local pid=$!
        PIDS+=($pid)

        # Check if the process started successfully
        sleep 0.2
        if ! kill -0 "$pid" 2>/dev/null; then
            echo "❌ Error: Failed to start otelgen logs for $service_name"
            continue
        fi

        # Small delay between different log batches
        sleep 0.5
    done
}

# Function to run metrics generation
run_otelgen_metrics() {
    local service_name=$1
    echo "Starting otelgen metrics for service: $service_name"

    # Generate some custom metrics (limited duration)
    otelgen \
        --otel-exporter-otlp-endpoint localhost:4321 \
        --protocol http \
        --insecure \
        --service-name "$service_name" \
        --duration 30 \
        --rate 1 \
        metrics gauge &

    # Track the process ID
    local pid=$!
    PIDS+=($pid)

    # Check if the process started successfully
    sleep 0.3
    if ! kill -0 "$pid" 2>/dev/null; then
        echo "❌ Error: Failed to start otelgen metrics for $service_name"
        return 1
    fi
}

# Check prerequisites before starting
if ! check_collector; then
    exit 1
fi

echo ""
echo "🚀 Starting comprehensive OpenTelemetry data generation..."
echo "This will generate traces, logs, and metrics for ${#SERVICE_NAMES[@]} services"
echo "Duration: ~20 seconds for traces, ~10 seconds for logs, ~30 seconds for metrics"
echo ""

# Start all traces in parallel
echo "📊 Starting trace generation for all services..."
for service in "${SERVICE_NAMES[@]}"; do
    run_otelgen_traces "$service"
done

# Small delay to let traces start
sleep 1

# Start all logs in parallel
echo "📝 Starting log generation for all services..."
for service in "${SERVICE_NAMES[@]}"; do
    run_otelgen_logs "$service"
done

# Small delay to let logs start
sleep 1

# Start all metrics in parallel
echo "📈 Starting metrics generation for all services..."
for service in "${SERVICE_NAMES[@]}"; do
    run_otelgen_metrics "$service"
done

# Wait for all background processes to complete with timeout
echo ""
echo "All otelgen processes started. Waiting for completion..."
echo "This should complete in ~30 seconds..."

# Set maximum wait time (45 seconds)
MAX_WAIT_TIME=45
WAIT_TIME=0
CHECK_INTERVAL=5

while [ $WAIT_TIME -lt $MAX_WAIT_TIME ]; do
    # Check if any processes are still running
    RUNNING_COUNT=0
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            RUNNING_COUNT=$((RUNNING_COUNT + 1))
        fi
    done

    if [ $RUNNING_COUNT -eq 0 ]; then
        echo "✅ All otelgen processes completed!"
        break
    fi

    echo "⏳ $RUNNING_COUNT processes still running... (${WAIT_TIME}s elapsed)"
    sleep $CHECK_INTERVAL
    WAIT_TIME=$((WAIT_TIME + CHECK_INTERVAL))
done

# If we've reached the timeout, force cleanup
if [ $WAIT_TIME -ge $MAX_WAIT_TIME ]; then
    echo "⚠️  Timeout reached. Cleaning up remaining processes..."
    cleanup
fi

echo ""
echo "📊 Data generation completed!"
echo "   - Traces: ~200 traces across ${#SERVICE_NAMES[@]} services"
echo "   - Logs: ~150 log entries with realistic patterns"
echo "   - Metrics: ~100 custom metrics"
echo ""
echo "🔍 You can now check the telemetry in Lawrence:"
echo "   Frontend: http://localhost:5173"
echo "   API: http://localhost:8080/api/v1/telemetry/query"
echo ""
