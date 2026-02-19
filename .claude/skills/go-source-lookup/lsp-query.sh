#!/usr/bin/env bash
# lsp-query.sh
# LSP query interface for gopls using command-line tools
# Handles symbol queries via gopls definition, hover, and reference commands
# Usage: lsp-query.sh --hover <file> <line> <col> | --definition <file> <line> <col> | --references <file> <line> <col>

set -euo pipefail

# ============================================================================
# Configuration
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOPLS_PATH="${GOPLS_PATH:-/home/weiting/.local/share/nvim/mason/packages/gopls/gopls}"
LSP_CACHE_DIR="${HOME}/.cache/go-source-lookup/lsp"
LSP_LOG="${LSP_CACHE_DIR}/gopls.log"

# ============================================================================
# Initialization
# ============================================================================

init_cache() {
    mkdir -p "$LSP_CACHE_DIR"
}

check_gopls() {
    if [[ ! -x "$GOPLS_PATH" ]]; then
        # Try to find gopls in PATH
        if ! GOPLS_PATH=$(command -v gopls); then
            echo "ERROR: gopls not found. Install it with: go install github.com/golang/tools/gopls@latest" >&2
            return 1
        fi
    fi
}

# ============================================================================
# Cache Management
# ============================================================================

get_cache_key() {
    local file=$1
    local line=$2
    local col=$3

    # Create cache key from file:line:col
    echo "$file:$line:$col" | sha256sum | cut -d' ' -f1
}

get_cache_file() {
    local file=$1
    local line=$2
    local col=$3
    local cache_key
    cache_key=$(get_cache_key "$file" "$line" "$col")

    echo "$LSP_CACHE_DIR/$cache_key.cache"
}

is_cache_valid() {
    local cache_file=$1
    local ttl=900  # 15 minutes

    if [[ ! -f "$cache_file" ]]; then
        return 1
    fi

    local file_age=$(($(date +%s) - $(stat -c%Y "$cache_file" 2>/dev/null || stat -f%m "$cache_file")))
    if [[ $file_age -gt $ttl ]]; then
        return 1
    fi

    return 0
}

get_from_cache() {
    local file=$1
    local line=$2
    local col=$3
    local cache_file
    cache_file=$(get_cache_file "$file" "$line" "$col")

    if is_cache_valid "$cache_file"; then
        cat "$cache_file"
        return 0
    fi

    return 1
}

save_to_cache() {
    local file=$1
    local line=$2
    local col=$3
    local content=$4
    local cache_file
    cache_file=$(get_cache_file "$file" "$line" "$col")

    echo "$content" > "$cache_file"
}

# ============================================================================
# Utility Functions
# ============================================================================

