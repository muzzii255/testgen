package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var recordings = make(map[string]Recording)

type Recording struct {
	Body    []BodyRecords     `json:"bodyRecords"`
	Headers map[string]string `json:"headers"`
}

type BodyRecords struct {
	Path         string    `json:"path"`
	Body         string    `json:"body"`
	StatusCode   int       `json:"statusCode"`
	ResponseBody string    `json:"responseBody"`
	Timestamp    time.Time `json:"timestamp"`
	Method       string    `json:"method"`
}

type Recorder struct {
	targetURL *url.URL
	outputDir string
}

func NewRecorder(targetURL, outputDir string) (*Recorder, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, err
	}

	return &Recorder{
		targetURL: target,
		outputDir: outputDir,
	}, nil
}

func (r *Recorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reqBody, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	url := req.URL.Path
	url = normalizeUrl(url)
	recording, ok := recordings[url]
	if !ok {
		recording = Recording{
			Headers: make(map[string]string),
			Body:    make([]BodyRecords, 0),
		}

		for k, v := range req.Header {
			if len(v) > 0 {
				recording.Headers[k] = v[0]
			}
		}
		recordings[url] = recording
	}
	body := BodyRecords{
		Path:      cleanUrl(req.URL.Path),
		Method:    req.Method,
		Body:      string(reqBody),
		Timestamp: time.Now(),
	}
	proxy := httputil.NewSingleHostReverseProxy(r.targetURL)

	proxy.ModifyResponse = func(resp *http.Response) error {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

		body.StatusCode = resp.StatusCode
		body.ResponseBody = string(respBody)
		recording.Body = append(recording.Body, body)
		recordings[url] = recording
		// r.save(recordings)
		return r.save()
	}

	proxy.ServeHTTP(w, req)
}

func (r *Recorder) save() error {
	fileData := make(map[string]map[string]Recording)
	for key, item := range recordings {
		filename := fmt.Sprintf("%s-%s.json",
			time.Now().Format("2006-01-02"),
			cleanPath(key),
		)
		if _, exists := fileData[filename]; !exists {
			fileData[filename] = make(map[string]Recording)
		}

		fileData[filename][key] = item

	}
	for filename, data := range fileData {
		data, _ := json.MarshalIndent(data, "", "  ")
		return os.WriteFile(filepath.Join(r.outputDir, filename), data, 0o644)

	}
	return nil
}

func cleanPath(a string) string {
	res := strings.Split(a, "/")
	fn := make([]string, 0)
	for _, i := range res {
		o := strings.Split(i, "?")

		i = o[0]
		if len(i) > 3 && strings.ReplaceAll(i, " ", "") != "" {
			fn = append(fn, i)
		}
	}
	if len(fn) == 0 {
		fn = append(fn, res[len(res)-1])
	}

	fnn := strings.Join(fn, "_")
	if fnn == " " || fnn == "" {
		fnn = "noname"
	}
	return fnn
}

func cleanUrl(a string) string {
	b := strings.SplitSeq(a, "/")
	i := 0
	for o := range b {
		if _, err := strconv.Atoi(o); err == nil {
			if i == 0 {
				a = strings.Replace(a, o, ":id", 1)
			} else {
				a = strings.Replace(a, o, fmt.Sprintf(":id%d", i), 1)
			}
			i++
		}
	}
	return a
}

func normalizeUrl(a string) string {
	parts := strings.Split(a, "/")
	normalized := make([]string, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			continue
		}
		if _, err := strconv.Atoi(part); err == nil {
			continue
		}
		normalized = append(normalized, part)
	}

	return "/" + strings.Join(normalized, "/")
}
