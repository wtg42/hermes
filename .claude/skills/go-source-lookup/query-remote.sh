#!/usr/bin/env bash
# query-remote.sh
# Query Go packages from remote sources (pkg.go.dev)
# Usage: query-remote.sh <package> [symbol]

set -euo pipefail

query_pkggodev() {
    local package=$1
    local symbol=${2:-}
    local url="https://pkg.go.dev/${package}"

    if [[ -n "$symbol" ]]; then
        url="${url}#${symbol}"
    fi

    # Try to fetch basic package info
    echo "=== Remote Query: ${package}${symbol:+.}${symbol} ==="
    echo "URL: $url"
    echo ""
    echo "Package: $package"
    echo "Symbol: ${symbol:-*}"
    echo "Source: https://pkg.go.dev"
    echo ""
    echo "To view full documentation, visit:"
    echo "$url"
}

main() {
    if [[ $# -lt 1 ]]; then
        echo "Usage: query-remote.sh <package> [symbol]" >&2
        return 1
    fi

    local package=$1
    local symbol=${2:-}

    query_pkggodev "$package" "$symbol"
}

main "$@"
