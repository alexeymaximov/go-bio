package mmap

import (
	"bytes"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// testFilePath is the template of the path to the test file.
var testFilePath = filepath.Join(os.TempDir(), "github.com+alexeymaximov+go-bio+mmap")

// testFileIndex is the current index of the test file.
var testFileIndex uint64 = 0

// testFileMode is the access mode of the test file.
var testFileMode = os.FileMode(0600)

// testData is the non-zero test data.
var testData = []byte{'H', 'E', 'L', 'L', 'O'}

// testDataLength is the length of test data.
var testDataLength = len(testData)

// testZeroData is the zero test data of the same length as test data.
var testZeroData = make([]byte, testDataLength)

// nextTestFilePath returns the path to the test file.
func nextTestFilePath(t *testing.T) string {
	testFileIndex++
	filePath := testFilePath + "_" + strconv.FormatUint(testFileIndex, 10)
	if _, err := os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	} else {
		if err := os.Remove(filePath); err != nil {
			t.Fatal(err)
		}
	}
	return filePath
}

// openNextTestFile opens and returns the test file.
// if copyPrevious == true previous test file will be copied if exists.
func openNextTestFile(t *testing.T, copyPrevious bool) *os.File {
	filePath := nextTestFilePath(t)
	fileData := testZeroData
	if copyPrevious && testFileIndex > 1 {
		var err error
		fileData, err = ioutil.ReadFile(testFilePath + "_" + strconv.FormatUint(testFileIndex-1, 10))
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := ioutil.WriteFile(filePath, fileData, testFileMode); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(filePath, os.O_RDWR, testFileMode)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// openTestMapping opens and returns a new mapping of the test file into the memory.
func openTestMapping(t *testing.T, mode Mode) *Mapping {
	f := openNextTestFile(t, false)
	defer closeTestEntity(t, f)
	m, err := Open(f.Fd(), 0, uintptr(testDataLength), mode, 0)
	if err != nil {
		t.Fatal(err)
	}
	return m
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
	f := openNextTestFile(t, true)
	defer closeTestEntity(t, f)
	m, err := Open(f.Fd(), 0, uintptr(testDataLength), ModeReadWrite, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testData, 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, testDataLength)
	if _, err := m.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testData) != 0 {
		t.Fatalf("data must be %q, %v found", testData, buf)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestWithClosedFile tests the work with the mapping of file which is closed before closing mapping.
// CASE: The duplication of the file descriptor MUST works correctly.
func TestWithClosedFile(t *testing.T) {
	m := openTestMapping(t, ModeReadWrite)
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testData, 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, testDataLength)
	if _, err := m.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testData) != 0 {
		t.Fatalf("data must be %q, %v found", testData, buf)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestUnalignedOffset tests using the unaligned start address of the mapping memory.
// CASE: The unaligned offset MUST works correctly.
// TODO: This is a strange test...
func TestUnalignedOffset(t *testing.T) {
	f := openNextTestFile(t, false)
	defer closeTestEntity(t, f)
	partialLength := uintptr(testDataLength - 1)
	m, err := Open(f.Fd(), 1, partialLength, ModeReadWrite, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	partialData := make([]byte, partialLength)
	copy(partialData, testData[1:])
	if _, err := m.WriteAt(partialData, 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, partialLength)
	if _, err := m.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, partialData) != 0 {
		t.Fatalf("data must be %q, %v found", partialData, buf)
	}
}

// TestSharedSync tests the synchronization of the mapped memory with the underlying file in the shared mode.
// CASE: The data which is read directly from the underlying file MUST be exactly the same
// as the previously written through the mapped memory.
func TestSharedSync(t *testing.T) {
	m := openTestMapping(t, ModeReadWrite)
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testData, 0); err != nil {
		t.Fatal(err)
	}
	if err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	f := openNextTestFile(t, true)
	defer closeTestEntity(t, f)
	buf := make([]byte, testDataLength)
	if _, err := f.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testData) != 0 {
		t.Fatalf("data must be %q, %v found", testData, buf)
	}
}

// TestPrivateSync tests the synchronization of the mapped memory with the underlying file in the private mode.
// CASE: The data which is read directly from the underlying file MUST NOT be affected
// by the previous write through the mapped memory.
func TestPrivateSync(t *testing.T) {
	m := openTestMapping(t, ModeWriteCopy)
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testData, 0); err != nil {
		t.Fatal(err)
	}
	if err := m.Sync(); err != nil {
		t.Fatal(err)
	}
	f := openNextTestFile(t, true)
	defer closeTestEntity(t, f)
	buf := make([]byte, testDataLength)
	if _, err := f.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testZeroData) != 0 {
		t.Fatalf("data must be %v, %v found", testZeroData, buf)
	}
}

