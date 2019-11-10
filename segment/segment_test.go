package segment

import (
	"math"
	"testing"
)

// Maximal values of the unsigned integer types.
const (
	maxUint8  = uint8(math.MaxUint8)
	maxUint16 = uint16(math.MaxUint16)
	maxUint32 = uint32(math.MaxUint32)
	maxUint64 = uint64(math.MaxUint64)
)

//------------------------------------------- TEST CASES ---------------------------------------------------------------

// TestOffset tests the segment offset.
// CASE: First byte MUST NOT be modified.
func TestOffset(t *testing.T) {
	data := make([]byte, 9)
	seg := New(1, data[1:])
	*seg.Uint64(1) = math.MaxUint64
	if data[0] != 0 {
		t.Fatalf("first byte must be zero, %d found", data[0])
	}
}

// TestScanUint tests the unsigned integers scanning.
// CASE: The read values MUST be exactly the same as the previously written.
func TestScanUint(t *testing.T) {
	seg := New(0, make([]byte, 16))
	off := int64(1)
	in8, in16, in32, in64 := maxUint8-1, maxUint16-200, maxUint32-3_000, maxUint64-40_000
	*seg.Uint8(off) = in8
	*seg.Uint16(off + Uint8Size) = in16
	*seg.Uint32(off + Uint8Size + Uint16Size) = in32
	*seg.Uint64(off + Uint8Size + Uint16Size + Uint32Size) = in64
	out8, out16, out32, out64 := uint8(1), uint16(1), uint32(1), uint64(1)
	if err := seg.ScanUint(off, &out8, &out16, &out32, &out64); err != nil {
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
