package cache

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

func TestSet(t *testing.T) {
	c := NewLocal()

	// Set a value.
	m1 := mpb.Measurement{
		DeviceId: "id_foo",
	}
	if err := c.Set(context.Background(), "foo", &m1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get it.
	var m2 mpb.Measurement
	if err := c.Get(context.Background(), "foo", &m2); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Compare.
	if diff := cmp.Diff(m2, m1, cmpopts.IgnoreUnexported(mpb.Measurement{}, tspb.Timestamp{}, wpb.FloatValue{})); diff != "" {
		t.Errorf("Unexpected result (-got +want):\n%s", diff)
	}
}

func TestAdd(t *testing.T) {
	c := NewLocal()

	// Add a value.
	m1 := mpb.Measurement{
		DeviceId: "id_foo",
	}
	if err := c.Add(context.Background(), "foo", &m1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get it.
	var m2 mpb.Measurement
	if err := c.Get(context.Background(), "foo", &m2); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Compare.
	if diff := cmp.Diff(m2, m1, cmpopts.IgnoreUnexported(mpb.Measurement{}, tspb.Timestamp{}, wpb.FloatValue{})); diff != "" {
		t.Errorf("Unexpected result (-got +want):\n%s", diff)
	}
}

func TestGetFailure(t *testing.T) {
	c := NewLocal()

	var m mpb.Measurement
	if err := c.Get(context.Background(), "foo", &m); err == nil {
		t.Errorf("expected error, got nil")
	} else if err != ErrCacheMiss {
		t.Errorf("want ErrCacheMiss, got %v", err)
	}
}

func TestAddFailure(t *testing.T) {
	c := NewLocal()

	// Add a value.
	m1 := mpb.Measurement{
		DeviceId: "id_foo",
	}
	if err := c.Add(context.Background(), "foo", &m1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Attempt to add it again.
	if err := c.Add(context.Background(), "foo", &m1); err == nil {
		t.Errorf("expected error, got nil")
	} else if err != ErrNotStored {
		t.Errorf("want ErrNotStored, got %v", err)
	}
}
