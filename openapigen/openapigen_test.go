package openapigen_test

import (
	"testing"

	"github.com/FS-Frost/openapi2go/openapigen"
	"github.com/stretchr/testify/require"
)

func TestFieldToString(t *testing.T) {
	type TestCase struct {
		Field        openapigen.Field
		ExpectedCode string
	}

	testCases := []TestCase{
		{
			Field: openapigen.Field{
				Name:     "SomeField",
				Type:     "integer",
				JsonName: "some_field",
				Required: true,
				Fields:   []openapigen.Field{},
			},
			ExpectedCode: "SomeField int `json:\"some_field\"`",
		},
		{
			Field: openapigen.Field{
				Name:     "SomeField",
				Type:     "integer",
				JsonName: "some_field",
				Required: false,
				Fields:   []openapigen.Field{},
			},
			ExpectedCode: "SomeField *int `json:\"some_field\"`",
		},
		{
			Field: openapigen.Field{
				Name:     "SomeField",
				Type:     "number",
				JsonName: "some_field",
				Required: true,
				Fields:   []openapigen.Field{},
			},
			ExpectedCode: "SomeField float64 `json:\"some_field\"`",
		},
	}

	for _, testCase := range testCases {
		actualCode := openapigen.FieldToString(testCase.Field)
		require.Equal(t, testCase.ExpectedCode, actualCode)
	}
}
