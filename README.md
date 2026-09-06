# webtyp/context
<img src="docs/img/badges.svg">

Minimalist context library for TinyGo.

## Features

- **No Maps**: Uses a fixed array of structs to avoid dynamic allocations.
- **Mutable**: Supports in-place modification via `Set(key, value)`.
- **Minimal Footprint**: Designed to keep binary size at a minimum.
- **TinyGo Compatible**: Fully compatible with TinyGo and WASM environments.
- **Fixed Capacity**: Maximum of 16 key-value pairs (prioritizes latest values).

## Installation

```bash
go get webtyp.com/context
```

## Usage

```go
import "webtyp.com/context"

// Create background context
ctx := context.Background()

// Add values (Keys and Values must be strings)
// In-place mutation (simpler for state machines):
_ = ctx.Set("user_id", "123")

// Nested creation (traditional style):
ctx, _ = context.WithValue(ctx, "role", "admin")

// Retrieve values
user := ctx.Value("user_id") // returns "123"
role := ctx.Value("role")    // returns "admin"
missing := ctx.Value("none") // returns ""

// Inspect keys
keys := ctx.Keys()           // returns []string{"user_id", "role"}
```

## Constraints

- Keys and Values are restricted to `string` type for performance and simplicity.
- Maximum capacity of **16** key-value pairs. Exceeding this will return an error (see `fmt` error pattern).
