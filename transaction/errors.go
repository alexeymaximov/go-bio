package transaction

import "fmt"

// ErrClosed is the error which returns when tries to access the closed transaction.
var ErrClosed = fmt.Errorf("transaction: transaction closed")

// ErrOutOfBounds is the error which returns when tries to accessing the offset which is out of available bounds.
var ErrOutOfBounds = fmt.Errorf("transaction: out of bounds")
