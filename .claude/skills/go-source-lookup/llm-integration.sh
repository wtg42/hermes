#!/usr/bin/env bash
# llm-integration.sh
# Main entry point for LLM integration
# Handles automatic query trigger detection and invocation
# Usage: llm-integration.sh --lsp <message> | --compile-error <output> | --check-deprecated <code>

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

show_usage() {
    cat <<EOF
Usage: llm-integration.sh [options]

Options:
  --lsp <message>              Process LSP warning
  --compile-error <output>     Process compilation error
  --check-deprecated <code>    Check for deprecated APIs
  --query <package> [symbol]   Direct query (without trigger detection)
  --help                       Show this help message

Examples:
  llm-integration.sh --lsp "undefined: fmt.Println"
  llm-integration.sh --compile-error "cannot use ... (type X) as type Y"
  llm-integration.sh --query fmt Println
EOF
}

handle_direct_query() {
    local package=$1
    local symbol=${2:-}

    echo "=== Go Source Lookup: ${package}${symbol:+.}${symbol} ===" >&2
    echo ""

    # Format and output result for LLM
    "${SCRIPT_DIR}/format-for-llm.sh" "$package" "$symbol"
}

handle_lsp_error() {
    local message=$1

    echo "=== Processing LSP Warning ===" >&2
    echo "Message: $message" >&2
    echo ""

    # Trigger detection and query
    "${SCRIPT_DIR}/trigger-logic.sh" lsp-warning "$message" || true

    # Format output
    local package symbol
    if [[ "$message" =~ undefined:\ ([a-zA-Z_][a-zA-Z0-9_.]*) ]]; then
        package="${BASH_REMATCH[1]}"
        symbol=""
    elif [[ "$message" =~ ([a-zA-Z_][a-zA-Z0-9_]*)\. ]]; then
        package="${BASH_REMATCH[1]}"
        symbol=""
    fi

    if [[ -n "$package" ]]; then
        handle_direct_query "$package" "$symbol"
    fi
}

handle_compile_error() {
    local output=$1

    echo "=== Processing Compilation Error ===" >&2
    echo "Error output:" >&2
    echo "$output" | head -3 >&2
    echo ""

    # Trigger detection and query
    "${SCRIPT_DIR}/trigger-logic.sh" compilation-error "$output" || true

    # Try to extract package from error
    local package
    if echo "$output" | grep -q "undefined:"; then
        package=$(echo "$output" | grep -o '[a-zA-Z_][a-zA-Z0-9_]*\.' | head -1 | sed 's/.$//')
        if [[ -n "$package" ]]; then
            handle_direct_query "$package" ""
        fi
    fi
}

handle_deprecated_check() {
    local code=$1

    echo "=== Checking for Deprecated APIs ===" >&2
    echo ""

    # Trigger detection
    "${SCRIPT_DIR}/trigger-logic.sh" deprecated-api "$code" || true
}

main() {
    if [[ $# -eq 0 ]]; then
        show_usage
        return 1
    fi

    local mode=$1
    shift || true

    case "$mode" in
        --lsp)
            if [[ $# -eq 0 ]]; then
                echo "ERROR: --lsp requires a message argument" >&2
                return 1
            fi
            handle_lsp_error "$1"
            ;;
        --compile-error)
            if [[ $# -eq 0 ]]; then
                echo "ERROR: --compile-error requires output argument" >&2
                return 1
            fi
            handle_compile_error "$1"
            ;;
        --check-deprecated)
            if [[ $# -eq 0 ]]; then
                echo "ERROR: --check-deprecated requires code argument" >&2
                return 1
            fi
            handle_deprecated_check "$1"
            ;;
        --query)
            if [[ $# -eq 0 ]]; then
                echo "ERROR: --query requires package argument" >&2
                return 1
            fi
            handle_direct_query "$@"
            ;;
        --help)
            show_usage
            ;;
        *)
            echo "ERROR: Unknown option: $mode" >&2
            show_usage
            return 1
            ;;
    esac
}

main "$@"
