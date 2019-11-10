package segment

import "fmt"

// ErrBadValue is the error which returns when the given value is of the type incompatible with the operation.
var ErrBadValue = fmt.Errorf("segment: bad value")

// ErrOutOfBounds is the error which returns when tries to accessing the offset which is out of available bounds.
var ErrOutOfBounds = fmt.Errorf("segment: out of bounds")

// Fault is the access violation error.
var Fault = fmt.Errorf("segmentation fault")
