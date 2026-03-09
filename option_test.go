package web

import "testing"

func TestWithAddress(t *testing.T) {
	app := defaultApp()
	opt := WithAddress(":1234")
	opt(app)
	if app.addr != ":1234" {
		t.Fatalf("expected addr :1234, got %v", app.addr)
	}
}
