package mmap

import "fmt"

// ErrBadOffset is an error which returns when the given length is not valid.
var ErrBadLength = fmt.Errorf("mmap: bad length")

// ErrBadMode is an error which returns when the given mapping mode is not valid.
var ErrBadMode = fmt.Errorf("mmap: bad mode")

// ErrBadOffset is an error which returns when the given offset is not valid.
var ErrBadOffset = fmt.Errorf("mmap: bad offset")

// ErrClosed is the error which returns when tries to access the closed mapping.
var ErrClosed = fmt.Errorf("mmap: mapping closed")

// ErrLocked is the error which returns when the mapping memory pages were already locked.
var ErrLocked = fmt.Errorf("mmap: mapping already locked")

// ErrNotLocked is the error which returns when the mapping memory pages are not locked.
var ErrNotLocked = fmt.Errorf("mmap: mapping is not locked")

// ErrReadOnly is the error which returns when tries to execute a write operation on the read-only mapping.
var ErrReadOnly = fmt.Errorf("mmap: mapping is read only")

// ErrOutOfBounds is the error which returns when tries to accessing the offset which is out of available bounds.
var ErrOutOfBounds = fmt.Errorf("mmap: out of bounds")
