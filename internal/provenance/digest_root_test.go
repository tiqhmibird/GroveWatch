package provenance

import "testing"

func TestMerkleRootCoversSizeAndMode(t *testing.T) {
	base := []FileRecord{{Path: "a.sh", Size: 3, Mode: "0644", Digest: "abc"}}
	root := merkleRoot(base)
	if got := merkleRoot([]FileRecord{{Path: "a.sh", Size: 3, Mode: "0755", Digest: "abc"}}); got == root {
		t.Fatal("a mode change must move the tree hash")
	}
	if got := merkleRoot([]FileRecord{{Path: "a.sh", Size: 4, Mode: "0644", Digest: "abc"}}); got == root {
		t.Fatal("a size change must move the tree hash")
	}
}
