#!/bin/bash
# Railway CLI Integration Tools for CMS-AI
# Advanced testing and debugging utilities

set -e

PROJECT_NAME="cms-ai"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_debug() { echo -e "${CYAN}[DEBUG]${NC} $1"; }

# Railway project context
ensure_railway_context() {
    if ! railway status &>/dev/null; then
        log_warning "Not in Railway project context. Attempting to link..."

        # Try to link to the project
        if railway link --project "$PROJECT_NAME" &>/dev/null; then
            log_success "Linked to Railway project: $PROJECT_NAME"
        else
            log_error "Failed to link to Railway project. Please run 'railway login' and 'railway link'"
            exit 1
        fi
    fi

    local current_project
    current_project=$(railway status | grep "Project:" | awk '{print $2}' || echo "unknown")
    log_info "Connected to Railway project: $current_project"
}

# Get Railway service URL
get_service_url() {
    local service_name="${1:-web}"

    log_info "Getting URL for service: $service_name"

    # Get the URL from Railway CLI
    local url
    url=$(railway domain --service "$service_name" 2>/dev/null | head -1)

    if [ -z "$url" ]; then
        # Fallback to known production URL
        url="https://cms-ai-production.up.railway.app"
        log_warning "Could not get URL from Railway CLI, using fallback: $url"
    else
        log_success "Service URL: $url"
    fi

    echo "$url"
}

# Monitor Railway logs in real-time during tests
monitor_logs() {
    local service_name="${1:-web}"
    local filter="${2:-.*}"

    log_info "Monitoring logs for service: $service_name (filter: $filter)"

    # Run logs in background and filter
    railway logs --service "$service_name" --follow | grep --line-buffered "$filter" &
    local logs_pid=$!

    # Return the PID so we can kill it later
    echo "$logs_pid"
}

# Quick deployment check
check_deployment_status() {
    log_info "Checking deployment status..."

    # Get deployment information
    railway status --json > /tmp/railway_status.json

    local status
    status=$(jq -r '.deployments[0].status' /tmp/railway_status.json 2>/dev/null || echo "unknown")

    local commit
    commit=$(jq -r '.deployments[0].meta.commitHash' /tmp/railway_status.json 2>/dev/null || echo "unknown")

    log_info "Latest deployment: $status (commit: ${commit:0:8})"

    if [ "$status" = "SUCCESS" ]; then
        log_success "Deployment is healthy"
        return 0
    else
        log_warning "Deployment status: $status"
        return 1
    fi
}

# Test Railway environment variables
test_environment() {
    log_info "Testing environment configuration..."

    # Check if required environment variables are set
    local required_vars=("JWT_SECRET" "NEXTAUTH_SECRET" "HUGGINGFACE_API_KEY")
    local missing_vars=()

    for var in "${required_vars[@]}"; do
        if ! railway variables --service web | grep -q "^$var="; then
            missing_vars+=("$var")
        fi
    done

    if [ ${#missing_vars[@]} -eq 0 ]; then
        log_success "All required environment variables are set"
    else
        log_warning "Missing environment variables: ${missing_vars[*]}"
        return 1
    fi
}

# Deploy and test workflow
deploy_and_test() {
    local commit_message="$1"

    log_info "Starting deploy and test workflow..."

    # 1. Check git status
    if ! git diff-index --quiet HEAD --; then
        log_warning "Uncommitted changes detected. Commit first or use --force"
        return 1
    fi

    # 2. Deploy latest changes
    log_info "Deploying to Railway..."
    railway up --detach

    # 3. Wait for deployment
    local max_wait=300 # 5 minutes
    local wait_time=0
    local deployment_status=""

    log_info "Waiting for deployment to complete..."
    while [ $wait_time -lt $max_wait ]; do
        sleep 10
        wait_time=$((wait_time + 10))

        deployment_status=$(railway status --json | jq -r '.deployments[0].status' 2>/dev/null || echo "UNKNOWN")

        case "$deployment_status" in
            "SUCCESS")
                log_success "Deployment completed successfully after ${wait_time}s"
                break
                ;;
            "FAILED"|"CRASHED")
                log_error "Deployment failed: $deployment_status"
                return 1
                ;;
            *)
                log_info "Deployment status: $deployment_status (${wait_time}s elapsed)"
                ;;
        esac
    done

    if [ "$deployment_status" != "SUCCESS" ]; then
        log_error "Deployment timed out or failed"
        return 1
    fi

    # 4. Run export test
    log_info "Running export test on deployed service..."
    "$SCRIPT_DIR/railway-export-test.sh" single
}

