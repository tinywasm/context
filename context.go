package context

import "errors"

// Context is a minimalist context compatible with TinyGo.
// No maps, no channels, uses a fixed array of 16 key-value pairs.
type Context struct {
	pairs [16]pair
	count uint8
}

type pair struct {
	key   string
	value string
}

// ErrCapacityExceeded is returned when the context reaches its maximum capacity of 16 pairs.
var ErrCapacityExceeded = errors.New("context: max 16 values exceeded")

// Background returns an empty Context (equivalent to context.Background).
func Background() *Context {
	return &Context{}
}

// TODO returns an empty Context (equivalent to context.TODO).
func TODO() *Context {
	return Background()
}

// WithValue creates a new Context with the additional key-value pair.
// Returns ErrCapacityExceeded if the capacity of 16 pairs is exceeded.
func WithValue(parent *Context, key, value string) (*Context, error) {
	ctx := &Context{}
	if parent != nil {
		ctx.pairs = parent.pairs
		ctx.count = parent.count
	}
	if ctx.count >= 16 {
		return nil, ErrCapacityExceeded
	}
	ctx.pairs[ctx.count] = pair{key: key, value: value}
	ctx.count++
	return ctx, nil
}

// Value searches for the value associated with key (reverse search to prioritize latest values).
func (c *Context) Value(key string) string {
	if c == nil {
		return ""
	}
	for i := int(c.count) - 1; i >= 0; i-- {
		if c.pairs[i].key == key {
			return c.pairs[i].value
		}
	}
	return ""
}
