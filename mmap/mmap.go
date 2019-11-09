// Package mmap provides the non-thread-safe cross-platform memory mapped file I/O.
package mmap

import (
	"github.com/alexeymaximov/go-bio/segment"
	"github.com/alexeymaximov/go-bio/transaction"
)

// Mode is a mapping mode.
type Mode int

const (
	// Share this mapping and allow the read-only access.
	ModeReadOnly Mode = iota

	// Share this mapping.
	// Updates to the mapping are visible to other processes
	// mapping the same region, and are carried through to the underlying file.
	// To precisely control when updates are carried through to the underlying file
	// requires the use of Mapping.Sync.
	ModeReadWrite

	// Create a private copy-on-write mapping.
	// Updates to the mapping are not visible to other processes
	// mapping the same region, and are not carried through to the underlying file.
	// It is unspecified whether changes made to the file are visible in the mapped region.
	ModeWriteCopy
)

// Flag is a mapping flag.
type Flag int

const (
	// Mapped memory pages may be executed.
	FlagExecutable Flag = 1 << iota
)

// generic is a cross-platform parts of a mapping.
type generic struct {
	// writable specifies whether the mapped memory pages may be written.
	writable bool
	// executable specifies whether the mapped memory pages may be executed.
	executable bool
	// address specifies the pointer to the mapped memory.
	address uintptr
	// memory specifies the byte slice which wraps the mapped memory.
	memory []byte
	// segment specifies the lazily initialized data segment on top of this mapping.
	segment *segment.Segment
}

// Writable returns true if the mapped memory pages may be written.
func (m *Mapping) Writable() bool {
	return m.writable
}

// Executable returns true if the mapped memory pages may be executed.
func (m *Mapping) Executable() bool {
	return m.executable
}

// Address returns the pointer to the mapped memory.
func (m *Mapping) Address() uintptr {
	return m.address
}

// Length returns the mapped memory length in bytes.
func (m *Mapping) Length() uintptr {
	return uintptr(len(m.memory))
}

// Memory returns the byte slice which wraps the mapped memory.
func (m *Mapping) Memory() []byte {
	return m.memory
}

// ReadAt reads len(buf) bytes at the given offset from start of the mapped memory from the mapped memory.
// If the given offset is outside of the accessible range the ErrUnavailable error will be returned.
// If there are not enough bytes to read then will be read how many there is
// and the number of read bytes will be returned with the ErrUnavailable error.
// Otherwise len(buf) will be returned with no errors.
// ReadAt implements the io.ReaderAt interface.
func (m *Mapping) ReadAt(buf []byte, offset int64) (int, error) {
	if m.memory == nil {
		return 0, ErrClosed
	}
	if offset < 0 || offset >= int64(len(m.memory)) {
		return 0, ErrUnavailable
	}
	n := copy(buf, m.memory[offset:])
	if n < len(buf) {
		return n, ErrUnavailable
	}
	return n, nil
}

// WriteAt writes len(buf) bytes at the given offset from start of the mapped memory into the mapped memory.
// If the given offset is outside of the accessible range the ErrUnavailable error will be returned.
// If there are not enough space to write all given bytes then will be written as much as possible
// and the number of written bytes will be returned with the ErrUnavailable error.
// Otherwise len(buf) will be returned with no errors.
// WriteAt implements the io.WriterAt interface.
func (m *Mapping) WriteAt(buf []byte, offset int64) (int, error) {
	if m.memory == nil {
		return 0, ErrClosed
	}
	if !m.writable {
		return 0, ErrReadOnly
	}
	if offset < 0 || offset >= int64(len(m.memory)) {
		return 0, ErrUnavailable
	}
	n := copy(m.memory[offset:], buf)
	if n < len(buf) {
		return n, ErrUnavailable
	}
	return n, nil
}

// Segment returns the data segment on top of this mapping.
func (m *Mapping) Segment() *segment.Segment {
	if m.segment == nil {
		m.segment = segment.New(m)
	}
	return m.segment
}

// Begin starts and returns a new transaction.
func (m *Mapping) Begin(offset int64, length uintptr) (*transaction.Tx, error) {
	if m.memory == nil {
		return nil, ErrClosed
	}
	if !m.writable {
		return nil, ErrReadOnly
	}
	return transaction.Begin(m.memory, offset, length)
}
