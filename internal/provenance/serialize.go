package provenance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Marshal serializes a snapshot to indented, deterministic JSON. Field order is
// fixed by the struct definition and all collections are pre-sorted during the
// scan, so identical inputs produce byte-identical output (modulo CreatedAt).
func Marshal(snap *Snapshot) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}
	return buf.Bytes(), nil
}

// Unmarshal parses a snapshot from JSON, rejecting a document written by an
// incompatible version before the digest is ever recomputed.
func Unmarshal(data []byte) (*Snapshot, error) {
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	if snap.Schema != SchemaVersion {
		return nil, fmt.Errorf(
			"unmarshal snapshot: unsupported schema %q (this build writes %q)",
			snap.Schema, SchemaVersion)
	}
	return &snap, nil
}

// MaxSnapshotBytes bounds how much of a snapshot stream is read into memory. A
// snapshot is text about a workspace; anything past this is refused rather than
// loaded, because a verifier accepts snapshots from other people.
const MaxSnapshotBytes = 64 << 20

// Read parses a snapshot from an io.Reader, refusing input past MaxSnapshotBytes.
func Read(r io.Reader) (*Snapshot, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxSnapshotBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}
	if len(data) > MaxSnapshotBytes {
		return nil, fmt.Errorf(
			"read snapshot: larger than the %d byte limit - refusing to load it", MaxSnapshotBytes)
	}
	return Unmarshal(data)
}

// canonicalView is a projection of a snapshot that excludes volatile fields
// (CreatedAt and the Digest itself). Hashing this view yields the stable
// identity used as Snapshot.Digest.
type canonicalView struct {
	Schema      string        `json:"schema"`
	Tool        ToolInfo      `json:"tool"`
	Workspace   WorkspaceInfo `json:"workspace"`
	Files       []FileRecord  `json:"files"`
	Tools       []ToolRecord  `json:"tools"`
	Environment []EnvRecord   `json:"environment"`
}

// ComputeDigest returns the canonical SHA-256 identity of a snapshot. It is
// stable across runs with identical inputs because it excludes CreatedAt and
// the previously computed Digest.
func ComputeDigest(snap *Snapshot) (string, error) {
	view := canonicalView{
		Schema:      snap.Schema,
		Tool:        snap.Tool,
		Workspace:   snap.Workspace,
		Files:       snap.Files,
		Tools:       snap.Tools,
		Environment: snap.Environment,
	}
	// json.Marshal on a fixed struct with sorted slices is deterministic.
	data, err := json.Marshal(view)
	if err != nil {
		// A tamper check must fail closed: hashing an empty buffer here would
		// give every snapshot in this state the same digest, and any snapshot
		// carrying that digest would verify.
		return "", fmt.Errorf("canonical view: %w", err)
	}
	return hashBytes(data), nil
}

// Verify recomputes the digest of snap and reports whether it matches the
// stored Digest. A false result means the snapshot was tampered with or
// produced by an incompatible version; a canonicalisation failure is also a
// false result, because the check cannot be completed.
func Verify(snap *Snapshot) bool {
	digest, err := ComputeDigest(snap)
	return err == nil && digest == snap.Digest
}
