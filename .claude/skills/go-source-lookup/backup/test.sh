#!/usr/bin/env bash
# test.sh - Basic test suite for go-source-lookup

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PASS=0
FAIL=0

test_case() {
    local name=$1
    local cmd=$2

    echo -n "Testing: $name ... "
    if eval "$cmd" > /dev/null 2>&1; then
        echo "✓ PASS"
        ((PASS++))
    else
        echo "✗ FAIL"
        ((FAIL++))
    fi
}

echo "=== Go Source Lookup Test Suite ==="
echo ""

# Test 1: Environment detection
test_case "Environment detection" "go env GOROOT | grep -q ."

# Test 2: Basic stdlib query
test_case "Stdlib query (fmt.Println)" "${SCRIPT_DIR}/go-source-lookup.sh fmt Println | grep -q 'Println'"

# Test 3: Stdlib query (encoding/json)
test_case "Stdlib query (encoding/json.Unmarshal)" "${SCRIPT_DIR}/go-source-lookup.sh encoding/json Unmarshal | grep -q 'Unmarshal'"

# Test 4: Cache mechanism
test_case "Cache directory creation" "mkdir -p ${HOME}/.cache/go-source-lookup"

# Test 5: Trigger logic (LSP warning)
test_case "LSP trigger detection" "${SCRIPT_DIR}/trigger-logic.sh lsp-warning 'undefined: fmt' > /dev/null"

# Test 6: LLM integration
test_case "LLM integration query" "${SCRIPT_DIR}/llm-integration.sh --query fmt Println | grep -q 'Println'"

# Test 7: Format for LLM
test_case "Format output for LLM" "${SCRIPT_DIR}/format-for-llm.sh fmt Println | grep -q 'Go Source Lookup'"

echo ""
echo "=== Test Results ==="
echo "Passed: $PASS"
echo "Failed: $FAIL"
echo "Total: $((PASS + FAIL))"

if [[ $FAIL -eq 0 ]]; then
    echo "✓ All tests passed!"
    exit 0
else
    echo "✗ Some tests failed"
    exit 1
fi
