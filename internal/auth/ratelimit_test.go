package auth

import (
	"testing"
	"time"
)

func TestLimiterAllow(t *testing.T) {
	limiter := NewLimiter(3, time.Hour)

	for i := 1; i <= 3; i++ {
		if !limiter.Allow("1.2.3.4") {
			t.Fatalf("attempt %d blocked, want allowed", i)
		}
	}

	if limiter.Allow("1.2.3.4") {
		t.Error("attempt 4 allowed, want blocked")
	}

	if !limiter.Allow("5.6.7.8") {
		t.Error("another client blocked, want allowed")
	}

	limiter.Reset("1.2.3.4")

	if !limiter.Allow("1.2.3.4") {
		t.Error("blocked after reset, want allowed")
	}
}

func TestLimiterWindowExpires(t *testing.T) {
	limiter := NewLimiter(1, time.Nanosecond)

	limiter.Allow("1.2.3.4")
	time.Sleep(time.Millisecond)

	if !limiter.Allow("1.2.3.4") {
		t.Error("blocked after the window passed, want allowed")
	}
}
