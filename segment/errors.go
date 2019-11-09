package segment

import "fmt"

// ErrUnknown is the error which returns when the given value has an unknown type.
var ErrUnknown = fmt.Errorf("segment: unknown value")
