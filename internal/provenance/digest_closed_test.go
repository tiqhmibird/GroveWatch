package provenance

import "testing"

func TestVerifyFailsClosedWithoutAComparableDigest(t *testing.T) {
	if Verify(&Snapshot{Digest: ""}) {
		t.Fatal("an empty snapshot must not verify")
	}
}
