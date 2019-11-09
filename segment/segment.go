// Package segment provides the data segment.
package segment

import (
	"encoding/binary"
	"io"
)

// Sizes of the unsigned integer types in bytes.
const (
	Uint8Size = 1 << iota
	Uint16Size
	Uint32Size
	Uint64Size
)

// ReadWriterAt is the interface that groups the basic io.ReadAt and io.WriteAt methods.
type ReadWriterAt interface {
	io.ReaderAt
	io.WriterAt
}

// Segment is a data segment.
// The supported data types are uint8, uint16, uint32 and uint64.
// All numeric values stored in the big-endian byte order.
type Segment struct {
	// driver specifies the data access driver.
	driver ReadWriterAt
}

// New returns a new data segment based on the given data access driver.
func New(driver ReadWriterAt) *Segment {
	return &Segment{
		driver: driver,
	}
}

// Get sequentially reads values pointed by v starting from the given offset.
func (seg *Segment) Get(offset int64, v ...interface{}) error {
	for _, val := range v {
		switch value := val.(type) {
		default:
			return ErrUnknown
		case *uint8:
			buf := make([]byte, Uint8Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			*value = buf[0]
			offset += Uint8Size
		case *uint16:
			buf := make([]byte, Uint16Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			*value = binary.BigEndian.Uint16(buf)
			offset += Uint16Size
		case *uint32:
			buf := make([]byte, Uint32Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			*value = binary.BigEndian.Uint32(buf)
			offset += Uint32Size
		case *uint64:
			buf := make([]byte, Uint64Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			*value = binary.BigEndian.Uint64(buf)
			offset += Uint64Size
		}
	}
	return nil
}

// Set sequentially writes values specified by v starting from the given offset.
func (seg *Segment) Set(offset int64, v ...interface{}) error {
	for _, val := range v {
		switch value := val.(type) {
		default:
			return ErrUnknown
		case uint8:
			buf := make([]byte, Uint8Size)
			buf[0] = value
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint8Size
		case uint16:
			buf := make([]byte, Uint16Size)
			binary.BigEndian.PutUint16(buf, value)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint16Size
		case uint32:
			buf := make([]byte, Uint32Size)
			binary.BigEndian.PutUint32(buf, value)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint32Size
		case uint64:
			buf := make([]byte, Uint64Size)
			binary.BigEndian.PutUint64(buf, value)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint64Size
		}
	}
	return nil
}

// Swap sequentially swaps values pointed by v with stored ones starting from the given offset.
func (seg *Segment) Swap(offset int64, v ...interface{}) error {
	for _, val := range v {
		switch value := val.(type) {
		default:
			return ErrUnknown
		case *uint8:
			buf := make([]byte, Uint8Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			tmp := *value
			*value = buf[0]
			buf[0] = tmp
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint8Size
		case *uint16:
			buf := make([]byte, Uint16Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			tmp := *value
			*value = binary.BigEndian.Uint16(buf)
			binary.BigEndian.PutUint16(buf, tmp)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint16Size
		case *uint32:
			buf := make([]byte, Uint32Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			tmp := *value
			*value = binary.BigEndian.Uint32(buf)
			binary.BigEndian.PutUint32(buf, tmp)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint32Size
		case *uint64:
			buf := make([]byte, Uint64Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			tmp := *value
			*value = binary.BigEndian.Uint64(buf)
			binary.BigEndian.PutUint64(buf, tmp)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint64Size
		}
	}
	return nil
}

// Inc sequentially increases values starting from the given offset using deltas specified by d.
func (seg *Segment) Inc(offset int64, d ...interface{}) error {
	for _, val := range d {
		switch value := val.(type) {
		default:
			return ErrUnknown
		case uint8:
			buf := make([]byte, Uint8Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			buf[0] += value
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint8Size
		case uint16:
			buf := make([]byte, Uint16Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			binary.BigEndian.PutUint16(buf, binary.BigEndian.Uint16(buf)+value)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint16Size
		case uint32:
			buf := make([]byte, Uint32Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			binary.BigEndian.PutUint32(buf, binary.BigEndian.Uint32(buf)+value)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint32Size
		case uint64:
			buf := make([]byte, Uint64Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			binary.BigEndian.PutUint64(buf, binary.BigEndian.Uint64(buf)+value)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint64Size
		}
	}
	return nil
}

// Dec sequentially decreases values starting from the given offset using deltas specified by d.
func (seg *Segment) Dec(offset int64, d ...interface{}) error {
	for _, val := range d {
		switch value := val.(type) {
		default:
			return ErrUnknown
		case uint8:
			buf := make([]byte, Uint8Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			buf[0] -= value
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint8Size
		case uint16:
			buf := make([]byte, Uint16Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			binary.BigEndian.PutUint16(buf, binary.BigEndian.Uint16(buf)-value)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint16Size
		case uint32:
			buf := make([]byte, Uint32Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			binary.BigEndian.PutUint32(buf, binary.BigEndian.Uint32(buf)-value)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint32Size
		case uint64:
			buf := make([]byte, Uint64Size)
			if _, err := seg.driver.ReadAt(buf, offset); err != nil {
				return err
			}
			binary.BigEndian.PutUint64(buf, binary.BigEndian.Uint64(buf)-value)
			if _, err := seg.driver.WriteAt(buf, offset); err != nil {
				return err
			}
			offset += Uint64Size
		}
	}
	return nil
}
