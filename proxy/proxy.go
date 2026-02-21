package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

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
	targetURL  *url.URL
	outputDir  string
	mu         sync.RWMutex
	recordings map[string]Recording
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
		targetURL:  target,
		outputDir:  outputDir,
		recordings: make(map[string]Recording),
	}, nil
}

func (r *Recorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("error reading body", "path", req.URL.Path, "err", err)
		return
	}
	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	url := req.URL.Path
	url = normalizeURL(url)

	r.mu.Lock()
	recording, ok := r.recordings[url]
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
		r.recordings[url] = recording
	}
	r.mu.Unlock()

	body := BodyRecords{
		Path:      cleanURL(req.URL.Path),
		Method:    req.Method,
		Body:      string(reqBody),
		Timestamp: time.Now(),
	}
	proxy := httputil.NewSingleHostReverseProxy(r.targetURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		slog.Error("proxy error", "path", req.URL.Path, "err", err)
		w.WriteHeader(http.StatusBadGateway)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
		body.StatusCode = resp.StatusCode
		body.ResponseBody = string(respBody)
		recording.Body = append(recording.Body, body)
		r.mu.Lock()
		r.recordings[url] = recording
		r.mu.Unlock()
		return nil
	}
	slog.Info("request proxied", "path", req.URL.Path)
	proxy.ServeHTTP(w, req)
}

func (r *Recorder) Save() {
	fileData := make(map[string]map[string]Recording)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for key, item := range r.recordings {
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
		mdata, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			slog.Error("error marshalling recording", "err", err)
			continue
		}
		err = os.WriteFile(filepath.Join(r.outputDir, filename), mdata, 0o644)
		if err != nil {
			slog.Error("error writing file", "file", filename, "err", err)
			continue
		}
	}
}

func cleanPath(url string) string {
	res := strings.Split(url, "/")
	urlChunks := make([]string, 0)
	for _, i := range res {
		o := strings.Split(i, "?")

		i = o[0]
		i = strings.Trim(i, "")
		if len(i) > 3 && i != "" {
			urlChunks = append(urlChunks, i)
		}
	}
	if len(urlChunks) == 0 {
		urlChunks = append(urlChunks, res[len(res)-1])
	}

	cleanPath := strings.Join(urlChunks, "_")
	if cleanPath == "" {
		cleanPath = "noname"
	}
	return cleanPath
}

func cleanURL(url string) string {
	b := strings.SplitSeq(url, "/")
	for o := range b {
		if _, err := strconv.Atoi(o); err == nil {
			url = strings.ReplaceAll(url, o, ":id")
		}
	}
	return url
}

func normalizeURL(url string) string {
	parts := strings.Split(url, "/")
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
