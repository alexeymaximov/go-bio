// Package transaction provides the transaction on the raw byte data.
package transaction

import (
	"math"
	"runtime"

	"github.com/alexeymaximov/go-bio/segment"
)

// Tx is a transaction on the raw byte data.
type Tx struct {
	// original specifies the raw byte data associated with this transaction.
	original []byte
	// lowOffset specifies the lowest offset, from start of the original,
	// which is available for this transaction.
	lowOffset int64
	// highOffset specifies the highest offset plus one, from start of the original,
	// which is available for this transaction.
	highOffset int64
	// snapshot specifies the snapshot of the original.
	snapshot []byte
	// segment specifies the lazily initialized data segment on top of the snapshot.
	segment *segment.Segment
}

// Begin starts and returns a new transaction.
// The given raw byte data starting from the given offset and ends after the given length
// copies to the snapshot which is allocated into the heap.
func Begin(data []byte, offset int64, length uintptr) (*Tx, error) {
	if length == 0 || length > math.MaxInt64 {
		return nil, ErrOutOfBounds
	}
	if offset < 0 || offset >= int64(len(data)) || offset > math.MaxInt64-int64(length) {
		return nil, ErrOutOfBounds
	}
	highOffset := offset + int64(length)
	if highOffset > int64(len(data)) {
		return nil, ErrOutOfBounds
	}
	tx := &Tx{
		original:   data,
		lowOffset:  offset,
		highOffset: highOffset,
		snapshot:   make([]byte, length),
	}
	copy(tx.snapshot, data[tx.lowOffset:tx.highOffset])
	runtime.SetFinalizer(tx, (*Tx).Rollback)
	return tx, nil
}

// Segment returns the data segment on top of the snapshot.
func (tx *Tx) Segment() *segment.Segment {
	if tx.segment == nil {
		tx.segment = segment.New(tx.lowOffset, tx.snapshot)
	}
	return tx.segment
}

// offset checks given offset and length to match the available bounds and returns the relative offset
// from start of the segment data or ErrOutOfBounds error at the access violation.
func (tx *Tx) offset(offset int64, length int) (int64, error) {
	if offset < tx.lowOffset {
		return 0, ErrOutOfBounds
	}
	offset -= tx.lowOffset
	if offset > math.MaxInt64-int64(length) || offset+int64(length) > tx.highOffset {
		return 0, ErrOutOfBounds
	}
	return offset, nil
}

// ReadAt reads len(buf) bytes at given offset from start of the original from the snapshot.
// If the given offset is out of the available bounds or there are not enough bytes to read
// the ErrOutOfBounds error will be returned. Otherwise len(buf) will be returned with no errors.
// ReadAt implements the io.ReaderAt interface.
func (tx *Tx) ReadAt(buf []byte, offset int64) (int, error) {
	if tx.snapshot == nil {
		return 0, ErrClosed
	}
	off, err := tx.offset(offset, len(buf))
	if err != nil {
		return 0, err
	}
	return copy(buf, tx.snapshot[off:]), nil
}

// WriteAt writes len(buf) bytes at given offset from start of the original into the snapshot.
// If the given offset is out of the available bounds or there are not enough space to write all given bytes
// the ErrOutOfBounds error will be returned. Otherwise len(buf) will be returned with no errors.
// WriteAt implements the io.WriterAt interface.
func (tx *Tx) WriteAt(buf []byte, offset int64) (int, error) {
	if tx.snapshot == nil {
		return 0, ErrClosed
	}
	off, err := tx.offset(offset, len(buf))
	if err != nil {
		return 0, err
	}
	return copy(tx.snapshot[off:], buf), nil
}

// Commit flushes the snapshot to the original, closes this transaction
// and frees all resources associated with it.
func (tx *Tx) Commit() error {
	if tx.snapshot == nil {
		return ErrClosed
	}
	copy(tx.original[tx.lowOffset:tx.highOffset], tx.snapshot)
	tx.snapshot = nil
	return nil
}

// Rollback closes this transaction and frees all resources associated with it.
func (tx *Tx) Rollback() error {
	if tx.snapshot == nil {
		return ErrClosed
	}
	tx.snapshot = nil
	return nil
}
