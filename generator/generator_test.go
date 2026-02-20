package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/muzzii255/testgen/proxy"
	"github.com/stretchr/testify/require"
)

func TestScanner_scanFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main
// @testgen router=/api/users struct=User
func main() {}
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	scanner := &Scanner{InputDir: tmpDir}
	results, err := scanner.scanFile(testFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	require.Equal(t, 1, len(results))
	require.Equal(t, "// @testgen router=/api/users struct=User", results[0])
}

func TestScanner_scanFileNoTags(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main
func main() {}
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	scanner := &Scanner{InputDir: tmpDir}
	results, err := scanner.scanFile(testFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	require.Equal(t, 0, len(results))
}

func TestScanner_getTags(t *testing.T) {
	scanner := &Scanner{}

	tests := []struct {
		name       string
		input      string
		wantRouter string
		wantStruct string
		wantErr    bool
	}{
		{
			name:       "valid tags",
			input:      " @testgen router=/api/users struct=User",
			wantRouter: "/api/users",
			wantStruct: "User",
			wantErr:    false,
		},
		{
			name:       "valid tags with package",
			input:      " @testgen router=/api/users struct=models.User",
			wantRouter: "/api/users",
			wantStruct: "models.User",
			wantErr:    false,
		},
		{
			name:       "missing router",
			input:      " @testgen struct=User",
			wantRouter: "",
			wantStruct: "",
			wantErr:    true,
		},
		{
			name:       "missing struct",
			input:      " @testgen router=/api/users",
			wantRouter: "",
			wantStruct: "",
			wantErr:    true,
		},
		{
			name:       "empty string",
			input:      "",
			wantRouter: "",
			wantStruct: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRouter, gotStruct, err := scanner.getTags(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantRouter, gotRouter)
			require.Equal(t, tt.wantStruct, gotStruct)
		})
	}
}

func TestScanner_ScanTags(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main
	// @testgen router=/api/users struct=models.User
	func main() {}
	`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	scanner := &Scanner{InputDir: tmpDir}
	results, err := scanner.ScanTags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	require.Equal(t, len(results), 1)

	data, ok := results["/api/users"]
	if !ok {
		t.Fatal("expected /api/users key")
	}

	require.Equal(t, "./models", data["folder"])
	require.Equal(t, "User", data["struct"])
	require.Equal(t, "models.User", data["name"])
}

func TestScanner_ScanTagsNoFolder(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main
// @testgen router=/api/health struct=Health
func main() {}
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	scanner := &Scanner{InputDir: tmpDir}
	results, err := scanner.ScanTags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, ok := results["/api/health"]
	if !ok {
		t.Fatalf("unexpected error: %v", err)
	}
	require.Equal(t, "./", data["folder"])
	require.Equal(t, "Health", data["struct"])
}

func TestFilterByMethod(t *testing.T) {
	rows := []struct {
		Method string
	}{
		{Method: "POST"},
		{Method: "GET"},
		{Method: "POST"},
		{Method: "DELETE"},
	}

	type BodyRecords struct {
		Method string
	}

	records := make([]proxy.BodyRecords, len(rows))
	for i, r := range rows {
		records[i] = proxy.BodyRecords{Method: r.Method}
	}

	result := filterByMethod(records, "POST")
	require.Equal(t, 2, len(result))

	result = filterByMethod(records, "GET")
	require.Equal(t, 1, len(result))

	result = filterByMethod(records, "PUT")
	require.Equal(t, 0, len(result))
}

func TestGetFuncName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/users", "users"},
		{"/api/users/123", "users"},
		{"/api/products/abc/details", "productsdetails"},
		{"/api", "NoName"},
		{"/", "NoName"},
	}

	for _, tt := range tests {
		result := getFuncName(tt.input)
		require.Equal(t, tt.expected, result)
	}
}
