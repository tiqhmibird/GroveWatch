package provenance

import (
	"strings"
	"testing"
)

func TestReadRejectsAnOversizedSnapshot(t *testing.T) {
	big := strings.NewReader(strings.Repeat("a", MaxSnapshotBytes+1))
	_, err := Read(big)
	if err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("want an oversized error, got %v", err)
	}
}
