#!/bin/bash
# Fast Railway Export Feature Testing Tool
# Uses Railway CLI for optimized testing workflow

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="cms-ai"
RAILWAY_URL="https://cms-ai-production.up.railway.app"
TOKEN_FILE="$SCRIPT_DIR/.railway_token"
TEST_RESULTS_FILE="$SCRIPT_DIR/test_results_$(date +%Y%m%d_%H%M%S).json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }

# Initialize test results
init_test_results() {
    cat > "$TEST_RESULTS_FILE" <<EOF
{
  "test_session": {
    "timestamp": "$(date -Iseconds)",
    "railway_url": "$RAILWAY_URL",
    "test_type": "export_workflow"
  },
  "results": {}
}
EOF
}

# Update test result
update_result() {
    local test_name="$1"
    local status="$2"
    local details="$3"
    local duration="$4"

    jq ".results[\"$test_name\"] = {
        \"status\": \"$status\",
        \"details\": \"$details\",
        \"duration_ms\": $duration,
        \"timestamp\": \"$(date -Iseconds)\"
    }" "$TEST_RESULTS_FILE" > "${TEST_RESULTS_FILE}.tmp" && mv "${TEST_RESULTS_FILE}.tmp" "$TEST_RESULTS_FILE"
}

# Check Railway CLI connection
check_railway_cli() {
    log_info "Checking Railway CLI connection..."

    if ! command -v railway &> /dev/null; then
        log_error "Railway CLI not installed. Install with: npm install -g @railway/cli"
        exit 1
    fi

    # Check if logged in
    if ! railway whoami &> /dev/null; then
        log_warning "Not logged into Railway. Attempting login..."
        railway login
    fi

    log_success "Railway CLI connected: $(railway whoami)"
}

# Get Railway service logs for debugging
get_service_logs() {
    local service_name="$1"
    local lines="${2:-50}"

    log_info "Fetching logs from Railway service: $service_name"
    railway logs --service "$service_name" -n "$lines" || log_warning "Could not fetch logs"
}

# Fast authentication using Railway env vars
fast_auth() {
    local start_time=$(date +%s)

    log_info "Performing fast authentication..."

    # Use Railway CLI to get production URL and test
    local test_email="test-$(date +%s)@example.com"
    local test_password="FastTest123!"

    # Health check first
    local health_response
    if ! health_response=$(curl -sf "$RAILWAY_URL/api/healthz" 2>/dev/null); then
        local end_time=$(date +%s)
        update_result "health_check" "FAIL" "Service not responding" $((end_time - start_time))
        log_error "Health check failed - service may be down"
        return 1
    fi

    # Quick signup/signin
    local signup_response
    signup_response=$(curl -sf -X POST "$RAILWAY_URL/v1/auth/signup" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$test_email\",\"password\":\"$test_password\"}" 2>/dev/null || echo '{"error":"signup_failed"}')

    # Extract token (signup creates and returns token)
    local token
    token=$(echo "$signup_response" | jq -r '.token // empty' 2>/dev/null)

    if [ -z "$token" ] || [ "$token" = "null" ]; then
        # Try signin if signup failed
        local signin_response
        signin_response=$(curl -sf -X POST "$RAILWAY_URL/v1/auth/signin" \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"$test_email\",\"password\":\"$test_password\"}" 2>/dev/null || echo '{"error":"signin_failed"}')

        token=$(echo "$signin_response" | jq -r '.token // empty' 2>/dev/null)
    fi

    if [ -z "$token" ] || [ "$token" = "null" ]; then
        local end_time=$(date +%s)
        update_result "authentication" "FAIL" "Could not obtain JWT token" $((end_time - start_time))
        log_error "Authentication failed"
        return 1
    fi

    # Save token for reuse
    echo "$token" > "$TOKEN_FILE"

    local end_time=$(date +%s)
    update_result "authentication" "PASS" "JWT token obtained" $((end_time - start_time))
    log_success "Authentication successful (${#token} chars token)"

    echo "$token"
}

