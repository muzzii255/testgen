package generator

var mainTest = `package gentests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var testApp *your.App // your app


func TestMain(m *testing.M) {
	// configure your app here
}

func setup() *your.App {
	return testApp
}

func testHelloWorld(t *testing.T) {
	app := setup()
	payload := map[string]string{"name": "Deeznuts"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var responseMap map[string]string
	json.NewDecoder(resp.Body).Decode(&responseMap)
	require.Equal(t, "hello Deeznuts", responseMap["result"])
}

// func TestSequentially(t *testing.T) {
// 	t.Run("HelloWorld", testHelloWorld)
// 	// Run your tests here
// }
`

var helperFile = `package gentests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)


func makeReq(t *testing.T, app *your.App, method, path string, body any) *http.Response {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("content-type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func decodeResp[T any](t *testing.T, resp *http.Response) (T, []byte) {
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var v T
	err = json.Unmarshal(body, &v)
	require.NoError(t, err)
	return v, body
}


func Ptr[T any](v T) *T { return &v }

`
