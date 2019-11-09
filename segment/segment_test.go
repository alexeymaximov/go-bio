package segment

import (
	"fmt"
	"testing"
)

// Maximal values of the unsigned integer types.
const (
	maxUint8  = uint8(255)
	maxUint16 = uint16(65_535)
	maxUint32 = uint32(4_294_967_295)
	maxUint64 = uint64(18_446_744_073_709_551_615)
)

// errTestUnavailable is the error which returns when tries to accessing the data which is not available.
var errTestUnavailable = fmt.Errorf("test: data not available")

// testDriver is a data access driver for tests.
type testDriver []byte

// ReadAt implements the io.ReaderAt interface.
func (d testDriver) ReadAt(buf []byte, offset int64) (int, error) {
	if offset < 0 || offset >= int64(len(d)) {
		return 0, errTestUnavailable
	}
	n := copy(buf, d[offset:])
	if n < len(buf) {
		return n, errTestUnavailable
	}
	return n, nil
}

// WriteAt implements the io.WriterAt interface.
func (d testDriver) WriteAt(buf []byte, offset int64) (int, error) {
	if offset < 0 || offset >= int64(len(d)) {
		return 0, errTestUnavailable
	}
	n := copy(d[offset:], buf)
	if n < len(buf) {
		return n, errTestUnavailable
	}
	return n, nil
}

//------------------------------------------- TEST CASES ---------------------------------------------------------------

// TestGetSet tests the simple data segment I/O operations.
// CASE: The read values MUST be exactly the same as the previously written.
func TestGetSet(t *testing.T) {
	seg := New(make(testDriver, 16))
	off := int64(1)
	in8, in16, in32, in64 := maxUint8-1, maxUint16-200, maxUint32-3_000, maxUint64-40_000
	if err := seg.Set(off, in8, in16, in32, in64); err != nil {
		t.Fatal(err)
	}
	out8, out16, out32, out64 := uint8(1), uint16(1), uint32(1), uint64(1)
	if err := seg.Get(off, &out8, &out16, &out32, &out64); err != nil {
		t.Fatal(err)
	}
	if in8 != out8 {
		t.Fatalf("uint8 value must be %d, %d found", in8, out8)
	}
	if in16 != out16 {
		t.Fatalf("uint16 value must be %d, %d found", in16, out16)
	}
	if in32 != out32 {
		t.Fatalf("uint32 value must be %d, %d found", in32, out32)
	}
	if in64 != out64 {
		t.Fatalf("uint64 value must be %d, %d found", in64, out64)
	}
}

// TestSwap tests the swap operation of the data segment.
// CASE 1: The values read on the swapping MUST be exactly the same as the previously written.
// CASE 2: The values read after the swapping MUST be exactly the same as the written on the swapping.
func TestSwap(t *testing.T) {
	seg := New(make(testDriver, 16))
	off := int64(1)
	in8, in16, in32, in64 := maxUint8, maxUint16, maxUint32, maxUint64
	if err := seg.Set(off, in8, in16, in32, in64); err != nil {
		t.Fatal(err)
	}
	src8, src16, src32, src64 := maxUint8-1, maxUint16-201, maxUint32-3_002, maxUint64-40_003
	dst8, dst16, dst32, dst64 := src8, src16, src32, src64
	if err := seg.Swap(off, &dst8, &dst16, &dst32, &dst64); err != nil {
		t.Fatal(err)
	}
	if in8 != dst8 {
		t.Fatalf("uint8 value must be %d, %d found", in8, dst8)
	}
	if in16 != dst16 {
		t.Fatalf("uint16 value must be %d, %d found", in16, dst16)
	}
	if in32 != dst32 {
		t.Fatalf("uint32 value must be %d, %d found", in32, dst32)
	}
	if in64 != dst64 {
		t.Fatalf("uint64 value must be %d, %d found", in64, dst64)
	}
	out8, out16, out32, out64 := uint8(1), uint16(1), uint32(1), uint64(1)
	if err := seg.Get(off, &out8, &out16, &out32, &out64); err != nil {
		t.Fatal(err)
	}
	if src8 != out8 {
		t.Fatalf("uint8 value must be %d, %d found", src8, out8)
	}
	if src16 != out16 {
		t.Fatalf("uint16 value must be %d, %d found", src16, out16)
	}
	if src32 != out32 {
		t.Fatalf("uint32 value must be %d, %d found", src32, out32)
	}
	if src64 != out64 {
		t.Fatalf("uint64 value must be %d, %d found", src64, out64)
	}
}

// TestIncDec tests the increment and decrement operations of the data segment.
// CASE 1: The values read after the previous writing and increasing MUST be equal to
// deltas minus one by the reason of overflow.
// CASE 2: The values read after the previous decreasing MUST be exactly the same as
// the originally written.
func TestIncDec(t *testing.T) {
	seg := New(make(testDriver, 16))
	off := int64(1)
	in8, in16, in32, in64 := maxUint8, maxUint16, maxUint32, maxUint64
	if err := seg.Set(off, in8, in16, in32, in64); err != nil {
		t.Fatal(err)
	}
	d8, d16, d32, d64 := uint8(1), uint16(2), uint32(3), uint64(4)
	if err := seg.Inc(off, d8, d16, d32, d64); err != nil {
		t.Fatal(err)
	}
	out8, out16, out32, out64 := uint8(0), uint16(0), uint32(0), uint64(0)
	if err := seg.Get(off, &out8, &out16, &out32, &out64); err != nil {
		t.Fatal(err)
	}
	if d8-1 != out8 {
		t.Fatalf("uint8 value must be %d, %d found", d8, out8)
	}
	if d16-1 != out16 {
		t.Fatalf("uint16 value must be %d, %d found", d16, out16)
	}
	if d32-1 != out32 {
		t.Fatalf("uint32 value must be %d, %d found", d32, out32)
	}
	if d64-1 != out64 {
		t.Fatalf("uint64 value must be %d, %d found", d64, out64)
	}
	if err := seg.Dec(off, d8, d16, d32, d64); err != nil {
		t.Fatal(err)
	}
	if err := seg.Get(off, &out8, &out16, &out32, &out64); err != nil {
		t.Fatal(err)
	}
	if in8 != out8 {
		t.Fatalf("uint8 value must be %d, %d found", in8, out8)
	}
	if in16 != out16 {
		t.Fatalf("uint16 value must be %d, %d found", in16, out16)
	}
	if in32 != out32 {
		t.Fatalf("uint32 value must be %d, %d found", in32, out32)
	}
	if in64 != out64 {
		t.Fatalf("uint64 value must be %d, %d found", in64, out64)
	}
}

// TestError tests the data segment error.
// CASE: errTestUnavailable MUST be returned on the writing by the reason of
// the segment size is one less than necessary.
func TestError(t *testing.T) {
	seg := New(make(testDriver, 15))
	off := int64(1)
	in8, in16, in32, in64 := maxUint8, maxUint16, maxUint32, maxUint64
	if err := seg.Set(off, in8, in16, in32, in64); err == nil {
		t.Fatal("expected errTestUnavailable, no error found")
	} else if err != errTestUnavailable {
		t.Fatalf("expected errTestUnavailable, [%v] error found", err)
	}
}
