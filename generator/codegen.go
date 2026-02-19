package generator

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/muzzii255/testgen/proxy"
	"github.com/muzzii255/testgen/structgen"
)

const testFileDir = "./gentest"

var statusMap = map[int]string{
	200: "http.StatusOK",
	201: "http.StatusCreated",
	202: "http.StatusAccepted",
	204: "http.StatusNoContent",
	301: "http.StatusMovedPermanently",
	302: "http.StatusFound",
	304: "http.StatusNotModified",
	307: "http.StatusTemporaryRedirect",
	308: "http.StatusPermanentRedirect",
	400: "http.StatusBadRequest",
	401: "http.StatusUnauthorized",
	403: "http.StatusForbidden",
	404: "http.StatusNotFound",
	405: "http.StatusMethodNotAllowed",
	409: "http.StatusConflict",
	422: "http.StatusUnprocessableEntity",
	429: "http.StatusTooManyRequests",
	500: "http.StatusInternalServerError",
	502: "http.StatusBadGateway",
	503: "http.StatusServiceUnavailable",
	504: "http.StatusGatewayTimeout",
}

type JsonFile struct {
	Filename   string
	BaseDir    string
	models     map[string]map[string]string
	recordings map[string]proxy.Recording
}

func (j *JsonFile) ReadFile() error {
	file, err := os.Open(j.Filename)
	if err != nil {
		return fmt.Errorf("error opening file %s :%v", j.Filename, err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file %s :%v", j.Filename, err)
	}

	if err := json.Unmarshal(data, &j.recordings); err != nil {
		return fmt.Errorf("error parsing recordings %s :%v", j.Filename, err)
	}

	scanner := Scanner{InputDir: j.BaseDir}
	models, err := scanner.ScanTags()
	if err != nil {
		return err
	}
	j.models = models
	return nil
}

func filterByMethod(rows []proxy.BodyRecords, method string) []proxy.BodyRecords {
	results := make([]proxy.BodyRecords, 0)
	for i := range rows {
		if rows[i].Method == method {
			results = append(results, rows[i])
		}
	}
	return results
}

func getFuncName(endpoint string) string {
	nm := make([]string, 0)
	for i := range strings.SplitSeq(endpoint, "/") {
		if len(i) > 3 {
			i = strings.ReplaceAll(i, "-", "")
			nm = append(nm, i)
		}
	}

	if len(nm) == 0 {
		return "NoName"
	}
	return strings.Join(nm, "")
}

func (j *JsonFile) detectFiles() error {
	if err := os.MkdirAll(testFileDir, 0o755); err != nil {
		return err
	}
	_, err := os.Stat("./gentest/main_test.go")
	if err != nil {
		e := os.WriteFile(filepath.Join(testFileDir, "main_test.go"), []byte(mainTest), 0o644)
		if e != nil {
			return e
		}
	}
	_, err = os.Stat("./gentest/testutils.go")
	if err != nil {
		e := os.WriteFile(filepath.Join(testFileDir, "testutils.go"), []byte(helperFile), 0o644)
		if e != nil {
			return e
		}
	}
	return nil
}

func (j *JsonFile) getFileName() string {
	fn := strings.Split(j.Filename, "-")
	if len(fn) > 1 {
		return strings.ReplaceAll(fn[len(fn)-1], ".json", "") + "_test.go"
	} else {
		return "no_name_test.go"
	}
}

func (j *JsonFile) genPostRun(endpoint string, rows []proxy.BodyRecords) string {
	if len(rows) == 0 {
		return ""
	}
	pkgPath, ok := j.models[endpoint]["folder"]
	if !ok {
		return ""
	}
	sname, ok := j.models[endpoint]["name"]
	if !ok {
		return ""
	}
	strct, ok := j.models[endpoint]["struct"]
	if !ok {
		return ""
	}
	funcName := getFuncName(endpoint)
	structGen := structgen.StructGenerator{BaseDir: j.BaseDir, PkgPath: pkgPath}
	var sb strings.Builder
	fmt.Fprintf(&sb, `t.Run("Create %s", func(t *testing.T) {`, funcName)
	if len(rows) == 1 {
		rawJson := make(map[string]any)
		if err := json.Unmarshal([]byte(rows[0].Body), &rawJson); err != nil {
			return ""
		}
		sm, ok := statusMap[rows[0].StatusCode]
		if !ok {
			sm = "http.StatusOK"
		}
		strctStr, err := structGen.MapField(strct, rawJson)
		if err != nil {
			slog.Error("error parsing struct during codegeneration", "struct", strct, "pkg", structGen.PkgPath, "err", err)
			sb.Reset()
			return ""
		}
		fmt.Fprintf(&sb, "payload := %s", sname)
		sb.WriteString(strctStr)
		sb.WriteString("\n\n")
		fmt.Fprintf(&sb, `resp := makeReq(t, app, http.MethodPost, "%s", payload)`, endpoint)
		sb.WriteString("\n")
		fmt.Fprintf(&sb, `require.Equal(t, %s, resp.StatusCode)`, sm)
		sb.WriteString("\n})")
	} else {
		fmt.Fprintf(&sb, "payloads := []struct{\nname string\npayload %s\nexpectedStatus int\n}{", sname)
		for i := range rows {
			rawJson := make(map[string]any)
			sm, ok := statusMap[rows[i].StatusCode]
			if !ok {
				sm = "http.StatusOK"
			}
			fmt.Fprintf(&sb, `{
				name: "%s %d",
				expectedStatus: %s,
				payload: %s `, funcName, i, sm, sname)
			if err := json.Unmarshal([]byte(rows[i].Body), &rawJson); err != nil {
				continue
			}
			strctStr, err := structGen.MapField(strct, rawJson)
			if err != nil {
				slog.Error("error parsing struct during codegeneration", "struct", strct, "pkg", structGen.PkgPath, "err", err)
				sb.Reset()
				return ""
			}
			sb.WriteString(strctStr)
			sb.WriteString(",},\n")
		}
		sb.WriteString("}\n")
		sb.WriteString("for _, pl := range payloads {\n")
		sb.WriteString("t.Run(pl.name, func(t *testing.T) {\n")
		fmt.Fprintf(&sb, `resp := makeReq(t, app, http.MethodPost, "%s", pl.payload)`, endpoint)
		sb.WriteString("\n")
		sb.WriteString("require.Equal(t, http.StatusOK, resp.StatusCode)\n")
		sb.WriteString("})\n")
		sb.WriteString("}\n")
		sb.WriteString("})")
	}
	return sb.String()
}

func (j *JsonFile) genGetRun(endpoint string, rows []proxy.BodyRecords) string {
	if len(rows) == 0 {
		return ""
	}
	funcName := getFuncName(endpoint)
	var sb strings.Builder
	fmt.Fprintf(&sb, `t.Run("Get %s", func(t *testing.T) {`, funcName)
	if len(rows) == 1 {
		sm, ok := statusMap[rows[0].StatusCode]
		if !ok {
			sm = "http.StatusOK"
		}
		fmt.Fprintf(&sb, `resp := makeReq(t, app, http.MethodGet, "%s", nil)`, endpoint)
		sb.WriteString("\n")
		fmt.Fprintf(&sb, `require.Equal(t, %s, resp.StatusCode)`, sm)
		sb.WriteString("\n})")
	} else {
		sb.WriteString("testCases := []struct{\nname string\npath string\nexpectedStatus int\n}{\n")
		for i := range rows {
			sm, ok := statusMap[rows[i].StatusCode]
			if !ok {
				sm = "http.StatusOK"
			}
			fmt.Fprintf(&sb, `{name: "%s %d", path: "%s", expectedStatus: %s},`, funcName, i, endpoint, sm)
			sb.WriteString("\n")
		}
		sb.WriteString("}\n")
		sb.WriteString("for _, tc := range testCases {\n")
		sb.WriteString("t.Run(tc.name, func(t *testing.T) {\n")
		sb.WriteString("resp := makeReq(t, app, http.MethodGet, tc.path, nil)\n")
		sb.WriteString("require.Equal(t, tc.expectedStatus, resp.StatusCode)\n")
		sb.WriteString("})\n")
		sb.WriteString("}\n")
		sb.WriteString("})")
	}
	return sb.String()
}

func (j *JsonFile) genDelRun(endpoint string, rows []proxy.BodyRecords) string {
	if len(rows) == 0 {
		return ""
	}
	funcName := getFuncName(endpoint)
	var sb strings.Builder
	fmt.Fprintf(&sb, `t.Run("Delete %s", func(t *testing.T) {`, funcName)
	if len(rows) == 1 {
		sm, ok := statusMap[rows[0].StatusCode]
		if !ok {
			sm = "http.StatusOK"
		}
		fmt.Fprintf(&sb, `resp := makeReq(t, app, http.MethodDelete, "%s", nil)`, endpoint)
		sb.WriteString("\n")
		fmt.Fprintf(&sb, `require.Equal(t, %s, resp.StatusCode)`, sm)
		sb.WriteString("\n})")
	} else {
		sb.WriteString("testCases := []struct{\nname string\npath string\nexpectedStatus int\n}{\n")
		for i := range rows {
			sm, ok := statusMap[rows[i].StatusCode]
			if !ok {
				sm = "http.StatusOK"
			}
			fmt.Fprintf(&sb, `{name: "%s %d", path: "%s", expectedStatus: %s},`, funcName, i, endpoint, sm)
			sb.WriteString("\n")
		}
		sb.WriteString("}\n")
		sb.WriteString("for _, tc := range testCases {\n")
		sb.WriteString("t.Run(tc.name, func(t *testing.T) {\n")
		sb.WriteString("resp := makeReq(t, app, http.MethodDelete, tc.path, nil)\n")
		sb.WriteString("require.Equal(t, tc.expectedStatus, resp.StatusCode)\n")
		sb.WriteString("})\n")
		sb.WriteString("}\n")
		sb.WriteString("})")
	}
	return sb.String()
}

func (j *JsonFile) genPutRun(endpoint string, rows []proxy.BodyRecords) string {
	if len(rows) == 0 {
		return ""
	}
	pkgPath, ok := j.models[endpoint]["folder"]
	if !ok {
		return ""
	}
	sname, ok := j.models[endpoint]["name"]
	if !ok {
		return ""
	}
	strct, ok := j.models[endpoint]["struct"]
	if !ok {
		return ""
	}
	funcName := getFuncName(endpoint)
	structGen := structgen.StructGenerator{BaseDir: j.BaseDir, PkgPath: pkgPath}
	var sb strings.Builder
	fmt.Fprintf(&sb, `t.Run("Update %s", func(t *testing.T) {`, funcName)
	if len(rows) == 1 {
		rawJson := make(map[string]any)
		if err := json.Unmarshal([]byte(rows[0].Body), &rawJson); err != nil {
			return ""
		}
		sm, ok := statusMap[rows[0].StatusCode]
		if !ok {
			sm = "http.StatusOK"
		}
		strctStr, err := structGen.MapField(strct, rawJson)
		if err != nil {
			slog.Error("error parsing struct during codegeneration", "struct", strct, "pkg", structGen.PkgPath, "err", err)
			sb.Reset()
			return ""
		}
		fmt.Fprintf(&sb, "payload := %s", sname)
		sb.WriteString(strctStr)
		sb.WriteString("\n\n")
		fmt.Fprintf(&sb, `resp := makeReq(t, app, http.MethodPut, "%s", payload)`, endpoint)
		sb.WriteString("\n")
		fmt.Fprintf(&sb, `require.Equal(t, %s, resp.StatusCode)`, sm)
		sb.WriteString("\n})")
	} else {

		fmt.Fprintf(&sb, "payloads := []struct{\nname string\npayload %s\nexpectedStatus int\n}{\n", sname)
		for i := range rows {
			rawJson := make(map[string]any)
			sm, ok := statusMap[rows[i].StatusCode]
			if !ok {
				sm = "http.StatusOK"
			}
			fmt.Fprintf(&sb, `{name: "%s %d",expectedStatus: %s, payload: %s `, funcName, i, sm, sname)
			if err := json.Unmarshal([]byte(rows[i].Body), &rawJson); err != nil {
				continue
			}
			strctStr, err := structGen.MapField(strct, rawJson)
			if err != nil {
				slog.Error("error parsing struct during codegeneration", "struct", strct, "pkg", structGen.PkgPath, "err", err)
				sb.Reset()
				return ""
			}
			sb.WriteString(strctStr)
			sb.WriteString(",},\n")
		}
		sb.WriteString("}\n")
		sb.WriteString("for _, pl := range payloads {\n")
		sb.WriteString("t.Run(pl.name, func(t *testing.T) {\n")
		fmt.Fprintf(&sb, `resp := makeReq(t, app, http.MethodPut, "%s", pl.payload)`, endpoint)
		sb.WriteString("\n")
		sb.WriteString("require.Equal(t, pl.expectedStatus, resp.StatusCode)\n")
		sb.WriteString("})\n")
		sb.WriteString("}\n")
		sb.WriteString("})")
	}
	return sb.String()
}

func (j *JsonFile) genTestFunction(ep string, rcrd proxy.Recording) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "func test%s(t *testing.T){\n", getFuncName(ep))
	sb.WriteString("\n\napp := setup()\n\n")
	sb.WriteString(j.genPostRun(ep, filterByMethod(rcrd.Body, "POST")))
	sb.WriteString("\n\n")
	sb.WriteString(j.genPutRun(ep, filterByMethod(rcrd.Body, "PUT")))
	sb.WriteString("\n\n")
	sb.WriteString(j.genGetRun(ep, filterByMethod(rcrd.Body, "GET")))
	sb.WriteString("\n\n")
	sb.WriteString(j.genDelRun(ep, filterByMethod(rcrd.Body, "DELETE")))
	sb.WriteString("\n\n}\n\n")

	return sb.String()
}

func (j *JsonFile) GenTest() error {
	err := j.detectFiles()
	if err != nil {
		return fmt.Errorf("error detecting files on %s, :%v", testFileDir, err)
	}
	fname := j.getFileName()
	var sb strings.Builder
	sb.WriteString(`package gentests
	import (
			"net/http"
			"testing"
			"github.com/stretchr/testify/require"
			)

	`)

	for key, item := range j.recordings {
		sb.WriteString(j.genTestFunction(key, item))
	}
	err = os.WriteFile(filepath.Join("./gentest", fname), []byte(sb.String()), 0o644)
	if err != nil {
		return fmt.Errorf("error writing test file %s :%v", fname, err)
	}

	return nil
}
