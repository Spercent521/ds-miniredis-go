package db

import (
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	d := New()
	d.SetString("k", "v", 0)
	got, ok := d.GetString("k")
	if !ok || got != "v" {
		t.Fatalf("expected v, got %q ok=%v", got, ok)
	}
}

func TestGetMissingKey(t *testing.T) {
	d := New()
	_, ok := d.GetString("no-such-key")
	if ok {
		t.Fatal("expected miss for absent key")
	}
}

func TestDel(t *testing.T) {
	d := New()
	d.SetString("a", "1", 0)
	d.SetString("b", "2", 0)
	n := d.Del("a", "b", "c")
	if n != 2 {
		t.Fatalf("expected 2 deletions, got %d", n)
	}
	_, ok := d.GetString("a")
	if ok {
		t.Fatal("key 'a' should have been deleted")
	}
}

func TestExpiry(t *testing.T) {
	d := New()
	// Set with 50ms TTL
	expireAt := time.Now().Add(50 * time.Millisecond).UnixMilli()
	d.SetString("tmp", "hello", expireAt)

	got, ok := d.GetString("tmp")
	if !ok || got != "hello" {
		t.Fatalf("expected key before expiry, got %q ok=%v", got, ok)
	}

	time.Sleep(80 * time.Millisecond)

	_, ok = d.GetString("tmp")
	if ok {
		t.Fatal("expected key to be expired after TTL")
	}
}

func TestSetOverwrite(t *testing.T) {
	d := New()
	d.SetString("x", "first", 0)
	d.SetString("x", "second", 0)
	got, ok := d.GetString("x")
	if !ok || got != "second" {
		t.Fatalf("expected 'second', got %q", got)
	}
}
