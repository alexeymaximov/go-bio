package transaction

import "fmt"

// ErrClosed is the error which returns when tries to access the closed transaction.
var ErrClosed = fmt.Errorf("transaction: transaction closed")

// ErrUnavailable is the error which returns when tries to accessing the data which is not available.
var ErrUnavailable = fmt.Errorf("transaction: data not available")
