#!/usr/bin/env bash
# trigger-logic.sh
# Detect LSP warnings, compilation errors, and deprecated APIs
# Usage: trigger-logic.sh <type> [context]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEDUP_DIR="${HOME}/.cache/go-source-lookup/dedup"
DEDUP_TTL=900  # 15 minutes

# ============================================================================
# Deduplication Logic
# ============================================================================

init_dedup() {
    mkdir -p "$DEDUP_DIR"
}

get_dedup_key() {
    echo "$1" | sha256sum | cut -d' ' -f1
}

is_duplicate_query() {
    local query=$1
    local key
    key=$(get_dedup_key "$query")
    local dedup_file="$DEDUP_DIR/${key}"

    if [[ -f "$dedup_file" ]]; then
        local file_age=$(($(date +%s) - $(stat -f%m "$dedup_file" 2>/dev/null || stat -c%Y "$dedup_file" 2>/dev/null)))
        if [[ $file_age -lt $DEDUP_TTL ]]; then
            return 0  # Is duplicate
        fi
    fi
    return 1  # Not duplicate
}

mark_query_processed() {
    local query=$1
    local key
    key=$(get_dedup_key "$query")
    local dedup_file="$DEDUP_DIR/${key}"
    touch "$dedup_file"
}

# ============================================================================
# LSP Warning Detection
# ============================================================================

handle_lsp_warning() {
    local lsp_message=$1

    echo "=== LSP Warning Detected ==="
    echo "Message: $lsp_message"
    echo ""

    # Parse common LSP error patterns
    if echo "$lsp_message" | grep -q "undefined"; then
        echo "Type: Undefined symbol"
        # Extract package and symbol from message
        # Format: "undefined: package.Symbol" or similar
        local match
        match=$(echo "$lsp_message" | grep -o '[a-zA-Z_][a-zA-Z0-9_.]*' | head -1)
        if [[ -n "$match" ]]; then
            trigger_query "$match" "lsp-undefined"
        fi
    elif echo "$lsp_message" | grep -q "wrong number of arguments"; then
        echo "Type: Wrong function signature"
        local match
        match=$(echo "$lsp_message" | grep -o '[a-zA-Z_][a-zA-Z0-9_.]*' | head -1)
        if [[ -n "$match" ]]; then
            trigger_query "$match" "lsp-signature"
        fi
    fi
}

# ============================================================================
# Compilation Error Detection
# ============================================================================

handle_compilation_error() {
    local error_output=$1

    echo "=== Compilation Error Detected ==="
    echo "$error_output" | head -5
    echo ""

    # Parse go build error patterns
    if echo "$error_output" | grep -q "undefined"; then
        local match
        match=$(echo "$error_output" | grep -o '[a-zA-Z_][a-zA-Z0-9_.]*' | head -1)
        if [[ -n "$match" ]]; then
            trigger_query "$match" "compile-undefined"
        fi
    elif echo "$error_output" | grep -q "cannot use"; then
        echo "Type: Type mismatch"
        echo "Recommendation: Check function signature with go doc"
    elif echo "$error_output" | grep -q "no such"; then
        echo "Type: Missing package or symbol"
        local match
        match=$(echo "$error_output" | grep -o '[a-zA-Z_][a-zA-Z0-9_.]*' | head -1)
        if [[ -n "$match" ]]; then
            trigger_query "$match" "compile-missing"
        fi
    fi
}

# ============================================================================
# Deprecated API Detection
# ============================================================================

handle_deprecated_api() {
    local source_code=$1

    echo "=== Deprecated API Detection ==="

    # Look for Deprecated markers in go doc comments
    if echo "$source_code" | grep -q "Deprecated"; then
        echo "Found Deprecated API usage"
        echo ""

        # Extract the deprecated function/type
        local deprecated_item
        deprecated_item=$(echo "$source_code" | grep -B2 "Deprecated" | head -1)
        echo "Item: $deprecated_item"

        # Try to find replacement in comments
        local replacement
        replacement=$(echo "$source_code" | grep -A2 "Deprecated" | grep -i "use\|instead\|replace" | head -1)
        if [[ -n "$replacement" ]]; then
            echo "Recommendation: $replacement"
        fi
    fi
}

# ============================================================================
# Query Trigger
# ============================================================================

trigger_query() {
    local package=$1
    local trigger_type=$2
    local query_key="${package}:${trigger_type}"

    # Check for duplicates
    if is_duplicate_query "$query_key"; then
        echo "INFO: Query already triggered recently, skipping"
        return 0
    fi

    mark_query_processed "$query_key"

    echo "Triggering query for: $package"
    echo ""

    # Call the main lookup script
    if [[ -x "${SCRIPT_DIR}/go-source-lookup.sh" ]]; then
        "${SCRIPT_DIR}/go-source-lookup.sh" "$package" "" "" || true
    fi
}

# ============================================================================
# Main Entry Point
# ============================================================================

main() {
    if [[ $# -lt 1 ]]; then
        echo "Usage: trigger-logic.sh <type> [context]" >&2
        echo "Types: lsp-warning, compilation-error, deprecated-api" >&2
        return 1
    fi

    init_dedup

    local trigger_type=$1
    local context=${2:-}

    case "$trigger_type" in
        lsp-warning)
            handle_lsp_warning "$context"
            ;;
        compilation-error)
            handle_compilation_error "$context"
            ;;
        deprecated-api)
            handle_deprecated_api "$context"
            ;;
        *)
            echo "ERROR: Unknown trigger type: $trigger_type" >&2
            return 1
            ;;
    esac
}

main "$@"
