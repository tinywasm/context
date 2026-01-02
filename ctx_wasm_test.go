//go:build wasm

package context

import "testing"

func TestContext(t *testing.T) {
	RunContextTests(t)
}
