package context

import "github.com/tinywasm/fmt"

// Context is a minimalist context compatible with TinyGo.
// No maps, no channels, uses a fixed array of 16 key-value pairs.
type Context struct {
	pairs [16]fmt.KeyValue
	count uint8
}

// errCapacityExceeded is returned when the context reaches its maximum capacity of 16 pairs.
var errCapacityExceeded = fmt.Err("context: max 16 values exceeded")

// Background returns an empty Context (equivalent to context.Background).
func Background() *Context {
	return &Context{}
}

// WithValue creates a new Context with the additional key-value pair.
// Returns errCapacityExceeded if the capacity of 16 pairs is exceeded.
func WithValue(parent *Context, key, value string) (*Context, error) {
	ctx := &Context{}
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

// Set adds or updates a key-value pair in-place.
// Returns an error if the capacity of 16 pairs is exceeded.
func (c *Context) Set(key, value string) error {
	if c.count >= 16 {
		return errCapacityExceeded
	}
	c.pairs[c.count] = fmt.KeyValue{Key: key, Value: value}
	c.count++
	return nil
}

// Value searches for the value associated with key (reverse search to prioritize latest values).
func (c *Context) Value(key string) string {
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
