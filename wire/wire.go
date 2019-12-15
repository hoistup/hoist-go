// Package wire implements the Hoist Wire specification (https://github.com/hoistup/hoist/tree/master/specs/wire.md).
package wire

import "github.com/hoistup/hoist-go/erks"

// Encoding tracks what Wire encoding is used.
type Encoding int

// DefaultEncoding is 1, which is the JSON encoding.
// This may change in the future, if other types are supported.
const DefaultEncoding = EncodingJSON

// Various encodings
const (
	EncodingJSON Encoding = 1
)

// Error kinds
type (
	ErkNilDetails      struct{ erks.Default }
	ErkJSONMarshalling struct{ erks.Default }
	ErkUnableToRead    struct{ erks.Default }
	ErkHeaderInvalid   struct{ erks.Default }
	ErkEncodingInvalid struct{ erks.Default }
)
