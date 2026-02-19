#!/usr/bin/env bash
# format-for-llm.sh
# Format query results for LLM consumption
# Usage: format-for-llm.sh <package> [symbol] [go-version]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

format_for_llm() {
    local package=$1
    local symbol=${2:-}
    local go_version=${3:-}

    # Get query results
    local query_result
    query_result=$("${SCRIPT_DIR}/go-source-lookup.sh" "$package" "$symbol" 2>/dev/null || echo "")

    if [[ -z "$query_result" ]]; then
        cat <<EOF
## Go Source Lookup Result

**Status:** Query failed

**Package:** $package
**Symbol:** ${symbol:-all}
**Go Version:** ${go_version:-unknown}

No results found. The package or symbol may not exist.
EOF
        return 1
    fi

    # Detect Go version if not provided
    if [[ -z "$go_version" ]]; then
        go_version=$(go version | awk '{print $3}')
    fi

    # Format result
    cat <<EOF
## Go Source Lookup Result

**Package:** $package
**Symbol:** ${symbol:-API Overview}
**Go Version:** $go_version
**Source:** Standard library / Dependency

### Function/Type Signature

\`\`\`go
$(echo "$query_result" | head -5)
\`\`\`

### Documentation

\`\`\`
$(echo "$query_result" | tail -n +2)
\`\`\`

---

**LLM Recommendations:**
1. Use this exact signature when calling the function
2. Check doc comments for parameter meanings
3. Look for "Deprecated" markers for outdated APIs
4. If deprecated, use the recommended replacement from the doc comments
EOF
}

format_multiple_results() {
    local packages=$1

    echo "## Multiple Query Results"
    echo ""
    echo "Packages queried:"
    echo "$packages" | while read -r pkg; do
        echo "- $pkg"
    done
}

main() {
    if [[ $# -lt 1 ]]; then
        echo "Usage: format-for-llm.sh <package> [symbol] [go-version]" >&2
        return 1
    fi

    local package=$1
    local symbol=${2:-}
    local go_version=${3:-}

    format_for_llm "$package" "$symbol" "$go_version"
}

main "$@"
