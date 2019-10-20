package wire_test

import (
	"errors"
	"testing"

	"github.com/hoistup/hoist-go/wire"
	"github.com/matryer/is"
)

type testReader struct {
	next []byte
}

func (r *testReader) Read(p []byte) (n int, err error) {
	if len(r.next) == 0 {
		return 0, errors.New("testReader out of bytes")
	}

	end := len(p)
	if end > len(r.next) {
		end = len(r.next)
	}

	copy(p, r.next[:end])
	r.next = r.next[end:]
	return end, nil
}

func TestDecode(t *testing.T) {
	table := []struct {
		Name           string
		Message        []byte
		ExpectedResult *wire.DecodeResult
		ExpectedError  error
	}{
		{
			Name:    "with present details and present params",
			Message: []byte(`1,17,19:{"service":"abc"}{"message":"hello"}`),
			ExpectedResult: &wire.DecodeResult{
				Encoding:   1,
				RawDetails: []byte(`{"service":"abc"}`),
				RawParams:  []byte(`{"message":"hello"}`),
			},
			ExpectedError: nil,
		},
		{
			Name:    "with null details and null params",
			Message: []byte("1,4,4:nullnull"),
			ExpectedResult: &wire.DecodeResult{
				Encoding:   1,
				RawDetails: []byte("null"),
				RawParams:  []byte("null"),
			},
			ExpectedError: nil,
		},
		{
			Name:           "with invalid message",
			Message:        []byte(""),
			ExpectedResult: nil,
			ExpectedError:  wire.ErrUnableToReadInfoHeader,
		},
		{
			Name:           "with non int in info header",
			Message:        []byte("1,d,3:"),
			ExpectedResult: nil,
			ExpectedError:  wire.ErrHeaderNotInt,
		},
		{
			Name:           "with empty info header",
			Message:        []byte(":"),
			ExpectedResult: nil,
			ExpectedError:  wire.ErrHeaderNotInt,
		},
		{
			Name:           "with invalid encoding version",
			Message:        []byte("2:"),
			ExpectedResult: nil,
			ExpectedError:  wire.ErrHeaderVersionInvalid,
		},
		{
			Name:           "encoding version 1: incorrect number of headers",
			Message:        []byte("1,4,5,6:"),
			ExpectedResult: nil,
			ExpectedError:  wire.ErrEncoding1Invalid,
		},
		{
			Name:           "encoding version 1: unable to read details",
			Message:        []byte("1,6,5:null"),
			ExpectedResult: nil,
			ExpectedError:  wire.ErrUnableToReadDetails,
		},
		{
			Name:           "encoding version 1: unable to read params",
			Message:        []byte("1,4,5:nullnull"),
			ExpectedResult: nil,
			ExpectedError:  wire.ErrUnableToReadParams,
		},
	}

	r := &testReader{}
	d := wire.NewDecoder(r)
	for _, entry := range table {
		t.Run(entry.Name, func(t *testing.T) {
			is := is.New(t)
			r.next = entry.Message

			res, err := d.Decode()
			is.True(errors.Is(err, entry.ExpectedError))
			is.Equal(res, entry.ExpectedResult)
		})
	}
}
