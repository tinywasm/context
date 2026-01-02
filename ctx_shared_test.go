package context

import "testing"

// RunContextTests runs all context tests. Called from environment-specific test files.
func RunContextTests(t *testing.T) {
	t.Run("Background", func(t *testing.T) {
		ctx := Background()
		if ctx == nil {
			t.Fatal("Background() returned nil")
		}
		if ctx.count != 0 {
			t.Errorf("expected count 0, got %d", ctx.count)
		}
	})

	t.Run("TODO", func(t *testing.T) {
		ctx := TODO()
		if ctx == nil {
			t.Fatal("TODO() returned nil")
		}
	})

	t.Run("NilReceiver", func(t *testing.T) {
		var ctx *Context
		if v := ctx.Value("key"); v != "" {
			t.Errorf("expected empty string for nil context, got '%s'", v)
		}
	})

	t.Run("WithValueAndGet", func(t *testing.T) {
		ctx := Background()
		ctx = WithValue(ctx, "user", "alice")
		ctx = WithValue(ctx, "role", "admin")

		if v := ctx.Value("user"); v != "alice" {
			t.Errorf("expected 'alice', got '%s'", v)
		}
		if v := ctx.Value("role"); v != "admin" {
			t.Errorf("expected 'admin', got '%s'", v)
		}
		if v := ctx.Value("missing"); v != "" {
			t.Errorf("expected '', got '%s'", v)
		}
	})

	t.Run("OverwriteValue", func(t *testing.T) {
		ctx := WithValue(Background(), "key", "v1")
		ctx = WithValue(ctx, "key", "v2")
		if v := ctx.Value("key"); v != "v2" {
			t.Errorf("expected 'v2', got '%s'", v)
		}
	})

	t.Run("MaxCapacity", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic on exceeding capacity")
			}
		}()
		ctx := Background()
		for i := 0; i < 9; i++ {
			ctx = WithValue(ctx, "k", "v")
		}
	})
}
