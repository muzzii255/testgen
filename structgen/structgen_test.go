package structgen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsBuiltinType(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"bool", true},
		{"string", true},
		{"int", true},
		{"int8", true},
		{"int16", true},
		{"int32", true},
		{"int64", true},
		{"uint", true},
		{"uint8", true},
		{"uint16", true},
		{"uint32", true},
		{"uint64", true},
		{"uintptr", true},
		{"byte", true},
		{"rune", true},
		{"float32", true},
		{"float64", true},
		{"complex64", true},
		{"complex128", true},
		{"error", true},
		{"int256", false},
		{"float128", false},
		{"User", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isBuiltinType(tt.input)
		require.Equal(t, tt.expected, result)
	}
}

func TestGetCorrectValue(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{int(42), "42"},
		{string("hello"), `"hello"`},
		{bool(true), "true"},
		{bool(false), "false"},
		{float64(3.14), "3.14"},
	}

	for _, tt := range tests {
		result := getCorrectValue(tt.input)
		require.Equal(t, tt.expected, result)
	}
}

func TestCleanPkgPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"./", "."},
		{"./models", "models."},
		{"./pkg/api", "pkg/api."},
		{"models", "models."},
	}

	for _, tt := range tests {
		result := cleanPkgPath(tt.input)
		require.Equal(t, tt.expected, result)
	}
}

func TestSliceString(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{}, "{}"},
		{[]string{"a"}, `{"a",}`},
		{[]string{"a", "b", "c"}, `{"a","b","c",}`},
	}

	for _, tt := range tests {
		result := sliceString(tt.input)
		require.Equal(t, tt.expected, result)
	}
}

func TestGetJsonTag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"`json:\"name\"`", "name"},
		{"`json:\"user_id\"`", "user_id"},
		{"`json:\"\"`", ""},
		{"`yaml:\"name\"`", ""},
		{"", ""},
	}

	for _, tt := range tests {
		result := getJsonTag(tt.input)
		require.Equal(t,tt.expected,result)
	}
}
