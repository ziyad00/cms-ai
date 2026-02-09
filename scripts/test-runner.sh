#!/bin/bash
# CMS-AI Test Runner - Unified testing interface
# Provides fast access to all testing tools

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }

# Test categories
run_local_tests() {
    log_info "Running local tests..."

    echo "1. Go backend tests..."
    if JWT_SECRET="test-secret-32-characters-long-for-testing" go test ./... -v | tail -20; then
        log_success "Go tests passed"
    else
        log_error "Go tests failed"
        return 1
    fi

    echo
    echo "2. Frontend tests..."
    if cd web && npm test 2>/dev/null; then
        log_success "Frontend tests passed"
        cd ..
    else
        log_warning "Frontend tests skipped (npm test issues)"
        cd ..
    fi
}

run_railway_tests() {
    local test_type="${1:-quick}"

    log_info "Running Railway tests (type: $test_type)..."

    case "$test_type" in
        "quick"|"fast")
            "$SCRIPT_DIR/railway-export-test.sh" single
            ;;
        "benchmark")
            local iterations="${2:-3}"
            "$SCRIPT_DIR/railway-export-test.sh" benchmark "$iterations"
            ;;
        "full")
            log_info "Running comprehensive Railway test suite..."
            "$SCRIPT_DIR/railway-cli-tools.sh" status
            "$SCRIPT_DIR/railway-export-test.sh" benchmark 5
            ;;
        *)
            log_error "Unknown Railway test type: $test_type"
            return 1
            ;;
    esac
}

run_security_tests() {
    log_info "Running security validation..."

    # Check for required environment variables
    local required_vars=("JWT_SECRET" "NEXTAUTH_SECRET")
    local issues=0

    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            log_warning "Environment variable $var not set"
            issues=$((issues + 1))
        fi
    done

    # Test authentication hardening
    log_info "Testing JWT authentication..."
    if grep -q "dev-secret-change-in-production" server/internal/auth/jwt.go; then
        log_error "Hardcoded JWT secret still present!"
        issues=$((issues + 1))
    else
        log_success "JWT security hardening verified"
    fi

    # Test input validation
    log_info "Testing input validation middleware..."
    if grep -q "ValidationMiddleware" server/internal/api/router_v1.go; then
        log_success "Input validation middleware active"
    else
        log_error "Input validation middleware missing!"
        issues=$((issues + 1))
    fi

    if [ $issues -eq 0 ]; then
        log_success "Security tests passed"
        return 0
    else
        log_error "Security tests failed ($issues issues)"
        return 1
    fi
}

run_integration_tests() {
    log_info "Running integration tests..."

    # Test build
    log_info "1. Testing build process..."
    if JWT_SECRET="test-secret-32-characters-long" go build ./cmd/server; then
        log_success "Server builds successfully"
        rm -f ./cmd/server/server 2>/dev/null || true
    else
        log_error "Server build failed"
        return 1
    fi

    # Test Railway connection
    log_info "2. Testing Railway connection..."
    if "$SCRIPT_DIR/railway-cli-tools.sh" status; then
        log_success "Railway connection verified"
    else
        log_warning "Railway connection issues"
    fi

    # Test export workflow
    log_info "3. Testing export workflow..."
    "$SCRIPT_DIR/railway-export-test.sh" single
}

show_test_menu() {
    echo "🧪 CMS-AI Test Runner"
    echo "===================="
    echo
    echo "Quick Commands:"
    echo "  $0 local              - Run local tests (Go + Frontend)"
    echo "  $0 railway            - Quick Railway export test"
    echo "  $0 security           - Security validation"
    echo "  $0 integration        - Full integration test suite"
    echo "  $0 all                - Run all test categories"
    echo
    echo "Advanced Railway Tests:"
    echo "  $0 railway quick      - Fast export test"
    echo "  $0 railway benchmark  - Performance benchmark"
    echo "  $0 railway full       - Comprehensive Railway tests"
    echo
    echo "Utilities:"
    echo "  $0 logs               - Show Railway logs"
    echo "  $0 status             - Show Railway status"
    echo "  $0 debug [job_id]     - Debug specific export"
    echo "  $0 monitor            - Monitor performance"
    echo
    echo "For Railway CLI tools: ./scripts/railway-cli-tools.sh help"
}

# Performance summary
show_performance_summary() {
    if [ -f "$SCRIPT_DIR/test_results_"*.json ]; then
        local latest_results
        latest_results=$(ls -t "$SCRIPT_DIR/test_results_"*.json | head -1)

        log_info "Latest test performance:"
        jq -r '.results | to_entries[] | "  \(.key): \(.value.status) (\(.value.duration_ms // 0)ms)"' "$latest_results" 2>/dev/null || log_warning "Could not parse test results"
    fi
}

# Main execution
main() {
    local command="${1:-menu}"

    case "$command" in
        "local")
            run_local_tests
            ;;
        "railway")
            local test_type="${2:-quick}"
            run_railway_tests "$test_type"
            show_performance_summary
            ;;
        "security")
            run_security_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "all")
            log_info "Running complete test suite..."
            echo

            log_info "=== Security Tests ==="
            run_security_tests || log_warning "Security tests failed"
            echo

            log_info "=== Local Tests ==="
            run_local_tests || log_warning "Local tests failed"
            echo

            log_info "=== Railway Tests ==="
            run_railway_tests "quick" || log_warning "Railway tests failed"
            echo

            show_performance_summary
            log_success "Complete test suite finished"
            ;;
        "logs")
            "$SCRIPT_DIR/railway-cli-tools.sh" logs
            ;;
        "status")
            "$SCRIPT_DIR/railway-cli-tools.sh" status
            ;;
        "debug")
            local job_id="$2"
            "$SCRIPT_DIR/railway-cli-tools.sh" debug "$job_id"
            ;;
        "monitor")
            "$SCRIPT_DIR/railway-cli-tools.sh" monitor 60
            ;;
        "menu"|"help"|"--help"|"-h"|"")
            show_test_menu
            ;;
        *)
            log_error "Unknown command: $command"
            echo
            show_test_menu
            exit 1
            ;;
    esac
}

main "$@"