# Test export workflow with timing
test_export_workflow() {
    local token="$1"
    local start_time=$(date +%s)

    log_info "Testing export workflow..."

    # Step 1: Generate template
    log_info "1/5 Generating template..."
    local template_response
    template_response=$(curl -sf -X POST "$RAILWAY_URL/v1/templates/generate" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d '{
            "name": "Fast Test Template",
            "prompt": "Create a simple 2-slide presentation about software testing with Railway CLI"
        }' 2>/dev/null)

    local template_id
    template_id=$(echo "$template_response" | jq -r '.template.id // .id // empty' 2>/dev/null)
    local template_version_id
    template_version_id=$(echo "$template_response" | jq -r '.version.id // empty' 2>/dev/null)

    if [ -z "$template_id" ] || [ "$template_id" = "null" ]; then
        local end_time=$(date +%s)
        update_result "template_generation" "FAIL" "No template ID returned" $((end_time - start_time))
        log_error "Template generation failed"
        return 1
    fi

    log_success "Template generated: $template_id (version: $template_version_id)"

    # Step 2: Create deck from template

    log_info "2/5 Creating deck from template..."
    local deck_response
    deck_response=$(curl -sf -X POST "$RAILWAY_URL/v1/decks" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{\"name\": \"Fast Test Deck - $(date +%H%M%S)\", \"sourceTemplateVersionId\": \"$template_version_id\", \"content\": \"Fast test deck for Railway CLI export verification\"}" 2>/dev/null)

    local deck_version_id
    deck_version_id=$(echo "$deck_response" | jq -r '.deckVersion.id // .version.id // .id // empty' 2>/dev/null)

    if [ -z "$deck_version_id" ] || [ "$deck_version_id" = "null" ]; then
        local end_time=$(date +%s)
        update_result "deck_creation" "FAIL" "No deck version ID returned" $((end_time - start_time))
        log_error "Deck creation failed"
        return 1
    fi

    log_success "Deck created: $deck_version_id"

    # Step 3: Export deck
    log_info "3/5 Starting export..."
    local export_response
    export_response=$(curl -sf -X POST "$RAILWAY_URL/v1/deck-versions/$deck_version_id/export" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" 2>/dev/null)

    local job_id
    job_id=$(echo "$export_response" | jq -r '.job.id // .id // empty' 2>/dev/null)

    if [ -z "$job_id" ] || [ "$job_id" = "null" ]; then
        local end_time=$(date +%s)
        update_result "export_start" "FAIL" "No job ID returned" $((end_time - start_time))
        log_error "Export start failed"
        return 1
    fi

    log_success "Export job started: $job_id"

    # Step 4: Poll for completion
    log_info "4/5 Waiting for export completion..."
    local max_attempts=30
    local attempt=0
    local job_status=""
    local asset_id=""

    while [ $attempt -lt $max_attempts ]; do
        sleep 2
        attempt=$((attempt + 1))

        local job_response
        job_response=$(curl -sf "$RAILWAY_URL/v1/jobs/$job_id" \
            -H "Authorization: Bearer $token" 2>/dev/null || echo '{}')

        job_status=$(echo "$job_response" | jq -r '.job.status // .status // empty' 2>/dev/null)
        asset_id=$(echo "$job_response" | jq -r '.job.outputRef // .outputRef // .assetId // empty' 2>/dev/null)

        case "$job_status" in
            "Done"|"completed"|"Completed"|"SUCCESS")
                log_success "Export completed after ${attempt} attempts (${asset_id})"
                break
                ;;
            "Failed"|"failed"|"Error"|"FAILED"|"deadletter")
                local end_time=$(date +%s)
                update_result "export_completion" "FAIL" "Job failed: $job_status" $((end_time - start_time))
                log_error "Export job failed: $job_status"
                return 1
                ;;
            "Running"|"running"|"Queued"|"queued"|"RUNNING"|"QUEUED")
                log_info "Job status: $job_status (attempt $attempt/$max_attempts)"
                ;;
            *)
                log_info "Unknown job status: $job_status (attempt $attempt/$max_attempts)"
                ;;
        esac

        if [ -n "$asset_id" ] && [ "$asset_id" != "null" ]; then
            break
        fi
    done

    if [ $attempt -eq $max_attempts ] || [ -z "$asset_id" ]; then
        local end_time=$(date +%s)
        update_result "export_completion" "FAIL" "Export timed out or no asset ID" $((end_time - start_time))
        log_error "Export timed out or failed to produce asset"
        return 1
    fi

    # Step 5: Test download
    log_info "5/5 Testing asset download..."
    local download_response
    download_response=$(curl -sf -I "$RAILWAY_URL/v1/assets/$asset_id" \
        -H "Authorization: Bearer $token" 2>/dev/null)

    if echo "$download_response" | grep -q "200 OK"; then
        local end_time=$(date +%s)
        update_result "export_workflow" "PASS" "Complete workflow successful" $((end_time - start_time))
        log_success "Export workflow completed successfully!"
        log_success "Asset ID: $asset_id"
        log_success "Download URL: $RAILWAY_URL/v1/assets/$asset_id"
        return 0
    else
        local end_time=$(date +%s)
        update_result "asset_download" "FAIL" "Asset not downloadable" $((end_time - start_time))
        log_error "Asset download test failed"
        return 1
    fi
}

