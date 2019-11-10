package mmap

import (
	"bytes"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// testFilePath is the path to the test file.
var testFilePath = filepath.Join(os.TempDir(), "mmap.test")

// testFileMode is the access mode of the test file.
var testFileMode = os.FileMode(0600)

// testFileLength is the length of the test file.
var testFileLength = uintptr(1 << 20)

// testBuffer is the non-zero test data.
var testBuffer = []byte{'H', 'E', 'L', 'L', 'O'}

// testBufferLength is the length of testBuffer.
var testBufferLength = len(testBuffer)

// zeroBuffer is the zero test data of the same length as testBuffer.
var zeroBuffer = make([]byte, testBufferLength)

// deleteTestFile deletes the test file.
func deleteTestFile() error {
	if _, err := os.Stat(testFilePath); err == nil || !os.IsNotExist(err) {
		return os.Remove(testFilePath)
	}
	return nil
}

// openTestFile opens and returns the test file.
// recreate argument specifies whether needed to recreate the existing file.
func openTestFile(t *testing.T, recreate bool) (*os.File, error) {
	if recreate {
		if err := deleteTestFile(); err != nil {
			return nil, err
		}
	}
	f, err := os.OpenFile(testFilePath, os.O_CREATE|os.O_RDWR, testFileMode)
	if err != nil {
		return nil, err
	}
	if err := f.Truncate(int64(testFileLength)); err != nil {
		closeTestEntity(t, f)
		return nil, err
	}
	return f, nil
}

// openTestMapping opens and returns a new mapping of the test file into the memory.
func openTestMapping(t *testing.T, mode Mode) (*Mapping, error) {
	f, err := openTestFile(t, true)
	if err != nil {
		return nil, err
	}
	defer closeTestEntity(t, f)
	return Open(f.Fd(), 0, testFileLength, mode, 0)
}

// closeTestEntity closes the given entity.
// It ignores ErrClosed error by the reason this error returns when mapping is closed twice.
func closeTestEntity(t *testing.T, closer io.Closer) {
	if err := closer.Close(); err != nil {
		if err != ErrClosed {
			t.Fatal(err)
		}
	}
}

//------------------------------------------- TEST CASES ---------------------------------------------------------------

// TestWithOpenedFile tests the work with the mapping of file which is not closed before closing mapping.
// CASE: The mapping MUST works correctly.
func TestWithOpenedFile(t *testing.T) {
	f, err := openTestFile(t, true)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, f)
	m, err := Open(f.Fd(), 0, testFileLength, ModeReadWrite, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testBuffer, 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, testBufferLength)
	if _, err := m.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testBuffer) != 0 {
		t.Fatalf("data must be %q, %v found", testBuffer, buf)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestWithClosedFile tests the work with the mapping of file which is closed before closing mapping.
// CASE: The duplication of the file descriptor MUST works correctly.
func TestWithClosedFile(t *testing.T) {
	m, err := openTestMapping(t, ModeReadWrite)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testBuffer, 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, testBufferLength)
	if _, err := m.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testBuffer) != 0 {
		t.Fatalf("data must be %q, %v found", testBuffer, buf)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestUnalignedOffset tests using the unaligned start address of the mapping memory.
// CASE: The unaligned offset MUST works correctly.
// TODO: This is a strange test...
func TestUnalignedOffset(t *testing.T) {
	f, err := openTestFile(t, true)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, f)
	offLen := uintptr(testBufferLength - 1)
	m, err := Open(f.Fd(), 1, offLen, ModeReadWrite, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	offBuf := make([]byte, offLen)
	copy(offBuf, testBuffer[1:])
	if _, err := m.WriteAt(offBuf, 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, offLen)
	if _, err := m.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, offBuf) != 0 {
		t.Fatalf("data must be %q, %v found", offBuf, buf)
	}
}

// TestSharedSync tests the synchronization of the mapped memory with the underlying file in the shared mode.
// CASE: The data which is read directly from the underlying file MUST be exactly the same
// as the previously written through the mapped memory.
func TestSharedSync(t *testing.T) {
	m, err := openTestMapping(t, ModeReadWrite)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testBuffer, 0); err != nil {
		t.Fatal(err)
	}
	if err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	f, err := openTestFile(t, false)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, f)
	buf := make([]byte, testBufferLength)
	if _, err := f.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testBuffer) != 0 {
		t.Fatalf("data must be %q, %v found", testBuffer, buf)
	}
}

