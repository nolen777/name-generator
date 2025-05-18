package spaces_fetcher

import "testing"

func TestGetFile(t *testing.T) {
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
