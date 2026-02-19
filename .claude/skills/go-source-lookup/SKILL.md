---
name: go-source-lookup
description: Query Go standard library and package source code. Automatically triggers on LSP warnings, compilation errors, or deprecated APIs.
license: MIT
compatibility: Requires Go environment ($GOROOT, $GOPATH)
metadata:
  author: claude-code
  version: "1.0"
  language: bash
---

Internal skill for Claude Code to automatically query Go source code and API documentation.

**Capabilities:**
- Query Go standard library source code (go doc)
- Query installed dependencies (go list -json -deps)
- Automatic trigger on LSP warnings, compilation errors, deprecated APIs
- Result caching (15-minute TTL)

**Environment:**
- $GOROOT: Go standard library location
- $GOPATH: Package cache location
- go command available in PATH

**Internal Hooks:**
- `lsp-warning`: Detect LSP warnings and trigger queries
- `compilation-error`: Parse go build/test errors
- `deprecated-api`: Detect and query deprecated APIs