// TestPartialRead tests the reading beyond the mapped memory.
// CASE 1: The ErrOutOfBounds MUST be returned.
// CASE 2: The reading buffer MUST NOT be modified.
func TestPartialRead(t *testing.T) {
	f := openNextTestFile(t, false)
	defer closeTestEntity(t, f)
	partialLength := uintptr(testDataLength - 1)
	m, err := Open(f.Fd(), 0, partialLength, ModeReadWrite, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testData[:partialLength], 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, testDataLength)
	if _, err := m.ReadAt(buf, 0); err == nil {
		t.Fatal("expected ErrOutOfBounds, no error found")
	} else if err != ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, [%v] error found", err)
	}
	if bytes.Compare(buf, testZeroData) != 0 {
		t.Fatalf("data must be %v, %v found", testZeroData, buf)
	}
}

// TestPartialWrite tests the writing beyond the mapped memory.
// CASE 1: The ErrOutOfBounds MUST be returned.
// CASE 2: The mapped memory MUST NOT be modified.
func TestPartialWrite(t *testing.T) {
	f := openNextTestFile(t, false)
	defer closeTestEntity(t, f)
	partialLength := uintptr(testDataLength - 1)
	m, err := Open(f.Fd(), 0, partialLength, ModeReadWrite, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestEntity(t, m)
	if _, err := m.WriteAt(testData, 0); err == nil {
		t.Fatal("expected ErrOutOfBounds, no error found")
	} else if err != ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, [%v] error found", err)
	}
	buf := make([]byte, partialLength)
	if _, err := m.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testZeroData[:partialLength]) != 0 {
		t.Fatalf("data must be %v, %v found", testZeroData, buf)
	}
}

// TestFileOpening tests the OpenFile function.
// CASE 1: The initializer must be called once.
// CASE 2: The data read on the second opening must be exactly the same as previously written on the first one.
func TestFileOpening(t *testing.T) {
	initCallCount := 0
	filePath := nextTestFilePath(t)
	open := func() (*Mapping, error) {
		return OpenFile(filePath, testFileMode, uintptr(testDataLength), 0, func(m *Mapping) error {
			initCallCount++
			_, err := m.WriteAt(testData, 0)
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
	buf := make([]byte, testDataLength)
	if _, err := m.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, testData) != 0 {
		t.Fatalf("data must be %v, %v found", testData, buf)
	}
}

// TestSegment tests the data segment.
// CASE: The read data must be exactly the same as the previously written unsigned 32-bit integer.
func TestSegment(t *testing.T) {
	m := openTestMapping(t, ModeReadWrite)
	defer closeTestEntity(t, m)
	*m.Segment().Uint32(0) = math.MaxUint32 - 1
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	f := openNextTestFile(t, true)
	defer closeTestEntity(t, f)
	buf := make([]byte, 4)
	if _, err := f.ReadAt(buf, 0); err != nil {
		t.Fatal(err)
	}
	uint32Data := []byte{254, 255, 255, 255}
	if bytes.Compare(buf, uint32Data) != 0 {
		t.Fatalf("data must be %v, %v found", uint32Data, buf)
	}
}