// TestPrivateSync tests the synchronization of the mapped memory with the underlying file in the private mode.
// CASE: The data which is read directly from the underlying file MUST NOT be affected
// by the previous write through the mapped memory.
func TestPrivateSync(t *testing.T) {
	m, err := openTestMapping(t, ModeWriteCopy)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testBuffer, 0); err != nil {
		t.Fatal(err)
	}
	if err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	f, err := openTestFile(t, false)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, f)
	buf := make([]byte, testBufferLength)
	if _, err := f.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, zeroBuffer) != 0 {
		t.Fatalf("data must be %q, %v found", zeroBuffer, buf)
	}
}

// TestPartialRead tests the reading beyond the mapped memory.
// CASE 1: The ErrOutOfBounds MUST be returned.
// CASE 2: The reading buffer MUST NOT be modified.
func TestPartialRead(t *testing.T) {
	f, err := openTestFile(t, true)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, f)
	partLen := uintptr(testBufferLength - 1)
	m, err := Open(f.Fd(), 0, partLen, ModeReadWrite, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testBuffer[:partLen], 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, testBufferLength)
	if _, err := m.ReadAt(buf, 0); err == nil {
		t.Fatal("expected ErrOutOfBounds, no error found")
	} else if err != ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, [%v] error found", err)
	}
	if bytes.Compare(buf, zeroBuffer) != 0 {
		t.Fatalf("data must be %q, %v found", zeroBuffer, buf)
	}
}

// TestPartialWrite tests the writing beyond the mapped memory.
// CASE 1: The ErrOutOfBounds MUST be returned.
// CASE 2: The mapped memory MUST NOT be modified.
func TestPartialWrite(t *testing.T) {
	f, err := openTestFile(t, true)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, f)
	partLen := uintptr(testBufferLength - 1)
	m, err := Open(f.Fd(), 0, partLen, ModeReadWrite, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testBuffer, 0); err == nil {
		t.Fatal("expected ErrOutOfBounds, no error found")
	} else if err != ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, [%v] error found", err)
	}
	partBuf := make([]byte, partLen)
	if _, err := m.ReadAt(partBuf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(partBuf, zeroBuffer[:partLen]) != 0 {
		t.Fatalf("data must be %q, %v found", zeroBuffer, partBuf)
	}
}

// TestFileOpening tests the OpenFile function.
// CASE 1: The initializer must be called once.
// CASE 2: The data read on the second opening must be exactly the same as previously written on the first one.
func TestFileOpening(t *testing.T) {
	if err := deleteTestFile(); err != nil {
		t.Fatal(err)
	}
	initCallCount := 0
	open := func() (*Mapping, error) {
		return OpenFile(testFilePath, testFileMode, uintptr(testBufferLength), 0, func(m *Mapping) error {
			initCallCount++
			_, err := m.WriteAt(testBuffer, 0)
			return err
		})
	}
	m, err := open()
	if err != nil {
		t.Fatal(err)
	}
	closeTestEntity(t, m)
	m, err = open()
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	if initCallCount > 1 {
		t.Fatalf("initializer must be called once, %d calls found", initCallCount)
	}
	buf := make([]byte, testBufferLength)
	if _, err := m.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testBuffer) != 0 {
		t.Fatalf("data must be %q, %v found", testBuffer, buf)
	}
}

// TestSegment tests the data segment.
// CASE: The read data must be exactly the same as the previously written unsigned 32-bit integer.
func TestSegment(t *testing.T) {
	m, err := openTestMapping(t, ModeReadWrite)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	*m.Segment().Uint32(0) = math.MaxUint32 - 1
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	f, err := openTestFile(t, false)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, f)
	buf := make([]byte, 4)
	if _, err := f.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, []byte{254, 255, 255, 255}) != 0 {
		t.Fatalf("data must be %q, %v found", testBuffer, buf)
	}
}
