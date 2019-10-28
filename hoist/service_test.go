package hoist_test

import (
	"testing"

	"github.com/hoistup/hoist-go/hoist"
	"github.com/matryer/is"
)

func TestNewService(t *testing.T) {
	is := is.New(t)

	myName := "abc"
	service := hoist.NewService(myName)
	exported := service.Export()

	expected := &hoist.ExportedService{
		Name:      myName,
		Functions: make(map[string]*hoist.ExportedFunction),
	}
	is.Equal(exported, expected)
}
