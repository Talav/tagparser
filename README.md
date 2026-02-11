# tagparser

[![tag](https://img.shields.io/github/tag/talav/tagparser.svg)](https://github.com/talav/tagparser/tags)
[![Go Reference](https://pkg.go.dev/badge/github.com/talav/tagparser.svg)](https://pkg.go.dev/github.com/talav/tagparser)
[![Go Report Card](https://goreportcard.com/badge/github.com/talav/tagparser)](https://goreportcard.com/report/github.com/talav/tagparser)
[![CI](https://github.com/talav/tagparser/workflows/Tagparser%20CI/badge.svg)](https://github.com/talav/tagparser/actions)
[![codecov](https://codecov.io/gh/talav/tagparser/graph/badge.svg)](https://codecov.io/gh/talav/tagparser)
[![License](https://img.shields.io/github/license/talav/tagparser)](./LICENSE)

A high-performance, production-ready parser for Go struct tags with comprehensive error reporting and zero-allocation options.

## Features

- **Strict parsing** with precise error positions
- **Zero-allocation mode** for performance-critical code
- **Quoted values** with escape sequences
- **Flexible syntax** supporting both named and options-only modes
- **Battle-tested** with extensive test coverage and fuzzing
- **DoS protection** with configurable size limits
- **No dependencies** (except testing)

## Installation

```bash
go get github.com/talav/tagparser
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/talav/tagparser"
)

func main() {
    // Parse a simple tag
    tag, err := tagparser.Parse(`json,omitempty,min=5`)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(tag.Options["min"]) // Output: 5
}
```

## Usage

### Basic Parsing

Parse tags treating all items as options:

```go
tag, err := tagparser.Parse(`foo,bar,boz='buzz fubar'`)
// tag.Name == ""
// tag.Options == map[string]string{"foo": "", "bar": "", "boz": "buzz fubar"}

tag2, _ := tagparser.Parse(`foo=bar,baz`)
// tag2.Name == ""
// tag2.Options == map[string]string{"foo": "bar", "baz": ""}
```

### Name Extraction

Parse tags with the first item as a name:

```go
tag, err := tagparser.ParseWithName(`json,omitempty,min=5`)
// tag.Name == "json"
// tag.Options == map[string]string{"omitempty": "", "min": "5"}

// If first item has equals, it's treated as a key-value pair
tag2, _ := tagparser.ParseWithName(`foo=bar,baz`)
// tag2.Name == ""
// tag2.Options == map[string]string{"foo": "bar", "baz": ""}
```

### Zero-Allocation Parsing

For performance-critical code, use callback-based parsing:

```go
opts := make(map[string]string, 4) // Pre-allocate

err := tagparser.ParseFunc(`json,omitempty,min=5`, func(key, value string) error {
    opts[key] = value
    return nil
})
// No allocations if map capacity is sufficient
```

With name extraction:

```go
var name string
opts := make(map[string]string)

err := tagparser.ParseFuncWithName(`json,omitempty,min=5`, func(key, value string) error {
    if key == "" {
        name = value // First item without equals
    } else {
        opts[key] = value
    }
    return nil
})
// name == "json"
// opts == map[string]string{"omitempty": "", "min": "5"}
```

### Real-World Examples

**JSON tags:**
```go
tag, _ := tagparser.ParseWithName(`"name,omitempty"`)
// Automatically handles Go struct tag quoting
// tag.Name == "name"
// tag.Options == map[string]string{"omitempty": ""}
```

**Validation tags:**
```go
tag, _ := tagparser.Parse(`required,email,min=8,max=100`)
// tag.Options == map[string]string{
//     "required": "",
//     "email": "",
//     "min": "8",
//     "max": "100",
// }
```

**Complex quoted values:**
```go
tag, _ := tagparser.Parse(`msg='Hello, World!',desc='It\'s great'`)
// tag.Options == map[string]string{
//     "msg": "Hello, World!",
//     "desc": "It's great",
// }
```

## Tag Syntax

### Basic Format

Tags are comma-separated items: `key1,key2=value2,key3='quoted value'`

### Quoting Rules

- **Bare words**: `foo=bar` (no quotes needed for simple values)
- **Single quotes**: `foo='bar, baz'` (for values with special characters)
- **Must enclose entirely**: `'foo bar'` ✅ but `foo'bar'` ❌

### Escape Sequences

- **In bare strings**: Escape commas and equals: `foo\=bar`, `foo\,bar`
- **In quoted strings**: Escape quotes and backslashes: `'foo\'bar'`, `'foo\\bar'`
- **Any non-alphanumeric**: `\!`, `\@`, `\#` all work

Examples:
```
foo\=bar     → "foo=bar"
'foo\'bar'   → "foo'bar"
'foo\\bar'   → "foo\bar"
foo=\=\,\!   → map["foo": "=,!"]
```

### Whitespace

- Leading/trailing ASCII whitespace is trimmed
- Escaped whitespace is preserved: `\ ` remains a space
- Example: ` foo = bar ` becomes `foo=bar`

### Special Cases

- **Empty values**: `key=` is valid (empty string value)
- **Duplicate keys**: Last value wins: `key=first,key=second` → `key=second`
- **Empty keys**: Not allowed (except for name in `ParseWithName`)
- **Empty input**: Returns empty Options map

## Error Handling

All parsing errors are returned as `*tagparser.Error` with:
- Original tag string
- Precise error position (1-based for readability)
- Human-readable error message
- Optional underlying cause (unwrappable)

```go
tag, err := tagparser.Parse(`foo='unterminated`)
if err != nil {
    var parseErr *tagparser.Error
    if errors.As(err, &parseErr) {
        fmt.Printf("Error at position %d: %s\n", parseErr.Pos+1, parseErr.Msg)
        // Output: Error at position 5: unterminated quote
    }
}
```

### Size Limits

Tags exceeding `MaxTagLength` (64KB) return `ErrTagTooLarge`:

```go
if errors.Is(err, tagparser.ErrTagTooLarge) {
    // Handle oversized input
}
```

## Performance

Benchmarks on MacBook Pro, 2025 (Go 1.25, ARM64):

```
BenchmarkParse_Simple          	 6851102	       170 ns/op	     376 B/op	       4 allocs/op
BenchmarkParseFunc_ZeroAlloc   	16681184	        67 ns/op	       0 B/op	       0 allocs/op
BenchmarkParse_Complex         	 4769185	       257 ns/op	     400 B/op	       5 allocs/op
```

**Key optimizations:**
- Fast path for simple tags (no quotes/escapes)
- Pre-allocated ASCII whitespace lookup table
- Zero-allocation callback mode
- Efficient string building with capacity hints


Run benchmarks: `go test -bench=. -benchmem`

## Testing

The library includes:
- **Unit tests**: 200+ test cases covering edge cases
- **Fuzz tests**: Native Go fuzzing for robustness
- **Benchmarks**: Performance tracking and regression detection

```bash
# Run tests
go test -v

# Run with race detector
go test -race

# Run fuzzing (Go 1.18+)
go test -fuzz=FuzzParse -fuzztime=30s

# Run benchmarks
go test -bench=. -benchmem
```

## API Stability

This library follows semantic versioning. The public API is stable:
- `Parse(tag string) (*Tag, error)`
- `ParseWithName(tag string) (*Tag, error)`
- `ParseFunc(tag string, callback func(key, value string) error) error`
- `ParseFuncWithName(tag string, callback func(key, value string) error) error`
- `type Tag struct { Name string; Options map[string]string }`
- `type Error struct { Tag string; Pos int; Msg string; Cause error }`

## Contributing

Contributions are welcome! Please:

1. **Open an issue** first to discuss major changes
2. **Add tests** for new functionality
3. **Run linters**: `golangci-lint run`
4. **Update docs** if changing public API
5. **Follow Go conventions** and existing code style

## License

See [LICENSE](LICENSE) file for details.

