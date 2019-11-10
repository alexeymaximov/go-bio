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
	data := make([]byte, testBufferLength)
	tx, err := Begin(data, 0, uintptr(testBufferLength))
	if err != nil {
		t.Fatal(err)
	}
	copy(data, testBuffer)
	snapshot := make([]byte, testBufferLength)
	if _, err := tx.ReadAt(snapshot, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(snapshot, zeroBuffer) != 0 {
		t.Fatalf("snapshot must be %q, %v found", zeroBuffer, data)
	}
}

// TestRollback tests the transaction rollback.
// CASE: The original data MUST NOT be affected by the previous write through the transaction.
func TestRollback(t *testing.T) {
	data := make([]byte, testBufferLength)
	tx, err := Begin(data, 0, uintptr(testBufferLength))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.WriteAt(testBuffer, 0); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(data, zeroBuffer) != 0 {
		t.Fatalf("original must be %q, %v found", zeroBuffer, data)
	}
}

// TestCommit tests the transaction commit.
// CASE: The original data MUST be exactly the same as the previously written through the transaction.
func TestCommit(t *testing.T) {
	data := make([]byte, testBufferLength)
	tx, err := Begin(data, 0, uintptr(testBufferLength))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.WriteAt(testBuffer, 0); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(data, testBuffer) != 0 {
		t.Fatalf("original must be %q, %v found", testBuffer, data)
	}
}

// TestPartialRead tests the reading beyond the transaction data.
// CASE 1: The ErrOutOfBounds MUST be returned.
// CASE 2: The reading buffer MUST NOT be modified.
func TestPartialRead(t *testing.T) {
	partLen := uintptr(testBufferLength - 1)
	data := make([]byte, partLen)
	tx, err := Begin(data, 0, partLen)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.WriteAt(testBuffer[:partLen], 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, testBufferLength)
	if _, err := tx.ReadAt(buf, 0); err == nil {
		t.Fatal("expected ErrOutOfBounds, no error found")
	} else if err != ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, [%v] error found", err)
	}
	if bytes.Compare(buf, zeroBuffer) != 0 {
		t.Fatalf("data must be %q, %v found", zeroBuffer, buf)
	}
}

// TestPartialWrite tests the writing beyond the transaction data.
// CASE 1: The ErrOutOfBounds MUST be returned.
// CASE 2: The transaction data MUST NOT be modified.
func TestPartialWrite(t *testing.T) {
	partLen := uintptr(testBufferLength - 1)
	data := make([]byte, partLen)
	tx, err := Begin(data, 0, partLen)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.WriteAt(testBuffer, 0); err == nil {
		t.Fatal("expected ErrOutOfBounds, no error found")
	} else if err != ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, [%v] error found", err)
	}
	partBuf := make([]byte, partLen)
	if _, err := tx.ReadAt(partBuf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(partBuf, zeroBuffer[:partLen]) != 0 {
		t.Fatalf("data must be %q, %v found", zeroBuffer, partBuf)
	}
}
