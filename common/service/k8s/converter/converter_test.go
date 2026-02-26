package converter

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewConverter(t *testing.T) {
	tests := []struct {
		name       string
		outputDir  string
		prefix     string
		standAlone bool
		expanded   bool
		kubernetes bool
		strict     bool
	}{
		{
			name:       "default configuration",
			outputDir:  "output",
			prefix:     "prefix",
			standAlone: false,
			expanded:   false,
			kubernetes: false,
			strict:     false,
		},
		{
			name:       "all flags true",
			outputDir:  "output",
			prefix:     "prefix",
			standAlone: true,
			expanded:   true,
			kubernetes: true,
			strict:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConverter(tt.outputDir, tt.prefix, tt.standAlone, tt.expanded, tt.kubernetes, tt.strict)

			if c.OutputDir != tt.outputDir {
				t.Errorf("expected OutputDir %s, got %s", tt.outputDir, c.OutputDir)
			}
			if c.Prefix != tt.prefix {
				t.Errorf("expected Prefix %s, got %s", tt.prefix, c.Prefix)
			}
			if c.StandAlone != tt.standAlone {
				t.Errorf("expected StandAlone %v, got %v", tt.standAlone, c.StandAlone)
			}
			if c.Expanded != tt.expanded {
				t.Errorf("expected Expanded %v, got %v", tt.expanded, c.Expanded)
			}
			if c.Kubernetes != tt.kubernetes {
				t.Errorf("expected Kubernetes %v, got %v", tt.kubernetes, c.Kubernetes)
			}
			if c.Strict != tt.strict {
				t.Errorf("expected Strict %v, got %v", tt.strict, c.Strict)
			}
		})
	}
}

func TestLoadSchema(t *testing.T) {
	// Setup test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"swagger": "2.0"}`))
	}))
	defer ts.Close()

	// Create temp file for file-based test
	tmpFile, err := os.CreateTemp("", "schema-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte(`{"swagger": "2.0"}`))
	tmpFile.Close()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid http url",
			url:     ts.URL,
			wantErr: false,
		},
		{
			name:    "valid file url",
			url:     "file://" + tmpFile.Name(),
			wantErr: false,
		},
		{
			name:    "invalid url",
			url:     "invalid://url",
			wantErr: true,
		},
		{
			name:    "non-existent file",
			url:     "file:///nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConverter("", "", false, false, false, false)
			_, err := c.loadSchema(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("loadSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriteSchemaFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "converter-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	c := NewConverter(tmpDir, "", false, false, false, false)

	tests := []struct {
		name    string
		content interface{}
		wantErr bool
	}{
		{
			name:    "valid json",
			content: map[string]string{"test": "value"},
			wantErr: false,
		},
		{
			name:    "invalid json",
			content: make(chan int), // channels can't be JSON marshaled
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.writeSchemaFile("test", tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("writeSchemaFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify file was created
				path := filepath.Join(tmpDir, "test.json")
				if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
					t.Errorf("expected file %s to exist", path)
				}
			}
		})
	}
}

func TestConvert(t *testing.T) {
	// Setup test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"swagger": "2.0", "definitions": {"Test": {"type": "object"}}}`))
	}))
	defer ts.Close()

	tmpDir, err := os.MkdirTemp("", "converter-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	c := NewConverter(tmpDir, "", false, false, false, false)

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid swagger schema",
			url:     ts.URL,
			wantErr: false,
		},
		{
			name:    "invalid schema",
			url:     "invalid://url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Convert(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewConverter2(t *testing.T) {
	tests := []struct {
		name       string
		outputDir  string
		prefix     string
		standAlone bool
		expanded   bool
		kubernetes bool
		strict     bool
	}{
		{
			name:       "default configuration",
			outputDir:  "output",
			prefix:     "prefix",
			standAlone: true,
			expanded:   true,
			kubernetes: true,
			strict:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConverter(tt.outputDir, tt.prefix, tt.standAlone, tt.expanded, tt.kubernetes, tt.strict)

			err := c.Convert("http://118.25.145.25:9090/assets/openapi.json")
			if err != nil {
				t.Errorf("NewConverter() error = %v", err)
			}
		})
	}
}
