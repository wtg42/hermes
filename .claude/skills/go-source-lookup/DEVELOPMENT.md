# Development Guide - go-source-lookup

## Architecture Overview

### Query Path

```
User/LLM Request
  ↓
llm-integration.sh (router)
  ↓
go-source-lookup.sh (core logic)
  ├─ Check cache
  ├─ Try stdlib (go doc)
  ├─ Try dependencies (go list -json -m)
  └─ Try remote (query-remote.sh)
  ↓
format-for-llm.sh (format output)
  ↓
Return to LLM
```

### Trigger Path

```
LSP Warning / Compilation Error / Deprecated API
  ↓
trigger-logic.sh (detect & deduplicate)
  ├─ Parse error/warning
  ├─ Check dedup cache
  └─ Call go-source-lookup.sh
  ↓
Result returned to LLM
```

## Key Components

### 1. go-source-lookup.sh
- **Purpose**: Core query logic
- **Functions**:
  - `detect_go_env()`: Get GOROOT, GOPATH, Go version
  - `init_cache()`: Setup cache directory
  - `get_cache_key()`: Hash package:symbol to cache key
  - `is_cache_valid()`: Check if cache entry is still fresh
  - `query_stdlib()`: Use `go doc` to query stdlib
  - `query_dependency()`: Use `go list -json -m` to find package info
  - `query_remote()`: Fallback to remote queries

### 2. trigger-logic.sh
- **Purpose**: Detect and deduplicate trigger events
- **Functions**:
  - `handle_lsp_warning()`: Parse LSP error messages
  - `handle_compilation_error()`: Parse `go build` output
  - `handle_deprecated_api()`: Detect Deprecated comments
  - `is_duplicate_query()`: Check dedup cache
  - `mark_query_processed()`: Mark as processed

### 3. llm-integration.sh
- **Purpose**: Main entry point for LLM
- **Routes**:
  - `--query`: Direct query
  - `--lsp`: LSP warning handling
  - `--compile-error`: Compilation error handling
  - `--check-deprecated`: Deprecated API detection

### 4. format-for-llm.sh
- **Purpose**: Format results for LLM consumption
- **Output**: Markdown with:
  - Function signature
  - Documentation
  - Go version
  - Recommendations

## Extending the Skill

### Adding New Error Pattern Recognition

Edit `trigger-logic.sh`:

```bash
handle_lsp_warning() {
    if echo "$lsp_message" | grep -q "new_pattern"; then
        echo "Type: New error type"
        # Extract and query
    fi
}
```

### Adding New Query Source

1. Create new function in `go-source-lookup.sh`:
```bash
query_custom_source() {
    local package=$1
    local symbol=$2
    # Implementation
}
```

2. Add to `lookup()` function:
```bash
if result=$(query_custom_source "$package" "$symbol"); then
    save_to_cache "$package" "$symbol" "$result"
    echo "$result"
    return 0
fi
```

### Improving Cache Strategy

Edit cache TTL in:
- `go-source-lookup.sh`: `CACHE_TTL=900`
- `trigger-logic.sh`: `DEDUP_TTL=900`

Current: 900 seconds (15 minutes)

### Integrating WebFetch

The `query_remote()` function can be enhanced to use WebFetch:

```bash
query_remote() {
    local package=$1
    local symbol=$2
    # Call WebFetch to fetch pkg.go.dev content
}
```

## Testing

Run the test suite:

```bash
./test.sh
```

### Manual Testing

```bash
# Test stdlib query
./go-source-lookup.sh fmt Println

# Test LSP trigger
./trigger-logic.sh lsp-warning "undefined: fmt.Println"

# Test LLM integration
./llm-integration.sh --query encoding/json Unmarshal

# Test formatting
./format-for-llm.sh fmt Println
```

## Performance Considerations

1. **Cache**: 15-minute TTL reduces redundant `go doc` calls
2. **Deduplication**: Prevents duplicate queries from LSP spam
3. **Local First**: Prioritizes local queries over remote
4. **Pattern Matching**: Simple regex to avoid expensive parsing

## Known Limitations

1. Remote queries don't fetch actual content (requires WebFetch integration)
2. Deprecated detection depends on source code format
3. LSP/compilation error parsing is pattern-based, may miss edge cases
4. No support for vendored dependencies

## Future Improvements

- [ ] WebFetch integration for remote content
- [ ] Better error message parsing (AI-powered?)
- [ ] Support for Go modules with multiple versions
- [ ] Performance profiling and optimization
- [ ] Integration with GoLand/VSCode plugins
- [ ] Batch query support
- [ ] Query result curation and ranking
