package cache

import (
	"testing"
	"time"
)

func TestStoreRoundTrip(t *testing.T) {
	s := New(t.TempDir())
	now := time.Now()
	if err := s.Write("payload.json", []byte(`{"ok":true}`), now); err != nil {
		t.Fatalf("write: %v", err)
	}
	data, fetchedAt, ok := s.Read("payload.json")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if string(data) != `{"ok":true}` {
		t.Errorf("data = %q", data)
	}
	if fetchedAt.Unix() != now.Unix() {
		t.Errorf("fetchedAt = %v, want ~%v", fetchedAt, now)
	}
	age, ok := s.Age("payload.json", now.Add(90*time.Second))
	if !ok || age < 80*time.Second || age > 100*time.Second {
		t.Errorf("age = %v ok=%v, want ~90s", age, ok)
	}
}

func TestStoreMissing(t *testing.T) {
	s := New(t.TempDir())
	if _, _, ok := s.Read("nope.json"); ok {
		t.Error("expected missing entry")
	}
	if _, ok := s.Age("nope.json", time.Now()); ok {
		t.Error("expected missing age")
	}
}
