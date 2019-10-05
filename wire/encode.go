package wire

import (
	"encoding/json"
	"fmt"

	"github.com/JosiahWitt/erk"
)

// Errors
var (
	ErrNilDetails            = erk.New(ErkNilDetails{}, "details cannot be nil")
	ErrUnableToEncodeParams  = erk.New(ErkJSONMarshalling{}, "could not encode params")
	ErrUnableToEncodeDetails = erk.New(ErkJSONMarshalling{}, "could not encode details")
)

// Encode details and params to the Hoist Wire specification.
//
// If the params are already JSON encoded, call EncodeWithJSONParams.
func Encode(details, params interface{}) ([]byte, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, erk.WrapAs(ErrUnableToEncodeParams, err)
	}

	return EncodeWithJSONParams(details, paramsJSON)
}

// EncodeWithJSONParams encodes details and JSON encoded params to the Hoist Wire specification.
//
// If the params are not JSON encoded, call Encode.
func EncodeWithJSONParams(details interface{}, paramsJSON []byte) ([]byte, error) {
	if details == nil {
		return nil, ErrNilDetails
	}

	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return nil, erk.WrapAs(ErrUnableToEncodeDetails, err)
	}

	return []byte(fmt.Sprintf("%d,%d,%d:%s%s", DefaultEncoding, len(detailsJSON), len(paramsJSON), detailsJSON, paramsJSON)), nil
}