# Debug export issues
debug_export_workflow() {
    local job_id="$1"

    log_info "Debugging export workflow..."

    if [ -n "$job_id" ]; then
        log_info "Debugging specific job: $job_id"

        # Start log monitoring
        local logs_pid
        logs_pid=$(monitor_logs "web" "job.*$job_id")

        # Give it a few seconds to capture logs
        sleep 5

        # Kill log monitoring
        kill "$logs_pid" 2>/dev/null || true
    else
        log_info "Debugging recent export activity..."

        # Show recent logs related to exports
        railway logs --service web -n 100 | grep -i -E "(export|job|render|pptx)" || log_warning "No recent export activity found"
    fi
}

# Performance monitoring
monitor_performance() {
    local duration="${1:-60}" # seconds

    log_info "Monitoring performance for ${duration}s..."

    # Start log monitoring for performance metrics
    local logs_pid
    logs_pid=$(monitor_logs "web" "duration_ms\|slow_request")

    # Run test during monitoring
    "$SCRIPT_DIR/railway-export-test.sh" benchmark 3 &
    local test_pid=$!

    # Wait for test completion or timeout
    wait "$test_pid" || log_warning "Performance test may have failed"

    # Stop log monitoring
    kill "$logs_pid" 2>/dev/null || true

    log_success "Performance monitoring completed"
}

# Railway database operations
manage_database() {
    local action="$1"

    case "$action" in
        "connect")
            log_info "Connecting to Railway database..."
            railway connect postgresql
            ;;
        "status")
            log_info "Checking database status..."
            railway variables --service web | grep -i database || log_warning "No database variables found"
            ;;
        "backup")
            log_info "Creating database backup..."
            # This would require setting up backup scripts
            log_warning "Database backup not implemented yet"
            ;;
        *)
            log_error "Unknown database action: $action"
            log_info "Available actions: connect, status, backup"
            ;;
    esac
}

# Show usage
show_usage() {
    echo "Railway CLI Tools for CMS-AI"
    echo "=============================="
    echo
    echo "Usage: $0 <command> [args...]"
    echo
    echo "Commands:"
    echo "  status               - Check deployment and environment status"
    echo "  test                 - Run quick export test"
    echo "  test-benchmark N     - Run performance benchmark (N iterations)"
    echo "  deploy-test          - Deploy latest changes and test"
    echo "  monitor [duration]   - Monitor performance (default: 60s)"
    echo "  debug [job_id]       - Debug export workflow issues"
    echo "  logs [service]       - Show recent logs (default: web)"
    echo "  url [service]        - Get service URL (default: web)"
    echo "  db <action>          - Database operations (connect|status|backup)"
    echo "  env                  - Check environment variables"
    echo
    echo "Examples:"
    echo "  $0 test                    # Quick export test"
    echo "  $0 test-benchmark 5        # Run 5 iterations benchmark"
    echo "  $0 debug abc123            # Debug specific job"
    echo "  $0 monitor 30              # Monitor for 30 seconds"
}

# Main command dispatcher
main() {
    local command="${1:-status}"

    echo "🚀 Railway CLI Tools for CMS-AI"
    echo "==============================="
    echo

    # Ensure Railway CLI is available and connected
    if ! command -v railway &> /dev/null; then
        log_error "Railway CLI not installed. Install with: npm install -g @railway/cli"
        exit 1
    fi

    ensure_railway_context

    case "$command" in
        "status")
            check_deployment_status
            test_environment
            ;;
        "test")
            "$SCRIPT_DIR/railway-export-test.sh" single
            ;;
        "test-benchmark")
            local iterations="${2:-3}"
            "$SCRIPT_DIR/railway-export-test.sh" benchmark "$iterations"
            ;;
        "deploy-test")
            deploy_and_test
            ;;
        "monitor")
            local duration="${2:-60}"
            monitor_performance "$duration"
            ;;
        "debug")
            local job_id="$2"
            debug_export_workflow "$job_id"
            ;;
        "logs")
            local service="${2:-web}"
            railway logs --service "$service" -n 50
            ;;
        "url")
            local service="${2:-web}"
            get_service_url "$service"
            ;;
        "db")
            local action="$2"
            manage_database "$action"
            ;;
        "env")
            test_environment
            ;;
        "help"|"--help"|"-h")
            show_usage
            ;;
        *)
            log_error "Unknown command: $command"
            echo
            show_usage
            exit 1
            ;;
    esac
}

main "$@"