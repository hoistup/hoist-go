// Package wire implements the Hoist Wire specification (https://github.com/hoistup/hoist/tree/master/specs/wire.md).
package wire

import "github.com/JosiahWitt/erk"

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
	ErkNilDetails      erk.DefaultKind
	ErkJSONMarshalling erk.DefaultKind
	ErkUnableToRead    erk.DefaultKind
	ErkHeaderInvalid   erk.DefaultKind
	ErkEncodingInvalid erk.DefaultKind
)
