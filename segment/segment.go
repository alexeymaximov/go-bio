// Package segment provides the data segment.
package segment

import (
	"encoding/binary"
	"math"
	"reflect"
	"unsafe"
)

// MaxUintptr is the maximum platform dependent unsigned integer
// large enough to store the uninterpreted bits of a pointer value.
const MaxUintptr = ^uintptr(0)

// Sizes of types in bytes.
const (
	Int8Size       = 1
	Int16Size      = 2
	Int32Size      = 4
	Int64Size      = 8
	Uint8Size      = 1
	Uint16Size     = 2
	Uint32Size     = 4
	Uint64Size     = 8
	Float32Size    = 4
	Float64Size    = 8
	Complex64Size  = 8
	Complex128Size = 16
)

// Segment is a data segment.
// See https://golang.org/ref/spec#Numeric_types for details.
type Segment struct {
	// offset specifies the offset of this segment.
	offset int64
	// data specifies the descriptor of the raw byte data associated with this segment.
	// TODO: Choose the valid type for this field and it's initialization mechanism.
	data reflect.SliceHeader
}

// New returns a new data segment.
func New(offset int64, data []byte) *Segment {
	return &Segment{
		offset: offset,
		data:   *(*reflect.SliceHeader)(unsafe.Pointer(&data)),
	}
}

// Pointer returns an untyped pointer to the value from this segment or panics at the access violation.
func (seg *Segment) Pointer(offset int64, length uintptr) uintptr {
	if offset < seg.offset || length > math.MaxInt64 {
		panic(Fault)
	}
	offset -= seg.offset
	if offset > math.MaxInt64-int64(length) || offset+int64(length) > int64(seg.data.Len) {
		panic(Fault)
	}
	if uint64(offset) > uint64(MaxUintptr-seg.data.Data) {
		panic(Fault)
	}
	return seg.data.Data + uintptr(offset)
}

// Int8 returns a pointer to the signed 8-bit integer from this segment or panics at the access violation.
func (seg *Segment) Int8(offset int64) *int8 {
	return (*int8)(unsafe.Pointer(seg.Pointer(offset, Int8Size)))
}

// Int16 returns a pointer to the signed 16-bit integer from this segment or panics at the access violation.
func (seg *Segment) Int16(offset int64) *int16 {
	return (*int16)(unsafe.Pointer(seg.Pointer(offset, Int16Size)))
}

// Int32 returns a pointer to the signed 32-bit integer from this segment or panics at the access violation.
func (seg *Segment) Int32(offset int64) *int32 {
	return (*int32)(unsafe.Pointer(seg.Pointer(offset, Int32Size)))
}

// Int64 returns a pointer to the signed 64-bit integer from this segment or panics at the access violation.
func (seg *Segment) Int64(offset int64) *int64 {
	return (*int64)(unsafe.Pointer(seg.Pointer(offset, Int64Size)))
}

// Uint8 returns a pointer to the unsigned 8-bit integer from this segment or panics at the access violation.
func (seg *Segment) Uint8(offset int64) *uint8 {
	return (*uint8)(unsafe.Pointer(seg.Pointer(offset, Uint8Size)))
}

// Uint16 returns a pointer to the unsigned 16-bit integer from this segment or panics at the access violation.
func (seg *Segment) Uint16(offset int64) *uint16 {
	return (*uint16)(unsafe.Pointer(seg.Pointer(offset, Uint16Size)))
}

// Uint32 returns a pointer to the unsigned 32-bit integer from this segment or panics at the access violation.
func (seg *Segment) Uint32(offset int64) *uint32 {
	return (*uint32)(unsafe.Pointer(seg.Pointer(offset, Uint32Size)))
}

// Uint16 returns a pointer to the unsigned 64-bit integer from this segment or panics at the access violation.
func (seg *Segment) Uint64(offset int64) *uint64 {
	return (*uint64)(unsafe.Pointer(seg.Pointer(offset, Uint64Size)))
}

// ScanUint sequentially reads the data into the unsigned integers pointed by v starting from the given offset.
func (seg *Segment) ScanUint(offset int64, v ...interface{}) error {
	data := *(*[]byte)(unsafe.Pointer(&seg.data))
	if offset < seg.offset {
		return ErrOutOfBounds
	}
	offset -= seg.offset
	for _, val := range v {
		switch value := val.(type) {
		default:
			return ErrBadValue
		case *uint8:
			if offset < 0 || offset > math.MaxInt64-Uint8Size || offset+Uint8Size > int64(len(data)) {
				return ErrOutOfBounds
			}
			*value = data[offset:][0]
			offset += Uint8Size
		case *uint16:
			if offset < 0 || offset > math.MaxInt64-Uint16Size || offset+Uint16Size > int64(len(data)) {
				return ErrOutOfBounds
			}
			*value = binary.LittleEndian.Uint16(data[offset : offset+Uint16Size])
			offset += Uint16Size
		case *uint32:
			if offset < 0 || offset > math.MaxInt64-Uint32Size || offset+Uint32Size > int64(len(data)) {
				return ErrOutOfBounds
			}
			*value = binary.LittleEndian.Uint32(data[offset : offset+Uint32Size])
			offset += Uint32Size
		case *uint64:
			if offset < 0 || offset > math.MaxInt64-Uint64Size || offset+Uint64Size > int64(len(data)) {
				return ErrOutOfBounds
			}
			*value = binary.LittleEndian.Uint64(data[offset : offset+Uint64Size])
			offset += Uint64Size
		}
	}
	return nil
}

// Float32 returns a pointer to the IEEE-754 32-bit floating-point number from this segment
// or panics at the access violation.
func (seg *Segment) Float32(offset int64) *float32 {
	return (*float32)(unsafe.Pointer(seg.Pointer(offset, Float32Size)))
}

// Float64 returns a pointer to the IEEE-754 64-bit floating-point number from this segment
// or panics at the access violation.
func (seg *Segment) Float64(offset int64) *float64 {
	return (*float64)(unsafe.Pointer(seg.Pointer(offset, Float64Size)))
}

// Complex64 returns a pointer to the complex number with float32 real and imaginary parts from this segment
// or panics at the access violation.
func (seg *Segment) Complex64(offset int64) *complex64 {
	return (*complex64)(unsafe.Pointer(seg.Pointer(offset, Complex64Size)))
}

// Complex128 returns a pointer to the complex number with float64 real and imaginary parts from this segment
// or panics at the access violation.
func (seg *Segment) Complex128(offset int64) *complex128 {
	return (*complex128)(unsafe.Pointer(seg.Pointer(offset, Complex128Size)))
}
