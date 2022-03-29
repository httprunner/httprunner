package hrp

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/httprunner/httprunner/v4/hrp/internal/boomer"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

const jsonContentType = "application/json; encoding=utf-8"

func methods(h http.HandlerFunc, methods ...string) http.HandlerFunc {
	methodMap := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		methodMap[m] = struct{}{}
		// GET implies support for HEAD
		if m == "GET" {
			methodMap["HEAD"] = struct{}{}
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := methodMap[r.Method]; !ok {
			http.Error(w, fmt.Sprintf("method %s not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func parseBody(r *http.Request) (data map[string]interface{}, err error) {
	if r.Body == nil {
		return nil, nil
	}

	// Always set resp.Data to the incoming request body, in case we don't know
	// how to handle the content type
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		r.Body.Close()
		return nil, err
	}
	err = json.Unmarshal(body, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func writeResponse(w http.ResponseWriter, status int, contentType string, body []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeader(status)
	w.Write(body)
}

func writeJSON(w http.ResponseWriter, body []byte, status int) {
	writeResponse(w, status, jsonContentType, body)
}

type StartRequestBody struct {
	Worker       string `json:"worker"` // all
	SpawnCount   int64  `json:"spawn_count"`
	SpawnRate    int64  `json:"spawn_rate"`
	TestCasePath string `json:"testcase_path"`
}

type ServerCode int

// server response code
const (
	Success ServerCode = iota
	ParamsError
	ServerError
	StopError
)

// ServerStatus stores http response code and message
type ServerStatus struct {
	Code    ServerCode `json:"code"`
	Message string     `json:"message"`
}

var EnumAPIResponseSuccess = ServerStatus{
	Code:    Success,
	Message: "success",
}

func EnumAPIResponseParamError(errMsg string) ServerStatus {
	return ServerStatus{
		Code:    ParamsError,
		Message: errMsg,
	}
}

func EnumAPIResponseServerError(errMsg string) ServerStatus {
	return ServerStatus{
		Code:    ServerError,
		Message: errMsg,
	}
}

func EnumAPIResponseStopError(errMsg string) ServerStatus {
	return ServerStatus{
		Code:    StopError,
		Message: errMsg,
	}
}

func CustomAPIResponse(errCode ServerCode, errMsg string) ServerStatus {
	return ServerStatus{
		Code:    errCode,
		Message: errMsg,
	}
}

type RebalanceRequestBody struct {
	Worker       string `json:"worker"`
	SpawnCount   int64  `json:"spawn_count"`
	SpawnRate    int64  `json:"spawn_rate"`
	TestCasePath string `json:"testcase_path"`
}

type StopRequestBody struct {
	Worker string `json:"worker"`
}

type QuitRequestBody struct {
	Worker string `json:"worker"`
}

type CommonResponseBody struct {
	ServerStatus
}

type APIGetWorkersRequestBody struct {
	ID          string  `json:"id"`
	State       int32   `json:"state"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
}

type APIGetWorkersResponseBody struct {
	ServerStatus
	Data []boomer.WorkerNode `json:"data"`
}

type apiHandler struct {
	boomer *HRPBoomer
}

func (b *HRPBoomer) NewAPIHandler() *apiHandler {
	return &apiHandler{boomer: b}
}

// Index renders an HTML index page
func (api *apiHandler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' camo.githubusercontent.com")
	fmt.Fprintf(w, "Welcome to httprunner page!")
}

func (api *apiHandler) Start(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	args := r.URL.Query()
	for k, vs := range args {
		for _, v := range vs {
			data[k] = v
		}
	}
	var resp *CommonResponseBody
	err := api.boomer.Start(data)
	if err != nil {
		resp = &CommonResponseBody{
			ServerStatus: EnumAPIResponseServerError(err.Error()),
		}
	} else {
		resp = &CommonResponseBody{
			ServerStatus: EnumAPIResponseSuccess,
		}
	}
	body, _ := json.Marshal(resp)
	writeJSON(w, body, http.StatusOK)
}

func (api *apiHandler) Stop(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	args := r.URL.Query()
	for k, vs := range args {
		for _, v := range vs {
			data[k] = v
		}
	}

	api.boomer.Stop()
	resp := &CommonResponseBody{
		ServerStatus: EnumAPIResponseSuccess,
	}
	body, _ := json.Marshal(resp)
	writeJSON(w, body, http.StatusOK)
}

func (api *apiHandler) Quit(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	args := r.URL.Query()
	for k, vs := range args {
		for _, v := range vs {
			data[k] = v
		}
	}

	resp := &CommonResponseBody{
		ServerStatus: EnumAPIResponseSuccess,
	}
	body, _ := json.Marshal(resp)
	writeJSON(w, body, http.StatusOK)
	api.boomer.Quit()
}

func (api *apiHandler) ReBalance(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	args := r.URL.Query()
	for k, vs := range args {
		for _, v := range vs {
			data[k] = v
		}
	}
	var resp *CommonResponseBody
	err := api.boomer.ReBalance(data)
	if err != nil {
		resp = &CommonResponseBody{
			ServerStatus: EnumAPIResponseParamError(err.Error()),
		}
	} else {
		resp = &CommonResponseBody{
			ServerStatus: EnumAPIResponseSuccess,
		}
	}
	body, _ := json.Marshal(resp)
	writeJSON(w, body, http.StatusOK)
}

func (api *apiHandler) GetWorkersInfo(w http.ResponseWriter, r *http.Request) {
	resp := &APIGetWorkersResponseBody{
		ServerStatus: EnumAPIResponseSuccess,
		Data:         api.boomer.GetWorkersInfo(),
	}

	body, _ := json.Marshal(resp)
	writeJSON(w, body, http.StatusOK)
}

func (api *apiHandler) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", methods(api.Index, "GET"))
	mux.HandleFunc("/start", methods(api.Start, "GET"))
	mux.HandleFunc("/stop", methods(api.Stop, "GET"))
	mux.HandleFunc("/quit", methods(api.Quit, "GET"))
	mux.HandleFunc("/rebalance", methods(api.ReBalance, "GET"))
	mux.HandleFunc("/workers", methods(api.GetWorkersInfo, "GET"))

	return mux
}

func (apiHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func (b *HRPBoomer) StartServer() {
	h := b.NewAPIHandler()
	mux := h.Handler()

	server := &http.Server{
		Addr:    ":9771",
		Handler: mux,
	}

	go func() {
		<-b.GetCloseChan()
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal("shutdown server:", err)
		}
	}()

	log.Println("Starting HTTP server...")
	err := server.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			log.Print("server closed under request")
		} else {
			log.Fatal("server closed unexpected")
		}
	}
}
