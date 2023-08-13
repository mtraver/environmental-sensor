package device

import (
	"fmt"
	"testing"
)

func TestGetRegistryPath(t *testing.T) {
	expected := fmt.Sprintf("projects/exampleProj/locations/%s/registries/exampleReg", gcpRegion)
	path := getRegistryPath("exampleProj", "exampleReg")

	if path != expected {
		t.Errorf("Expected %q, got %q", expected, path)
	}
}