# Run performance benchmark
run_performance_benchmark() {
    local iterations="${1:-3}"

    log_info "Running performance benchmark ($iterations iterations)..."

    local total_time=0
    local success_count=0

    for i in $(seq 1 $iterations); do
        log_info "Benchmark iteration $i/$iterations"

        local start_time=$(date +%s)

        if [ -f "$TOKEN_FILE" ]; then
            local token=$(cat "$TOKEN_FILE")
        else
            local token
            token=$(fast_auth) || continue
        fi

        if test_export_workflow "$token"; then
            success_count=$((success_count + 1))
        fi

        local end_time=$(date +%s)
        local iteration_time=$((end_time - start_time))
        total_time=$((total_time + iteration_time))

        log_info "Iteration $i completed in ${iteration_time}ms"

        # Short break between iterations
        sleep 1
    done

    local avg_time=$((total_time / iterations))
    local success_rate=$((success_count * 100 / iterations))

    log_success "Benchmark Results:"
    log_success "  Iterations: $iterations"
    log_success "  Success Rate: $success_rate%"
    log_success "  Average Time: ${avg_time}ms"
    log_success "  Total Time: ${total_time}ms"

    update_result "performance_benchmark" "COMPLETE" "Rate: $success_rate%, Avg: ${avg_time}ms" "$avg_time"
}

# Main execution
main() {
    echo "🚀 Fast Railway Export Testing Tool"
    echo "=================================="

    init_test_results

    # Parse command line arguments
    local mode="${1:-single}"

    case "$mode" in
        "benchmark")
            local iterations="${2:-5}"
            check_railway_cli
            run_performance_benchmark "$iterations"
            ;;
        "logs")
            get_service_logs "${2:-web}"
            ;;
        "single"|*)
            check_railway_cli

            # Get or reuse auth token
            local token
            if [ -f "$TOKEN_FILE" ] && [ -s "$TOKEN_FILE" ]; then
                token=$(cat "$TOKEN_FILE")
                log_info "Reusing existing auth token"
            else
                token=$(fast_auth) || exit 1
            fi

            # Run export test
            test_export_workflow "$token" || exit 1
            ;;
    esac

    log_success "Test results saved to: $TEST_RESULTS_FILE"
    log_info "View results: cat $TEST_RESULTS_FILE | jq"
}

# Cleanup on exit
cleanup() {
    if [ -n "$TEST_RESULTS_FILE" ] && [ -f "$TEST_RESULTS_FILE" ]; then
        log_info "Final results:"
        jq -r '.results | to_entries[] | "  \(.key): \(.value.status) (\(.value.duration_ms)ms)"' "$TEST_RESULTS_FILE"
    fi
}
trap cleanup EXIT

main "$@"