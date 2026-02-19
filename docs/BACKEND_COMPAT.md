# Implementation Plan: Backend Compatibility Refactor

## Context

The `tinywasm/context` package currently exposes only a custom WASM-optimized `*Context` struct
that is incompatible with the stdlib `context.Context` interface. Libraries like `tinywasm/agent`
that use SQLite and `net/http` require stdlib `context.Context` on the backend. This plan adds a
dual build-tag implementation so that:

- **Backend (`!wasm`):** `Context` IS `context.Context` (type alias) — zero overhead, full stdlib compatibility.
- **WASM (`wasm`):** `Context` is `*tinyCtx` (renamed from `*Context`) which satisfies the `context.Context` interface — minimalist, no channels, no maps.

This makes `tinywasm/context` a true isomorphic drop-in for stdlib `context`.

---

## Breaking Changes

| Change | Old | New |
|--------|-----|-----|
| Internal struct renamed | `*Context` (exported) | `*tinyCtx` (unexported); `type Context = *tinyCtx` |
| Method renamed | `Value(key string) string` | `Get(key string) string` |
| New method added | — | `Value(key any) any` (satisfies `context.Context`) |
| New functions | — | `TODO()`, `WithCancel()`, `WithTimeout()` |

Bumps minor version (v0.1.0 — breaking change within v0 is acceptable).

---

## File Restructure

```
context/
├── backStlib.go         # //go:build !wasm  — type alias + stdlib wrappers
├── frontWasm.go         # //go:build wasm   — tinyCtx struct + context.Context impl
├── ctx_shared_test.go   # (update: rename Value calls to Get)
├── ctx_stdlib_test.go   # (add tests for WithTimeout, WithCancel, WithValue)
├── ctx_wasm_test.go     # (add interface compliance test)
└── go.mod               # (unchanged)
```

**Delete:** `context.go` — its content is split between the two new files.

---

## Backend File: `backStlib.go` (!wasm)

```go
//go:build !wasm

package context

import (
	stdlib  "context"
	stdtime "time"
)

// Context is a type alias for stdlib context.Context on backend (non-WASM) builds.
// On WASM builds, Context is *tinyCtx which also satisfies context.Context.
// Using a type alias (=) means Context values are interchangeable with context.Context
// everywhere — net/http, database/sql, etc. accept them directly.
type Context = stdlib.Context

// CancelFunc is a type alias for stdlib context.CancelFunc on backend builds.
type CancelFunc = stdlib.CancelFunc

// Background returns a non-nil, empty Context (equivalent to context.Background).
func Background() Context { return stdlib.Background() }

// TODO returns a non-nil, empty Context (equivalent to context.TODO).
// Use when it is unclear which Context to use or it is not yet available.
func TODO() Context { return stdlib.TODO() }

// WithCancel returns a copy of parent with a new Done channel.
// The returned CancelFunc must be called to release resources.
func WithCancel(parent Context) (Context, CancelFunc) {
	return stdlib.WithCancel(parent)
}

// WithTimeout returns a copy of parent with a cancellation deadline set to
// ns nanoseconds from now. Pass multiples of time constants for clarity:
//
//	ctx, cancel := context.WithTimeout(parent, 30*twtime.Second)
//
// The returned CancelFunc must be called to release resources.
func WithTimeout(parent Context, ns int64) (Context, CancelFunc) {
	return stdlib.WithTimeout(parent, stdtime.Duration(ns))
}

// WithValue returns a copy of parent with the given string key-value pair stored.
// Only string keys are supported for cross-platform compatibility with tinyCtx.Get.
func WithValue(parent Context, key, value string) Context {
	return stdlib.WithValue(parent, key, value)
}
```

---

## WASM File: `frontWasm.go` (wasm)

