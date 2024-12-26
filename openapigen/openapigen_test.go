package openapigen_test

import (
	"openapi2go/openapigen"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequiredFieldToString(t *testing.T) {
	field := openapigen.Field{
		Name:     "SomeField",
		Type:     "integer",
		JsonName: "some_field",
		Required: true,
		Fields:   []openapigen.Field{},
	}

	actualCode := openapigen.FieldToString(field)
	expectedCode := "SomeField int `json:\"some_field\"`"
	require.Equal(t, expectedCode, actualCode)
}

func TestOptionalFieldToString(t *testing.T) {
	field := openapigen.Field{
		Name:     "SomeField",
		Type:     "integer",
		JsonName: "some_field",
		Required: false,
		Fields:   []openapigen.Field{},
	}

	actualCode := openapigen.FieldToString(field)
	expectedCode := "SomeField *int `json:\"some_field\"`"
	require.Equal(t, expectedCode, actualCode)
}
