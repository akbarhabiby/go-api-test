package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/akbarhabiby/go-api-test/helpers"
	"github.com/akbarhabiby/go-api-test/middlewares"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Response struct {
	IP         string      `json:"ip"`
	Time       time.Time   `json:"time"`
	RequestURI string      `json:"uri"`
	Method     string      `json:"method"`
	RawBody    any         `json:"rawBody"`
	FormBody   url.Values  `json:"formBody"`
	Headers    http.Header `json:"headers"`
}

const logPath = "/tmp/api-test-logs.json"
const PORT = "3000"

var (
	logFile *os.File
	err     error
	mu      sync.Mutex
)

func init() {
	runtime.GOMAXPROCS(1)
	logFile, err = os.OpenFile(logPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	bt, _ := io.ReadAll(logFile)
	if len(bt) < 1 {
		bt, _ := json.Marshal(make([]string, 0))
		logFile.Write(bt)
	}
}

func addLog(resp Response) {
	mu.Lock()
	defer mu.Unlock()
	logs := make([]Response, 0)
	logFile, _ := os.ReadFile(logPath)
	json.Unmarshal(logFile, &logs)
	logs = append([]Response{resp}, logs...)
	if len(logs) > 50 {
		logs = logs[:50]
	}
	bt, _ := json.MarshalIndent(logs, "", "  ")
	os.WriteFile(logPath, bt, 0644)
}

func main() {
	mux := http.DefaultServeMux
	mux.HandleFunc("/logs", logs)
	mux.HandleFunc("/", api)

	var handler http.Handler = mux

	// * Middlewares
	handler = middlewares.RateLimiter(handler)

	// * Server
	server := new(http.Server)
	server.Addr = fmt.Sprintf(":%s", PORT)
	// server.Handler = handler
	server.Handler = h2c.NewHandler(handler, &http2.Server{MaxConcurrentStreams: 500, MaxReadFrameSize: 1048576})

	fmt.Printf("api-test running on port %s\n", PORT)
	server.ListenAndServe()
}

func api(w http.ResponseWriter, req *http.Request) {
	// * Set Response Header and Status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// * Get Request Body
	var rawBody any
	contentType := req.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		req.ParseForm()
	} else if strings.Contains(contentType, "multipart/form-data") {
		req.ParseMultipartForm(16 << 20)
	} else {
		json.NewDecoder(req.Body).Decode(&rawBody)
	}

	// * Mapping Response
	resp := Response{
		IP:         helpers.GetRealIP(req),
		Time:       time.Now().UTC(),
		RequestURI: req.RequestURI,
		Method:     req.Method,
		RawBody:    rawBody,
		FormBody:   req.Form,
		Headers:    req.Header,
	}
	if req.MultipartForm != nil && len(req.MultipartForm.File) > 0 {
		for key, values := range req.MultipartForm.File {
			names := make([]string, 0)
			for _, file := range values {
				names = append(names, fmt.Sprintf("[FileName: [%s] Size: [%d]]", file.Filename, file.Size))
			}
			resp.FormBody.Add(key, strings.Join(names, ";"))
		}
	}
	go addLog(resp)
	json.NewEncoder(w).Encode(resp)
}

func logs(w http.ResponseWriter, req *http.Request) {
	// * Set Response Header and Status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	mu.Lock()
	defer mu.Unlock()
	logs := make([]Response, 0)
	logFile, _ := os.ReadFile(logPath)
	json.Unmarshal(logFile, &logs)
	json.NewEncoder(w).Encode(logs)
}
