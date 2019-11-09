// Package transaction provides the non-thread-safe transaction on the raw byte data.
package transaction

import (
	"runtime"
)

// Tx is a transaction on the raw byte data.
type Tx struct {
	// data specifies the raw byte data associated with this transaction.
	original []byte
	// lowOffset specifies the lowest offset, from start of the original,
	// which is available for this transaction.
	lowOffset int64
	// highOffset specifies the highest offset plus one, from start of the original,
	// which is available for this transaction.
	highOffset int64
	// snapshot specifies the snapshot of the original.
	snapshot []byte
}

// Begin starts and returns a new transaction.
// The given raw byte data starting from the given offset and ends after the given length
// copies to the snapshot which is allocated into the heap.
func Begin(data []byte, offset int64, length uintptr) (*Tx, error) {
	if offset < 0 || offset >= int64(len(data)) {
		return nil, ErrUnavailable
	}
	highOffset := offset + int64(length)
	if length == 0 || highOffset > int64(len(data)) {
		return nil, ErrUnavailable
	}
	tx := &Tx{
		original:   data,
		lowOffset:  offset,
		highOffset: highOffset,
		snapshot:   make([]byte, length),
	}
	copy(tx.snapshot, data[offset:highOffset])
	runtime.SetFinalizer(tx, (*Tx).Rollback)
	return tx, nil
}

// Offset returns the lowest offset from start of the original which is available for this transaction.
func (tx *Tx) Offset() int64 {
	return tx.lowOffset
}

// Length returns the length, in bytes, of the data which is available for this transaction.
func (tx *Tx) Length() uintptr {
	return uintptr(len(tx.snapshot))
}

// ReadAt reads len(buf) bytes at given offset from start of the original from the snapshot.
// If the given offset is outside of the accessible range the ErrUnavailable error will be returned.
// If there are not enough bytes to read then will be read how many there is
// and the number of read bytes will be returned with the ErrUnavailable error.
// Otherwise len(buf) will be returned with no errors.
// ReadAt implements the io.ReaderAt interface.
func (tx *Tx) ReadAt(buf []byte, offset int64) (int, error) {
	if tx.snapshot == nil {
		return 0, ErrClosed
	}
	if offset < tx.lowOffset || offset >= tx.highOffset {
		return 0, ErrUnavailable
	}
	n := copy(buf, tx.snapshot[offset-tx.lowOffset:])
	if n < len(buf) {
		return n, ErrUnavailable
	}
	return n, nil
}

// WriteAt writes len(buf) bytes at given offset from start of the original into the snapshot.
// If the given offset is outside of the accessible range the ErrUnavailable error will be returned.
// If there are not enough space to write all given bytes then will be written as much as possible
// and the number of written bytes will be returned with the ErrUnavailable error.
// Otherwise len(buf) will be returned with no errors.
// WriteAt implements the io.WriterAt interface.
func (tx *Tx) WriteAt(buf []byte, offset int64) (int, error) {
	if tx.snapshot == nil {
		return 0, ErrClosed
	}
	if offset < tx.lowOffset || offset >= tx.highOffset {
		return 0, ErrUnavailable
	}
	n := copy(tx.snapshot[offset-tx.lowOffset:], buf)
	if n < len(buf) {
		return n, ErrUnavailable
	}
	return n, nil
}

// Commit flushes the snapshot to the original, closes this transaction
// and frees all resources associated with it.
// Note that it doesn't check that the original is still available for writing.
func (tx *Tx) Commit() error {
	if tx.snapshot == nil {
		return ErrClosed
	}
	if n := copy(tx.original[tx.lowOffset:tx.highOffset], tx.snapshot); n < len(tx.snapshot) {
		return ErrUnavailable
	}
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
