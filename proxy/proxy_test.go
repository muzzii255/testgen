package proxy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "noname"},
		{"/", "noname"},
		{"/api/users", "users"},
		{"/api/users/123", "users"},
		{"/api/users/123/profile", "users_profile"},
		{"/api", "api"},
		{"/api/v1/users?name=test", "users"},
		{"/api/v1/users/abc", "users"},
		{"///", "noname"},
		{"/a/b/c/d/e/f/g", "g"},
	}

	for _, tt := range tests {
		result := cleanPath(tt.input)
		require.Equal(t, tt.expected, result)
	}
}

func TestCleanURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"/", "/"},
		{"/api/users", "/api/users"},
		{"/api/users/123", "/api/users/:id"},
		{"/api/users/123/profile", "/api/users/:id/profile"},
		{"/api/456/items/789", "/api/:id/items/:id"},
		{"/api/abc/def", "/api/abc/def"},
	}

	for _, tt := range tests {
		result := cleanURL(tt.input)
		require.Equal(t, tt.expected, result)
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "/"},
		{"/", "/"},
		{"/api/users", "/api/users"},
		{"/api/users/123", "/api/users"},
		{"/api/456/items/789", "/api/items"},
		{"/api/v1/users/123", "/api/v1/users"},
		{"///api///users///123", "/api/users"},
		{"/api//v1//users", "/api/v1/users"},
	}

	for _, tt := range tests {
		result := normalizeURL(tt.input)
		require.Equal(t, tt.expected, result)
	}
}
