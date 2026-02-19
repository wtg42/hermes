#!/usr/bin/env bash
# go-source-lookup.sh
# Query Go source code from stdlib or dependencies
# Usage: go-source-lookup.sh <package> <symbol> [--remote]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CACHE_DIR="${HOME}/.cache/go-source-lookup"
CACHE_TTL=900  # 15 minutes in seconds

# ============================================================================
# Environment Detection
# ============================================================================

detect_go_env() {
    local go_root go_path go_version

    # Detect $GOROOT
    go_root="$(go env GOROOT)"
    if [[ -z "$go_root" ]]; then
        echo "ERROR: GOROOT not found. Go environment not properly configured." >&2
        return 1
    fi

    # Detect $GOPATH
    go_path="$(go env GOPATH)"
    if [[ -z "$go_path" ]]; then
        go_path="$HOME/go"
    fi

    # Detect Go version
    go_version="$(go version | awk '{print $3}')"

    echo "GOROOT=$go_root"
    echo "GOPATH=$go_path"
    echo "GO_VERSION=$go_version"
}

# ============================================================================
# Caching Mechanism
# ============================================================================

init_cache() {
    mkdir -p "$CACHE_DIR"
}

get_cache_key() {
    local package=$1
    local symbol=$2
    echo "$(echo "${package}:${symbol}" | sha256sum | cut -d' ' -f1)"
}

is_cache_valid() {
    local cache_file=$1
    if [[ ! -f "$cache_file" ]]; then
        return 1
    fi

    local file_age=$(($(date +%s) - $(stat -f%m "$cache_file" 2>/dev/null || stat -c%Y "$cache_file" 2>/dev/null)))
    if [[ $file_age -gt $CACHE_TTL ]]; then
        return 1
    fi
    return 0
}

get_from_cache() {
    local package=$1
    local symbol=$2
    local cache_key
    cache_key=$(get_cache_key "$package" "$symbol")
    local cache_file="$CACHE_DIR/${cache_key}"

    if is_cache_valid "$cache_file"; then
        cat "$cache_file"
        return 0
    fi
    return 1
}

save_to_cache() {
    local package=$1
    local symbol=$2
    local content=$3
    local cache_key
    cache_key=$(get_cache_key "$package" "$symbol")
    local cache_file="$CACHE_DIR/${cache_key}"

    echo "$content" > "$cache_file"
}

# ============================================================================
# Query Implementations
# ============================================================================

query_stdlib() {
    local package=$1
    local symbol=$2
    local query_str="${package}"

    if [[ -n "$symbol" ]]; then
        query_str="${package}.${symbol}"
    fi

    # Use go doc to query stdlib
    go doc "$query_str" 2>/dev/null || {
        echo "ERROR: Symbol not found in standard library: ${query_str}" >&2
        return 1
    }
}

extract_version_info() {
    local package=$1

    # Try go list -json -m to get version
    if [[ -f "go.mod" ]]; then
        go list -json -m "$package" 2>/dev/null | grep -o '"Version":"[^"]*"' | cut -d'"' -f4 || echo "unknown"
    else
        echo "not installed"
    fi
}

query_dependency() {
    local package=$1
    local symbol=$2

    # Try to query using go doc (might work if package is importable)
    local query_str="${package}"
    if [[ -n "$symbol" ]]; then
        query_str="${package}.${symbol}"
    fi

    # Get version info
    local version
    version=$(extract_version_info "$package")

    go doc "$query_str" 2>/dev/null || {
        echo "INFO: go doc query failed for dependency"
        echo "Package: $package"
        echo "Version: $version"
        return 2  # Partial success - got version info
    }
}

query_remote() {
    local package=$1
    local symbol=$2
    local script_dir
    script_dir="$(dirname "${BASH_SOURCE[0]}")"

    if [[ -x "${script_dir}/query-remote.sh" ]]; then
        "${script_dir}/query-remote.sh" "$package" "$symbol"
    else
        # Fallback: provide URL guidance
        local query_str="${package}"
        if [[ -n "$symbol" ]]; then
            query_str="${package}.${symbol}"
        fi

        echo "=== Remote Query: ${query_str} ==="
        echo "Package: $package"
        echo "Symbol: ${symbol:-*}"
        echo "URL: https://pkg.go.dev/${query_str}"
    fi
}

# ============================================================================
# Main Lookup Logic
# ============================================================================

lookup() {
    local package=$1
    local symbol=${2:-}
    local use_remote=${3:-}

    init_cache

    # Check cache first
    if cached_result=$(get_from_cache "$package" "$symbol"); then
        echo "=== CACHED RESULT ==="
        echo "$cached_result"
        return 0
    fi

    echo "=== Looking up: ${package}${symbol:+.}${symbol} ==="

    # Try stdlib first
    if [[ "$package" == "fmt" || "$package" == "encoding/json" || "$package" == "os" || "$package" == "io" ]]; then
        if result=$(query_stdlib "$package" "$symbol"); then
            save_to_cache "$package" "$symbol" "$result"
            echo "$result"
            return 0
        fi
    fi

    # Try dependencies
    if result=$(query_dependency "$package" "$symbol" 2>/dev/null); then
        save_to_cache "$package" "$symbol" "$result"
        echo "$result"
        return 0
    fi

    # Try remote if requested or if local lookup failed
    if [[ -n "$use_remote" ]]; then
        query_remote "$package" "$symbol"
        return 0
    fi

    echo "ERROR: Could not find symbol: ${package}.${symbol:-*}" >&2
    return 1
}

# ============================================================================
# CLI Entry Point
# ============================================================================

main() {
    if [[ $# -lt 1 ]]; then
        echo "Usage: go-source-lookup.sh <package> [symbol] [--remote]" >&2
        echo "Example: go-source-lookup.sh fmt Println" >&2
        echo "Example: go-source-lookup.sh encoding/json Unmarshal" >&2
        return 1
    fi

    local package=$1
    local symbol=${2:-}
    local use_remote=${3:-}

    # Detect environment
    eval "$(detect_go_env)"

    # Run lookup
    lookup "$package" "$symbol" "$use_remote"
}

main "$@"
