# go-source-lookup Skill

Query Go standard library and package source code documentation. Automatically triggers on LSP warnings, compilation errors, or deprecated APIs.

## Overview

This skill enables Claude Code to:
- Query Go standard library source code using `go doc`
- Query installed dependencies from `$GOPATH/pkg/mod`
- Query remote packages from pkg.go.dev
- Automatically detect LSP warnings, compilation errors, and deprecated APIs
- Cache query results (15-minute TTL)
- Format results for LLM consumption

## Architecture

```
llm-integration.sh (main entry point)
  ├── go-source-lookup.sh (core query logic)
  │   ├── Query stdlib (go doc)
  │   ├── Query dependencies
  │   └── Cache management
  ├── trigger-logic.sh (automatic detection)
  │   ├── LSP warning detection
  │   ├── Compilation error parsing
  │   ├── Deprecated API detection
  │   └── Query deduplication
  ├── query-remote.sh (remote package queries)
  │   └── pkg.go.dev queries
  └── format-for-llm.sh (result formatting)
      └── Markdown output for LLM
```

## Usage

### For LLM Integration (Internal)

```bash
# Direct query
llm-integration.sh --query fmt Println

# LSP warning detection
llm-integration.sh --lsp "undefined: fmt.Println"

# Compilation error
llm-integration.sh --compile-error "error: cannot use..."

# Deprecated API check
llm-integration.sh --check-deprecated "code containing Deprecated API"
```

### For Manual Query

```bash
# Query standard library
./go-source-lookup.sh fmt Println
./go-source-lookup.sh encoding/json Unmarshal

# Query with remote fallback
./go-source-lookup.sh github.com/some/package SomeFunction --remote
```

### For Testing

```bash
./test.sh
```

## Files

- **go-source-lookup.sh** — Main query logic (stdlib, dependencies, caching)
- **trigger-logic.sh** — LSP/compile/deprecated detection and deduplication
- **llm-integration.sh** — LLM integration entry point
- **query-remote.sh** — Remote package queries
- **format-for-llm.sh** — Result formatting for LLM
- **test.sh** — Test suite
- **SKILL.md** — Skill configuration

## Environment Requirements

- Go 1.18+ installed
- `$GOROOT` and `$GOPATH` properly configured
- `go` command available in PATH

## Cache

Query results are cached in `~/.cache/go-source-lookup/` with 15-minute TTL.

Deduplication cache is in `~/.cache/go-source-lookup/dedup/` to prevent duplicate queries within 15 minutes.

## Limitations

- Remote queries currently provide URLs (actual content fetching requires WebFetch tool integration)
- Deprecated API detection relies on `// Deprecated:` comments in source code
- Limited pattern matching for LSP/compilation error extraction
