package web

import "testing"

func TestWithAddress(t *testing.T) {
	app := defaultApp()
	opt := WithAddress(":1234")
	opt(app)
	if app.Address() != ":1234" {
		t.Fatalf("expected addr :1234, got %v", app.Address())
	}
}

func TestWithMaxResponseBodyBytes(t *testing.T) {
	app := defaultApp()
	opt := WithMaxResponseBodyBytes(1024)
	opt(app)
	if app.MaxResponseBodyBytes() != 1024 {
		t.Fatalf("expected max body bytes 1024, got %d", app.MaxResponseBodyBytes())
	}
}
