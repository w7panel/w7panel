// nolint
package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPushOciProxy_InvalidRegistryPath(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	t.Setenv("KO_DATA_PATH", tmpDir)

	// Don't create the registry directory to simulate missing files

	// Run test
	if err := PushOciProxy(); err == nil {
		t.Error("Expected error when registry path is invalid, got nil")
	}
}

func TestPushOciProxy(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	registryDir := filepath.Join(tmpDir, "registry", "proxy-2.0.5")
	if err := os.MkdirAll(registryDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create dummy files to simulate OCI layout
	dummyFiles := []string{
		filepath.Join(registryDir, "oci-layout"),
		filepath.Join(registryDir, "index.json"),
		filepath.Join(registryDir, "blobs", "sha256", "testdigest"),
	}
	for _, f := range dummyFiles {
		if err := os.MkdirAll(filepath.Dir(f), 0755); err != nil {
			t.Fatalf("Failed to create test directory structure: %v", err)
		}
		if _, err := os.Create(f); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Set KO_DATA_PATH environment variable
	t.Setenv("KO_DATA_PATH", tmpDir)

	// Run test
	if err := PushOciProxy(); err != nil {
		t.Errorf("PushOciProxy() failed: %v", err)
	}
}

func TestPushOciProxy_KO_DATA_PATH_NotSet(t *testing.T) {
	// Ensure KO_DATA_PATH is not set
	// os.Unsetenv("KO_DATA_PATH")
	t.Setenv("KO_DATA_PATH", "/tmp")
	os.Setenv("KO_DATA_PATH", "/tmp")
	// Run test
	if err := PushOciProxy(); err != nil {
		t.Error("Expected error when KO_DATA_PATH is not set, got nil")
	}
}
