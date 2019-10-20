package wire

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/JosiahWitt/erk"
)

// Errors
var (
	ErrUnableToReadInfoHeader = erk.New(ErkUnableToRead{}, "unable to read info header")
	ErrHeaderNotInt           = erk.New(ErkHeaderInvalid{}, "'{{.rawInfoPart}}' is not an int in info header")
	ErrHeaderMissingVersion   = erk.New(ErkHeaderInvalid{}, "info header must contain the encoding version")
	ErrHeaderVersionInvalid   = erk.New(ErkHeaderInvalid{}, "encoding version '{{.version}}' not implemented")
	ErrEncoding1Invalid       = erk.New(ErkEncodingInvalid{}, "requires exactly three info header parts (version, details length, params length) got '{{.infoHeader}}'")
	ErrUnableToReadDetails    = erk.New(ErkUnableToRead{}, "unable to read details")
	ErrUnableToReadParams     = erk.New(ErkUnableToRead{}, "unable to read params")
)

// DecodeResult contains the decoded parts.
type DecodeResult struct {
	Encoding   Encoding
	RawDetails []byte
	RawParams  []byte
}

// Decoder is created with NewDecoder, and stores the io.Reader for use between Decode() calls.
type Decoder struct {
	reader *bufio.Reader
}

// NewDecoder creates a decoder, which decodes from the provided io.Reader with each Decode() call.
func NewDecoder(rawReader io.Reader) *Decoder {
	return &Decoder{
		reader: bufio.NewReader(rawReader),
	}
}

// Decode the next message on the io.Reader.
func (d *Decoder) Decode() (*DecodeResult, error) {
	rawInfoHeader, err := d.reader.ReadString(':')
	if err != nil {
		return nil, erk.WrapAs(ErrUnableToReadInfoHeader, err)
	}

	infoHeader, err := parseInfoHeader(strings.TrimSuffix(rawInfoHeader, ":"))
	if err != nil {
		return nil, err
	}

	return d.decode(infoHeader)
}

func parseInfoHeader(rawInfoHeader string) ([]int, error) {
	rawInfoParts := strings.Split(rawInfoHeader, ",")
	infoHeader := make([]int, 0, len(rawInfoParts))

	// Build up the integer info header parts
	for _, rawInfoPart := range rawInfoParts {
		infoPart, err := strconv.Atoi(rawInfoPart)
		if err != nil {
			return nil, erk.WithParam(ErrHeaderNotInt, "rawInfoPart", rawInfoPart)
		}

		infoHeader = append(infoHeader, infoPart)
	}

	return infoHeader, nil
}

func (d *Decoder) decode(infoHeader []int) (*DecodeResult, error) {
	if len(infoHeader) < 1 {
		return nil, ErrHeaderMissingVersion // This case seems impossible to reach, since a message of ":" yields an ErrHeaderNotInt
	}

	switch infoHeader[0] {
	case 1:
		return d.decodeJSON(infoHeader)

	default:
		return nil, erk.WithParam(ErrHeaderVersionInvalid, "version", infoHeader[0])
	}
}

func (d *Decoder) decodeJSON(infoHeader []int) (*DecodeResult, error) {
	if len(infoHeader) != 3 {
		return nil, erk.WithParam(ErrEncoding1Invalid, "infoHeader", infoHeader)
	}

	rawDetails := make([]byte, infoHeader[1])
	if _, err := io.ReadFull(d.reader, rawDetails); err != nil {
		return nil, erk.WrapAs(ErrUnableToReadDetails, err)
	}

	rawParams := make([]byte, infoHeader[2])
	if _, err := io.ReadFull(d.reader, rawParams); err != nil {
		return nil, erk.WrapAs(ErrUnableToReadParams, err)
	}

	return &DecodeResult{
		Encoding:   EncodingJSON,
		RawDetails: rawDetails,
		RawParams:  rawParams,
	}, nil
}