get_working_dir() {
    local file=$1

    # If file is in /tmp, use /tmp as working dir
    if [[ "$file" == /tmp/* ]]; then
        dirname "$file"
    else
        # Otherwise try to find Go module root
        local dir
        dir=$(cd "$(dirname "$file")" && pwd)

        # Look for go.mod upwards
        while [[ "$dir" != "/" ]]; do
            if [[ -f "$dir/go.mod" ]]; then
                echo "$dir"
                return 0
            fi
            dir=$(dirname "$dir")
        done

        # Default to file directory
        echo "$(dirname "$file")"
    fi
}

# ============================================================================
# LSP Query Methods using gopls commands
# ============================================================================

lsp_hover() {
    local file=$1
    local line=$2
    local col=$3
    local work_dir
    work_dir=$(get_working_dir "$file")

    # Check cache first
    if get_from_cache "$file" "$line" "$col"; then
        log_info "Cache hit: $file:$line:$col (hover)"
        return 0
    fi

    # Use gopls definition command at position
    cd "$work_dir" || return 1

    local result
    result=$("$GOPLS_PATH" definition "$file:$line:$col" 2>>"$LSP_LOG" || echo "")

    if [[ -z "$result" ]]; then
        # Fallback: try to get type info using gopls signature
        result=$("$GOPLS_PATH" signature "$file:$line:$col" 2>>"$LSP_LOG" || echo "")
    fi

    if [[ -n "$result" ]]; then
        save_to_cache "$file" "$line" "$col" "$result"
        echo "$result"
    else
        log_info "No information available at $file:$line:$col"
        echo "No information available at $file:$line:$col"
        return 1
    fi
}

lsp_definition() {
    local file=$1
    local line=$2
    local col=$3
    local work_dir
    work_dir=$(get_working_dir "$file")

    # Check cache first
    if get_from_cache "$file" "$line" "$col"; then
        log_info "Cache hit: $file:$line:$col (definition)"
        return 0
    fi

    cd "$work_dir" || return 1

    local result
    result=$("$GOPLS_PATH" definition "$file:$line:$col" 2>>"$LSP_LOG" || echo "")

    if [[ -n "$result" ]]; then
        save_to_cache "$file" "$line" "$col" "$result"
        echo "$result"
    else
        log_info "No definition found at $file:$line:$col"
        echo "No definition found at $file:$line:$col"
        return 1
    fi
}

lsp_references() {
    local file=$1
    local line=$2
    local col=$3
    local work_dir
    work_dir=$(get_working_dir "$file")

    # Check cache first
    if get_from_cache "$file" "$line" "$col"; then
        log_info "Cache hit: $file:$line:$col (references)"
        return 0
    fi

    cd "$work_dir" || return 1

    local result
    result=$("$GOPLS_PATH" references "$file:$line:$col" 2>>"$LSP_LOG" || echo "")

    if [[ -n "$result" ]]; then
        save_to_cache "$file" "$line" "$col" "$result"
        echo "$result"
    else
        log_info "No references found at $file:$line:$col"
        echo "No references found at $file:$line:$col"
        return 1
    fi
}

# ============================================================================
# Error Handling & Logging
# ============================================================================

log_error() {
    local msg=$1
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $msg" >> "$LSP_LOG"
    echo "ERROR: $msg" >&2
}

log_info() {
    local msg=$1
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $msg" >> "$LSP_LOG"
}

fallback_query() {
    local package=$1
    local symbol=$2

    # Fallback to go doc if gopls fails
    log_info "Falling back to go doc for $package.$symbol"

    if [[ -n "$symbol" ]]; then
        go doc "$package.$symbol" 2>/dev/null || echo "Symbol not found: $package.$symbol"
    else
        go doc "$package" 2>/dev/null || echo "Package not found: $package"
    fi
}

# ============================================================================
# CLI Entry Point
# ============================================================================

show_usage() {
    cat <<EOF
Usage: lsp-query.sh [options]

Options:
  --hover <file> <line> <col>          Get hover information at position
  --definition <file> <line> <col>     Jump to definition
  --references <file> <line> <col>     Find all references
  --help                               Show this help message

Examples:
  lsp-query.sh --hover main.go 10 5
  lsp-query.sh --definition main.go 15 10
  lsp-query.sh --references main.go 20 3

Log file: $LSP_LOG
EOF
}

main() {
    if [[ $# -lt 1 ]]; then
        show_usage
        return 1
    fi

    init_cache
    check_gopls || {
        log_error "gopls not available"
        return 1
    }

    log_info "gopls path: $GOPLS_PATH"

    local mode=$1
    shift || true

    case "$mode" in
        --hover)
            if [[ $# -lt 3 ]]; then
                log_error "--hover requires <file> <line> <col>"
                return 1
            fi
            lsp_hover "$1" "$2" "$3" || fallback_query "unknown" "unknown"
            ;;
        --definition)
            if [[ $# -lt 3 ]]; then
                log_error "--definition requires <file> <line> <col>"
                return 1
            fi
            lsp_definition "$1" "$2" "$3"
            ;;
        --references)
            if [[ $# -lt 3 ]]; then
                log_error "--references requires <file> <line> <col>"
                return 1
            fi
            lsp_references "$1" "$2" "$3"
            ;;
        --help)
            show_usage
            ;;
        *)
            log_error "Unknown option: $mode"
            show_usage
            return 1
            ;;
    esac
}

main "$@"
