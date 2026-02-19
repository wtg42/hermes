---
name: go-source-lookup
description: Query Go symbols using LSP and traditional methods. Automatically triggers on LSP warnings, compilation errors, or deprecated APIs.
license: MIT
compatibility: Requires Go environment ($GOROOT, $GOPATH, gopls)
metadata:
  author: claude-code
  version: "2.0"
  language: bash
---

Internal skill for Claude Code to automatically query Go source code and API documentation with LSP support.

**Capabilities:**
- **LSP-based symbol queries** via gopls (file:line:col position queries)
- **Query Go standard library** source code (go doc)
- **Query installed dependencies** (go list -json -deps)
- **Automatic trigger** on LSP warnings, compilation errors, deprecated APIs
- **Result caching** (15-minute TTL, coordinate-based for LSP)
- **Symbol extraction and formatting** for LLM consumption

**New Interfaces:**
- `llm-integration.sh --lsp-query <file> <line> <col>`: Direct LSP symbol query
- `lsp-query.sh --hover <file> <line> <col>`: Get symbol information at position
- `lsp-extractor.sh --format-result <output>`: Format symbol info for LLM

**Environment:**
- $GOROOT: Go standard library location
- $GOPATH: Package cache location
- `go` command available in PATH
- `gopls`: Language Server Protocol implementation (auto-discovered from Mason or $PATH)

**Internal Hooks:**
- `lsp-warning`: Detect LSP warnings and trigger LSP queries
- `compilation-error`: Parse go build/test errors
- `deprecated-api`: Detect and query deprecated APIs
