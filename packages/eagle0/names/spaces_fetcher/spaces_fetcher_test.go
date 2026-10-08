package spaces_fetcher

import (
	"os"
	"testing"
)

func TestGetFile(t *testing.T) {
	if os.Getenv("DIGITALOCEAN_ACCESS_KEY_ID") == "" || os.Getenv("DIGITALOCEAN_SECRET_KEY") == "" {
		t.Skip("Spaces integration test requires DigitalOcean credentials")
	}
	// Test the GetFile function
	path := "names.tsv"
	data, err := GetFile(path)
	if err != nil {
		t.Fatalf("Failed to get file: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("Expected non-empty data, got empty")
	}
}