```go
//go:build wasm

package context

import (
	"time"

	"github.com/tinywasm/fmt"
)

// tinyCtx is the minimalist WASM context.
// No channels, no maps — uses a fixed array of 16 key-value pairs.
// Satisfies the stdlib context.Context interface (Done returns nil, Err returns nil).
type tinyCtx struct {
	pairs [16]fmt.KeyValue
	count uint8
}

// Context is a type alias for *tinyCtx on WASM builds.
// On backend builds, Context is context.Context (stdlib type alias).
type Context = *tinyCtx

// CancelFunc is a no-op cancel function on WASM.
// No goroutines or channels are used for cancellation on this platform.
type CancelFunc = func()

var errCapacityExceeded = fmt.Err("context: max 16 values exceeded")

// --- context.Context interface compliance ---

// Deadline returns zero time and false. WASM contexts have no deadlines.
func (c *tinyCtx) Deadline() (time.Time, bool) { return time.Time{}, false }

// Done returns nil. Selecting on a nil channel blocks forever — equivalent to
// a non-cancellable context. WASM code should not rely on context cancellation.
func (c *tinyCtx) Done() <-chan struct{} { return nil }

// Err always returns nil (no cancellation on WASM).
func (c *tinyCtx) Err() error { return nil }

// Value returns the value associated with key if key is a string; otherwise nil.
// Satisfies the context.Context interface (accepts `any` key type).
func (c *tinyCtx) Value(key any) any {
	if k, ok := key.(string); ok {
		return c.get(k)
	}
	return nil
}

// --- tinywasm-specific string API ---

// Get retrieves the string value for key (reverse search, prioritizes latest entry).
// Renamed from the old Value(key string) string method for compatibility with
// context.Context.Value(key any) any.
func (c *tinyCtx) Get(key string) string {
	if c == nil {
		return ""
	}
	return c.get(key)
}

func (c *tinyCtx) get(key string) string {
	if c == nil {
		return ""
	}
	for i := int(c.count) - 1; i >= 0; i-- {
		if c.pairs[i].Key == key {
			return c.pairs[i].Value
		}
	}
	return ""
}

// Set adds a key-value pair in-place. Returns errCapacityExceeded if 16-pair limit reached.
func (c *tinyCtx) Set(key, value string) error {
	if c.count >= 16 {
		return errCapacityExceeded
	}
	c.pairs[c.count] = fmt.KeyValue{Key: key, Value: value}
	c.count++
	return nil
}

// Keys returns all keys in the context (including duplicate keys, in insertion order).
func (c *tinyCtx) Keys() []string {
	if c == nil || c.count == 0 {
		return nil
	}
	keys := make([]string, c.count)
	for i := uint8(0); i < c.count; i++ {
		keys[i] = c.pairs[i].Key
	}
	return keys
}

// --- Constructor functions ---

// Background returns an empty Context. Equivalent to context.Background().
func Background() Context { return &tinyCtx{} }

// TODO returns an empty Context. Equivalent to context.TODO().
func TODO() Context { return &tinyCtx{} }

// WithCancel returns the parent unchanged and a no-op cancel function.
// Cancellation is not supported on WASM.
func WithCancel(parent Context) (Context, CancelFunc) {
	return parent, func() {}
}

// WithTimeout returns the parent unchanged and a no-op cancel function.
// Timeouts are not enforced on WASM — ns is ignored.
func WithTimeout(parent Context, ns int64) (Context, CancelFunc) {
	return parent, func() {}
}

// WithValue returns a new context derived from parent with the key-value pair added.
// Returns errCapacityExceeded if the context already holds 16 pairs.
func WithValue(parent Context, key, value string) (Context, error) {
	ctx := &tinyCtx{}
	if parent != nil {
		ctx.pairs = parent.pairs
		ctx.count = parent.count
	}
	if ctx.count >= 16 {
		return nil, errCapacityExceeded
	}
	ctx.pairs[ctx.count] = fmt.KeyValue{Key: key, Value: value}
	ctx.count++
	return ctx, nil
}
```

> **API asymmetry note:** On WASM `WithValue` returns `(Context, error)` due to capacity
> constraints; on backend it returns only `Context`. Libraries that call `WithValue` should
> guard with a build tag or use `Set()` (mutable, WASM-only) instead. In practice, backend
> packages (SQLite, net/http) exclusively use `WithValue` so this asymmetry is acceptable.

---

## Test Updates

### `ctx_shared_test.go`
- Rename all `ctx.Value("key")` calls → `ctx.Get("key")` (WASM-only method set)
- Existing capacity, nil-receiver, and insertion order tests remain unchanged

### `ctx_stdlib_test.go` (add)
```go
func TestBackend_WithTimeout(t *testing.T) {
    ctx, cancel := WithTimeout(Background(), int64(50*time.Millisecond))
    defer cancel()
    select {
    case <-ctx.Done():
        // expected: timeout fired
    case <-time.After(200 * time.Millisecond):
        t.Error("expected context to be cancelled within 200ms")
    }
}

func TestBackend_WithCancel(t *testing.T) {
    ctx, cancel := WithCancel(Background())
    cancel()
    select {
    case <-ctx.Done():
        // expected
    default:
        t.Error("expected context to be done after cancel()")
    }
}

func TestBackend_InterfaceCompliance(t *testing.T) {
    var _ stdlibctx.Context = Background()
}
```

### `ctx_wasm_test.go` (add)
```go
func TestWASM_InterfaceCompliance(t *testing.T) {
    var _ stdlibctx.Context = Background()
}

func TestWASM_DoneIsNil(t *testing.T) {
    ctx := Background()
    if ctx.Done() != nil {
        t.Error("expected Done() to return nil on WASM")
    }
}

func TestWASM_GetReplaceValue(t *testing.T) {
    ctx := Background()
    _ = ctx.Set("k", "v")
    if ctx.Get("k") != "v" {
        t.Error("expected Get to return the set value")
    }
}
```

---

## Verification

```bash
cd /home/cesar/Dev/Pkg/tinywasm/context && gotest .
```

All tests must pass. Then publish:

```bash
cd /home/cesar/Dev/Pkg/tinywasm/context && gopush
```
