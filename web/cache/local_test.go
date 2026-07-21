package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

func TestLocalSet(t *testing.T) {
	c := NewLocal()

	// Set a value.
	const key = "foo"
	want := &mpb.Measurement{
		DeviceId: "id_foo",
	}
	if err := c.Set(context.Background(), key, want); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get it.
	got, err := c.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(got, want, cmpopts.IgnoreUnexported(mpb.Measurement{}, tspb.Timestamp{}, wpb.FloatValue{})); diff != "" {
		t.Fatalf("mismatch (-got +want):\n%s", diff)
	}
}

func TestLocalAdd(t *testing.T) {
	c := NewLocal()

	// Add a value.
	const key = "foo"
	want := &mpb.Measurement{
		DeviceId: "id_" + key,
	}
	if err := c.Add(context.Background(), key, want); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get it.
	got, err := c.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(got, want, cmpopts.IgnoreUnexported(mpb.Measurement{}, tspb.Timestamp{}, wpb.FloatValue{})); diff != "" {
		t.Fatalf("mismatch (-got +want):\n%s", diff)
	}
}

func TestLocalGetFailure(t *testing.T) {
	c := NewLocal()

	if _, err := c.Get(context.Background(), "foo"); err == nil {
		t.Fatalf("expected error, got nil")
	} else if !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("want ErrCacheMiss, got %v", err)
	}
}

func TestLocalAddFailure(t *testing.T) {
	c := NewLocal()

	// Add a value.
	const key = "foo"
	m := &mpb.Measurement{
		DeviceId: "id_" + key,
	}
	if err := c.Add(context.Background(), key, m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Attempt to add it again.
	if err := c.Add(context.Background(), key, m); err == nil {
		t.Fatalf("expected error, got nil")
	} else if !errors.Is(err, ErrNotStored) {
		t.Fatalf("want ErrNotStored, got %v", err)
	}
}
