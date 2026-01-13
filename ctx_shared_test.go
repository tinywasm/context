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

	t.Run("NilReceiver", func(t *testing.T) {
		var ctx *Context
		if v := ctx.Value("key"); v != "" {
			t.Errorf("expected empty string for nil context, got '%s'", v)
		}
	})

	t.Run("WithValueAndGet", func(t *testing.T) {
		ctx := Background()
		var err error
		ctx, err = WithValue(ctx, "user", "alice")
		if err != nil {
			t.Fatal(err)
		}
		ctx, err = WithValue(ctx, "role", "admin")
		if err != nil {
			t.Fatal(err)
		}

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
		ctx, err := WithValue(Background(), "key", "v1")
		if err != nil {
			t.Fatal(err)
		}
		ctx, err = WithValue(ctx, "key", "v2")
		if err != nil {
			t.Fatal(err)
		}
		if v := ctx.Value("key"); v != "v2" {
			t.Errorf("expected 'v2', got '%s'", v)
		}
	})

	t.Run("MaxCapacity", func(t *testing.T) {
		ctx := Background()
		var err error
		for i := 0; i < 16; i++ {
			ctx, err = WithValue(ctx, "k", "v")
			if err != nil {
				t.Fatalf("unexpected error at index %d: %v", i, err)
			}
		}
		// 17th should fail
		_, err = WithValue(ctx, "k", "v")
		if err != errCapacityExceeded {
			t.Errorf("expected errCapacityExceeded, got %v", err)
		}
	})

	t.Run("MutableSet", func(t *testing.T) {
		ctx := Background()
		if err := ctx.Set("name", "bob"); err != nil {
			t.Fatal(err)
		}
		if v := ctx.Value("name"); v != "bob" {
			t.Errorf("expected 'bob', got '%s'", v)
		}

		// Verify it mutates the same pointer
		if err := ctx.Set("job", "dev"); err != nil {
			t.Fatal(err)
		}
		if v := ctx.Value("job"); v != "dev" {
			t.Errorf("expected 'dev', got '%s'", v)
		}
		if v := ctx.Value("name"); v != "bob" {
			t.Errorf("expected 'bob' still there, got '%s'", v)
		}

		// Test capacity with Set
		ctx = Background()
		for i := 0; i < 16; i++ {
			if err := ctx.Set("k", "v"); err != nil {
				t.Fatalf("unexpected error at index %d: %v", i, err)
			}
		}
		if err := ctx.Set("k", "v"); err != errCapacityExceeded {
			t.Errorf("expected errCapacityExceeded, got %v", err)
		}
	})
}
