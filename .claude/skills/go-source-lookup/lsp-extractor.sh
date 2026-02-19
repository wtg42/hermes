#!/usr/bin/env bash
# lsp-extractor.sh
# Extract and format symbol information from LSP results
# Handles signature extraction, documentation cleaning, type info, deprecated API detection
# Usage: lsp-extractor.sh --format-result <lsp-output> | --extract-signature <output> | --detect-deprecated <output>

set -euo pipefail

# ============================================================================
# Signature Extraction
# ============================================================================

extract_signature() {
    local lsp_output=$1

    # Extract function signature from gopls output
    # Format: "filename:line:col: symbol type definition"
    local signature
    signature=$(echo "$lsp_output" | head -1 | sed 's/.*: //')

    if [[ -n "$signature" ]]; then
        echo "### Function Signature"
        echo ""
        echo "\`\`\`go"
        echo "$signature"
        echo "\`\`\`"
        echo ""
    fi
}

# ============================================================================
# Documentation Extraction & Cleaning
# ============================================================================

extract_documentation() {
    local lsp_output=$1

    # Extract documentation from gopls output
    # Usually follows after the signature
    local doc
    doc=$(echo "$lsp_output" | tail -n +2 | grep -v "^[[:space:]]*$" | head -10 || echo "")

    if [[ -n "$doc" ]]; then
        echo "### Documentation"
        echo ""
        echo "$doc" | while read -r line; do
            # Remove Markdown formatting for LLM clarity
            line="${line//#/\\#}"  # Escape Markdown headers
            line="${line//**/}"    # Remove bold
            line="${line//__/}"    # Remove italic
            echo "$line"
        done
        echo ""
    fi
}

# ============================================================================
# Type Information Extraction
# ============================================================================

extract_type_info() {
    local lsp_output=$1

    # Try to extract type information from gopls output
    local type_info
    type_info=$(echo "$lsp_output" | grep -o 'type [^ ]*' | head -1 || echo "")

    if [[ -n "$type_info" ]]; then
        echo "### Type Information"
        echo ""
        echo "- $type_info"
        echo ""
    fi
}

# ============================================================================
# Deprecated API Detection
# ============================================================================

detect_deprecated() {
    local lsp_output=$1

    # Detect "Deprecated" markers in gopls output
    if echo "$lsp_output" | grep -iq "deprecated"; then
        echo "### ⚠️ Deprecated API"
        echo ""
        echo "This API is marked as deprecated."
        echo ""

        # Try to find replacement suggestion
        local replacement
        replacement=$(echo "$lsp_output" | grep -i "use\|instead\|replace" | head -1 || echo "")

        if [[ -n "$replacement" ]]; then
            echo "**Suggested replacement:**"
            echo ""
            echo "$replacement"
            echo ""
        fi

        return 0
    fi

    return 1
}

# ============================================================================
# Format Complete Result for LLM
# ============================================================================

format_result() {
    local lsp_output=$1
    local package=${2:-}
    local symbol=${3:-}
    local go_version=${4:-$(go version | awk '{print $3}')}

    cat <<EOF
## Go Symbol Query Result

**Package:** ${package:-unknown}
**Symbol:** ${symbol:-API}
**Go Version:** $go_version

EOF

    extract_signature "$lsp_output"
    extract_type_info "$lsp_output"
    extract_documentation "$lsp_output"

    if detect_deprecated "$lsp_output"; then
        : # deprecated info already printed
    fi

    cat <<EOF
---

**LLM Notes:**
1. Use the exact signature above when calling the function
2. Check parameter types and return values carefully
3. Review documentation for important behavior details
4. If deprecated, use the suggested replacement
EOF
}

# ============================================================================
# CLI Entry Point
# ============================================================================

show_usage() {
    cat <<EOF
Usage: lsp-extractor.sh [options]

Options:
  --format-result <output> [package] [symbol] [version]
                           Format LSP result for LLM consumption
  --extract-signature <output>
                           Extract just the function signature
  --detect-deprecated <output>
                           Check if API is deprecated
  --help                   Show this help message

Examples:
  lsp-extractor.sh --format-result "func Println(...)" fmt Println
  lsp-extractor.sh --detect-deprecated "Deprecated: use X instead"
EOF
}

main() {
    if [[ $# -lt 1 ]]; then
        show_usage
        return 1
    fi

    local mode=$1
    shift || true

    case "$mode" in
        --format-result)
            if [[ $# -lt 1 ]]; then
                echo "ERROR: --format-result requires output argument" >&2
                return 1
            fi
            format_result "$@"
            ;;
        --extract-signature)
            if [[ $# -lt 1 ]]; then
                echo "ERROR: --extract-signature requires output argument" >&2
                return 1
            fi
            extract_signature "$1"
            ;;
        --detect-deprecated)
            if [[ $# -lt 1 ]]; then
                echo "ERROR: --detect-deprecated requires output argument" >&2
                return 1
            fi
            detect_deprecated "$1"
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
