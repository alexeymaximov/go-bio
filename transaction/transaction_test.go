package transaction

import (
	"bytes"
	"testing"
)

// testBuffer is the non-zero test data.
var testBuffer = []byte{'H', 'E', 'L', 'L', 'O'}

// testBufferLength is the length of testBuffer.
var testBufferLength = len(testBuffer)

// zeroBuffer is the zero test data of the same length as testBuffer.
var zeroBuffer = make([]byte, testBufferLength)

//------------------------------------------- TEST CASES ---------------------------------------------------------------

// TestSnapshot tests the snapshot.
// CASE: The data read from the snapshot MUST NOT be affected by the previous modification of the original data.
func TestSnapshot(t *testing.T) {
	buf := make([]byte, testBufferLength)
	tx, err := Begin(buf, 0, uintptr(testBufferLength))
	if err != nil {
		t.Fatal(err)
	}
	copy(buf, testBuffer)
	snapshot := make([]byte, testBufferLength)
	if _, err := tx.ReadAt(snapshot, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(snapshot, zeroBuffer) != 0 {
		t.Fatalf("snapshot must be %q, %v found", zeroBuffer, buf)
	}
}

// TestRollback tests the transaction rollback.
// CASE: The original data MUST NOT be affected by the previous write through the transaction.
func TestRollback(t *testing.T) {
	buf := make([]byte, testBufferLength)
	tx, err := Begin(buf, 0, uintptr(testBufferLength))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.WriteAt(testBuffer, 0); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, zeroBuffer) != 0 {
		t.Fatalf("original must be %q, %v found", zeroBuffer, buf)
	}
}

// TestCommit tests the transaction commit.
// CASE: The original data MUST be exactly the same as the previously written through the transaction.
func TestCommit(t *testing.T) {
	buf := make([]byte, testBufferLength)
	tx, err := Begin(buf, 0, uintptr(testBufferLength))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.WriteAt(testBuffer, 0); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testBuffer) != 0 {
		t.Fatalf("original must be %q, %v found", testBuffer, buf)
	}
}
