package hrp

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/json"
)

func newWSSession() *wsSession {
	return &wsSession{
		wsConnMap:         make(map[string]*websocket.Conn),
		pongResponseChan:  make(chan string, 1),
		closeResponseChan: make(chan *wsCloseRespObject, 1),
	}
}

type wsSession struct {
	wsConnMap         map[string]*websocket.Conn // save all websocket connections
	pongResponseChan  chan string                // channel used to receive pong response message
	closeResponseChan chan *wsCloseRespObject    // channel used to receive close response message
}

const (
	wsOpen         WSActionType = "open"
	wsPing         WSActionType = "ping"
	wsWriteAndRead WSActionType = "wr"
	wsRead         WSActionType = "r"
	wsWrite        WSActionType = "w"
	wsClose        WSActionType = "close"
)

const (
	defaultTimeout     = 30000                        // default timeout 30 seconds for open, read and close
	defaultCloseStatus = websocket.CloseNormalClosure // default normal close status
	defaultWriteWait   = 5 * time.Second              // default timeout 5 seconds for writing control message
)

type WSActionType string

func (at WSActionType) toString() string {
	switch at {
	case wsOpen:
		return "open new connection"
	case wsPing:
		return "send ping and expect pong"
	case wsWriteAndRead:
		return "write and read"
	case wsRead:
		return "read only"
	case wsWrite:
		return "write only"
	case wsClose:
		return "close current connection"
	default:
		return "unexpected action type"
	}
}

type MessageType int

func (mt MessageType) toString() string {
	switch mt {
	case websocket.TextMessage:
		return "text"
	case websocket.BinaryMessage:
		return "binary"
	case websocket.PingMessage:
		return "ping"
	case websocket.PongMessage:
		return "pong"
	case websocket.CloseMessage:
		return "close"
	case 0:
		return "continuation"
	case -1:
		return "no frame"
	default:
		return "unexpected message type"
	}
}

// WebSocketConfig TODO: support reconnection ability
type WebSocketConfig struct {
	ReconnectionTimes    int64 `json:"reconnection_times,omitempty" yaml:"reconnection_times,omitempty"`       // maximum reconnection times when the connection closed by remote server
	ReconnectionInterval int64 `json:"reconnection_interval,omitempty" yaml:"reconnection_interval,omitempty"` // interval between each reconnection in milliseconds
	MaxMessageSize       int64 `json:"max_message_size,omitempty" yaml:"max_message_size,omitempty"`           // maximum message size during writing/reading
}

const (
	defaultReconnectionTimes    = 0
	defaultReconnectionInterval = 5000
	defaultMaxMessageSize       = 0
)

// checkWebSocket validates the websocket configuration parameters
func (wsConfig *WebSocketConfig) checkWebSocket() {
	if wsConfig == nil {
		return
	}
	if wsConfig.ReconnectionTimes <= 0 {
		wsConfig.ReconnectionTimes = defaultReconnectionTimes
	}
	if wsConfig.ReconnectionInterval <= 0 {
		wsConfig.ReconnectionInterval = defaultReconnectionInterval
	}
	if wsConfig.MaxMessageSize <= 0 {
		wsConfig.MaxMessageSize = defaultMaxMessageSize
	}
}

// StepWebSocket implements IStep interface.
type StepWebSocket struct {
	StepConfig
	WebSocket *WebSocketAction `json:"websocket,omitempty" yaml:"websocket,omitempty"`
}

func (s *StepWebSocket) Name() string {
	if s.StepName != "" {
		return s.StepName
	}
	return fmt.Sprintf("%s %s", s.WebSocket.Type, s.WebSocket.URL)
}

func (s *StepWebSocket) Type() StepType {
	return StepType(fmt.Sprintf("websocket-%v", s.WebSocket.Type))
}

func (s *StepWebSocket) Config() *StepConfig {
	return &s.StepConfig
}

func (s *StepWebSocket) Run(r *SessionRunner) (*StepResult, error) {
	return runStepWebSocket(r, s)
}

func (s *StepWebSocket) withUrl(url ...string) *StepWebSocket {
	if len(url) == 0 {
		return s
	}
	if len(url) > 1 {
		log.Warn().Msg("too many WebSocket step URL specified, using first URL")
	}
	s.WebSocket.URL = url[0]
	return s
}

func (s *StepWebSocket) OpenConnection(url ...string) *StepWebSocket {
	s.WebSocket.Type = wsOpen
	return s.withUrl(url...)
}

func (s *StepWebSocket) PingPong(url ...string) *StepWebSocket {
	s.WebSocket.Type = wsPing
	return s.withUrl(url...)
}

func (s *StepWebSocket) WriteAndRead(url ...string) *StepWebSocket {
	s.WebSocket.Type = wsWriteAndRead
	return s.withUrl(url...)
}

func (s *StepWebSocket) Read(url ...string) *StepWebSocket {
	s.WebSocket.Type = wsRead
	return s.withUrl(url...)
}

func (s *StepWebSocket) Write(url ...string) *StepWebSocket {
	s.WebSocket.Type = wsWrite
	return s.withUrl(url...)
}

func (s *StepWebSocket) CloseConnection(url ...string) *StepWebSocket {
	s.WebSocket.Type = wsClose
	return s.withUrl(url...)
}

func (s *StepWebSocket) WithParams(params map[string]interface{}) *StepWebSocket {
	s.WebSocket.Params = params
	return s
}

func (s *StepWebSocket) WithHeaders(headers map[string]string) *StepWebSocket {
	s.WebSocket.Headers = headers
	return s
}

func (s *StepWebSocket) NewConnection() *StepWebSocket {
	s.WebSocket.NewConnection = true
	return s
}

func (s *StepWebSocket) WithTextMessage(message interface{}) *StepWebSocket {
	s.WebSocket.TextMessage = message
	return s
}

func (s *StepWebSocket) WithBinaryMessage(message interface{}) *StepWebSocket {
	s.WebSocket.BinaryMessage = message
	return s
}

func (s *StepWebSocket) WithTimeout(timeout int64) *StepWebSocket {
	s.WebSocket.Timeout = timeout
	return s
}

func (s *StepWebSocket) WithCloseStatus(closeStatus int64) *StepWebSocket {
	s.WebSocket.CloseStatusCode = closeStatus
	return s
}

// Validate switches to step validation.
func (s *StepWebSocket) Validate() *StepWebSocketValidation {
	return &StepWebSocketValidation{
		StepWebSocket: s,
	}
}

// Extract switches to step extraction.
func (s *StepWebSocket) Extract() *StepWebSocketExtraction {
	s.StepConfig.Extract = make(map[string]string)
	return &StepWebSocketExtraction{
		StepWebSocket: s,
	}
}

// StepWebSocketExtraction implements IStep interface.
type StepWebSocketExtraction struct {
	*StepWebSocket
}

// WithJmesPath sets the JMESPath expression to extract from the response.
func (s *StepWebSocketExtraction) WithJmesPath(jmesPath string, varName string) *StepWebSocketExtraction {
	s.StepConfig.Extract[varName] = jmesPath
	return s
}

// Validate switches to step validation.
func (s *StepWebSocketExtraction) Validate() *StepWebSocketValidation {
	return &StepWebSocketValidation{
		StepWebSocket: s.StepWebSocket,
	}
}

func (s *StepWebSocketExtraction) Name() string {
	return s.StepName
}

func (s *StepWebSocketExtraction) Type() StepType {
	stepType := StepType(fmt.Sprintf("websocket-%v", s.WebSocket.Type))
	return stepType + stepTypeSuffixExtraction
}

func (s *StepWebSocketExtraction) Struct() *StepConfig {
	return &s.StepConfig
}

func (s *StepWebSocketExtraction) Run(r *SessionRunner) (*StepResult, error) {
	if s.WebSocket != nil {
		return runStepWebSocket(r, s.StepWebSocket)
	}
	return nil, errors.New("unexpected protocol type")
}

// StepWebSocketValidation implements IStep interface.
type StepWebSocketValidation struct {
	*StepWebSocket
}

func (s *StepWebSocketValidation) Name() string {
	if s.StepName != "" {
		return s.StepName
	}
	return fmt.Sprintf("%s %s", s.WebSocket.Type, s.WebSocket.URL)
}

func (s *StepWebSocketValidation) Type() StepType {
	stepType := StepType(fmt.Sprintf("websocket-%v", s.WebSocket.Type))
	return stepType + stepTypeSuffixValidation
}

func (s *StepWebSocketValidation) Config() *StepConfig {
	return &s.StepConfig
}

func (s *StepWebSocketValidation) Run(r *SessionRunner) (*StepResult, error) {
	if s.WebSocket != nil {
		return runStepWebSocket(r, s.StepWebSocket)
	}
	return nil, errors.New("unexpected protocol type")
}

func (s *StepWebSocketValidation) AssertEqual(jmesPath string, expected interface{}, msg string) *StepWebSocketValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "equals",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepWebSocketValidation) AssertEqualFold(jmesPath string, expected interface{}, msg string) *StepWebSocketValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "equal_fold",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepWebSocketValidation) AssertLengthEqual(jmesPath string, expected interface{}, msg string) *StepWebSocketValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_equals",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepWebSocketValidation) AssertContains(jmesPath string, expected interface{}, msg string) *StepWebSocketValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "contains",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

type WebSocketAction struct {
	Type            WSActionType           `json:"type" yaml:"type"`
	URL             string                 `json:"url" yaml:"url"`
	Params          map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
	Headers         map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	NewConnection   bool                   `json:"new_connection,omitempty" yaml:"new_connection,omitempty"` // TODO support
	TextMessage     interface{}            `json:"text,omitempty" yaml:"text,omitempty"`
	BinaryMessage   interface{}            `json:"binary,omitempty" yaml:"binary,omitempty"`
	CloseStatusCode int64                  `json:"close_status,omitempty" yaml:"close_status,omitempty"`
	Timeout         int64                  `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func (w *WebSocketAction) GetTimeout() int64 {
	if w.Timeout <= 0 {
		return defaultTimeout
	}
	return w.Timeout
}

func (w *WebSocketAction) GetCloseStatusCode() int64 {
	// close status code range: [1000, 4999]. ref: https://datatracker.ietf.org/doc/html/rfc6455#section-11.7
	if w.CloseStatusCode < 1000 || w.CloseStatusCode > 4999 {
		return defaultCloseStatus
	}
	return w.CloseStatusCode
}

func runStepWebSocket(r *SessionRunner, step IStep) (stepResult *StepResult, err error) {
	stepWebSocket := step.(*StepWebSocket)
	webSocket := stepWebSocket.WebSocket
	variables := stepWebSocket.Variables
	start := time.Now()
	stepResult = &StepResult{
		Name:        step.Name(),
		StepType:    step.Type(),
		Success:     false,
		ContentSize: 0,
		StartTime:   start.UnixMilli(),
	}

	defer func() {
		// update testcase summary
		if err != nil {
			stepResult.Attachments = err.Error()
		}
		stepResult.Elapsed = time.Since(start).Milliseconds()
	}()

	sessionData := &SessionData{
		ReqResps: &ReqResps{},
	}
	parser := r.caseRunner.parser
	config := r.caseRunner.Config.Get()

	dummyReq := &Request{
		URL:     webSocket.URL,
		Params:  webSocket.Params,
		Headers: webSocket.Headers,
	}
	rb := newRequestBuilder(parser, config, dummyReq)

	err = rb.prepareUrlParams(variables)
	if err != nil {
		return
	}

	err = rb.prepareHeaders(variables)
	if err != nil {
		return
	}
	parsedURL := rb.req.URL.String()
	parsedHeader := rb.req.Header

	// add request object to step variables, could be used in setup hooks
	variables["hrp_step_name"] = step.Name
	variables["hrp_step_request"] = rb.requestMap

	// deal with setup hooks
	for _, setupHook := range stepWebSocket.SetupHooks {
		_, err = parser.Parse(setupHook, variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "run setup hooks failed")
		}
	}

	var resp interface{}

	// do websocket action
	if r.caseRunner.hrpRunner.requestsLogOn {
		fmt.Printf("-------------------- websocket action: %v --------------------\n", webSocket.Type.toString())
	}
	switch webSocket.Type {
	case wsOpen:
		log.Info().Int64("timeout(ms)", webSocket.GetTimeout()).Str("url", parsedURL).Msg("open websocket connection")
		// use the current websocket connection if existed
		if getWsClient(r, parsedURL) != nil {
			break
		}
		resp, err = openWithTimeout(parsedURL, parsedHeader, r, stepWebSocket)
		if err != nil {
			return stepResult, errors.Wrap(err, "open connection failed")
		}
	case wsPing:
		log.Info().Int64("timeout(ms)", webSocket.GetTimeout()).Str("url", parsedURL).Msg("send ping and expect pong")
		err = writeWebSocket(parsedURL, r, stepWebSocket, stepWebSocket.Variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "send ping message failed")
		}
		timer := time.NewTimer(time.Duration(webSocket.GetTimeout()) * time.Millisecond)
		// asynchronous receiving pong message with timeout
		go func() {
			select {
			case <-timer.C:
				timer.Stop()
				log.Warn().Msg("pong timeout")
			case pongResponse := <-r.ws.pongResponseChan:
				resp = pongResponse
				log.Info().Msg("pong message arrives")
			}
		}()
	case wsWriteAndRead:
		log.Info().Int64("timeout(ms)", webSocket.GetTimeout()).Str("url", parsedURL).Msg("write a message and read response")
		err = writeWebSocket(parsedURL, r, stepWebSocket, variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "write message failed")
		}
		resp, err = readMessageWithTimeout(parsedURL, r, stepWebSocket)
		if err != nil {
			return stepResult, errors.Wrap(err, "read message failed")
		}
	case wsRead:
		log.Info().Int64("timeout(ms)", webSocket.GetTimeout()).Str("url", parsedURL).Msg("read only")
		resp, err = readMessageWithTimeout(parsedURL, r, stepWebSocket)
		if err != nil {
			return stepResult, errors.Wrap(err, "read message failed")
		}
	case wsWrite:
		log.Info().Str("url", parsedURL).Msg("write only")
		err = writeWebSocket(parsedURL, r, stepWebSocket, variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "write message failed")
		}
	case wsClose:
		log.Info().Int64("timeout(ms)", webSocket.GetTimeout()).Str("url", parsedURL).Msg("close webSocket connection")
		resp, err = closeWithTimeout(parsedURL, r, stepWebSocket, variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "close connection failed")
		}
	default:
		return stepResult, errors.Errorf("unexpected websocket frame type: %v", webSocket.Type)
	}
	if r.caseRunner.hrpRunner.requestsLogOn {
		err = printWebSocketResponse(resp)
		if err != nil {
			return stepResult, errors.Wrap(err, "print response failed")
		}
	}

	respObj, err := getResponseObject(r.caseRunner.hrpRunner.t, r.caseRunner.parser, resp)
	if err != nil {
		err = errors.Wrap(err, "get response object error")
		return
	}

	if respObj != nil {
		// add response object to step variables, could be used in teardown hooks
		variables["hrp_step_response"] = respObj.respObjMeta
	}

	// deal with teardown hooks
	for _, teardownHook := range stepWebSocket.TeardownHooks {
		_, err = parser.Parse(teardownHook, variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "run teardown hooks failed")
		}
	}

	if respObj != nil {
		sessionData.ReqResps.Request = rb.requestMap
		sessionData.ReqResps.Response = builtin.FormatResponse(respObj.respObjMeta)

		// extract variables from response
		extractors := stepWebSocket.StepConfig.Extract
		extractMapping := respObj.Extract(extractors, variables)
		stepResult.ExportVars = extractMapping

		// override step variables with extracted variables
		variables = mergeVariables(variables, extractMapping)

		// validate response
		err = respObj.Validate(stepWebSocket.Validators, variables)
		sessionData.Validators = respObj.validationResults
		if err == nil {
			stepResult.Success = true
		}
		stepResult.ContentSize = getContentSize(resp)
		stepResult.Data = sessionData
	} else {
		stepResult.Success = true
	}

	return stepResult, nil
}

func getWsClient(r *SessionRunner, url string) *websocket.Conn {
	if client, ok := r.ws.wsConnMap[url]; ok {
		return client
	}

	return nil
}

func printWebSocketResponse(resp interface{}) error {
	if resp == nil {
		fmt.Println("(response body is empty in this step)")
		fmt.Println("----------------------------------------")
		return nil
	}
	if httpResp, ok := resp.(*http.Response); ok {
		return printResponse(httpResp)
	}
	fmt.Println("==================== response ====================")
	switch r := resp.(type) {
	case *wsReadRespObject:
		if r.messageType == websocket.TextMessage {
			fmt.Printf("message type: %v\r\nmessage: %s\r\n", MessageType(r.messageType).toString(), r.Message)
		} else if r.messageType == websocket.BinaryMessage {
			fmt.Printf("message type: %v\r\nmessage: %v\r\ncorresponding string: %s\r\n", MessageType(r.messageType).toString(), r.Message, r.Message)
		} else {
			return errors.New("unexpected response type")
		}
	case *wsCloseRespObject:
		fmt.Printf("close status code: %v\r\nmessage: %v\r\n", r.StatusCode, r.Text)
	case string:
		fmt.Println(r)
	default:
		return errors.New("unexpected response type")
	}
	fmt.Println("----------------------------------------")
	return nil
}

func openWithTimeout(urlStr string, requestHeader http.Header, r *SessionRunner, step *StepWebSocket) (*http.Response, error) {
	openResponseChan := make(chan *http.Response)
	errorChan := make(chan error)
	go func() {
		conn, resp, err := r.caseRunner.hrpRunner.wsDialer.Dial(urlStr, requestHeader)
		if err != nil {
			errorChan <- errors.Wrap(err, "dial tcp failed")
			return
		}
		// handshake end here
		defer resp.Body.Close()

		// the following handlers handle and transport control message from server
		conn.SetPongHandler(func(appData string) error {
			r.ws.pongResponseChan <- appData
			return nil
		})
		conn.SetCloseHandler(func(code int, text string) error {
			message := websocket.FormatCloseMessage(code, "")
			conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(defaultWriteWait))
			select {
			case r.ws.closeResponseChan <- &wsCloseRespObject{
				StatusCode: code,
				Text:       text,
			}:
			default:
				log.Warn().Msg("close response channel is block, drop the response")
			}

			return nil
		})

		r.ws.wsConnMap[urlStr] = conn
		openResponseChan <- resp
	}()

	timer := time.NewTimer(time.Duration(step.WebSocket.GetTimeout()) * time.Millisecond)
	select {
	case <-timer.C:
		timer.Stop()
		return nil, errors.New("open timeout")
	case err := <-errorChan:
		return nil, err
	case openResponse := <-openResponseChan:
		return openResponse, nil
	}
}

func readMessageWithTimeout(urlString string, r *SessionRunner, step *StepWebSocket) (*wsReadRespObject, error) {
	wsConn := getWsClient(r, urlString)
	if wsConn == nil {
		return nil, errors.New("try to use existing connection, but there is no connection")
	}
	readResponseChan := make(chan *wsReadRespObject)
	errorChan := make(chan error)
	go func() {
		messageType, message, err := wsConn.ReadMessage()
		if err != nil {
			errorChan <- err
		} else {
			readResponseChan <- &wsReadRespObject{
				messageType: messageType,
				Message:     message,
			}
		}
	}()
	timer := time.NewTimer(time.Duration(step.WebSocket.GetTimeout()) * time.Millisecond)
	select {
	case <-timer.C:
		timer.Stop()
		return nil, errors.New("read timeout")
	case err := <-errorChan:
		return nil, err
	case readResult := <-readResponseChan:
		return readResult, nil
	}
}

func writeWebSocket(urlString string, r *SessionRunner, step *StepWebSocket, stepVariables map[string]interface{}) error {
	wsConn := getWsClient(r, urlString)
	if wsConn == nil {
		return errors.New("try to use existing connection, but there is no connection")
	}
	// check priority: text message > binary message
	if step.WebSocket.TextMessage != nil {
		parsedMessage, parseErr := r.caseRunner.parser.Parse(step.WebSocket.TextMessage, stepVariables)
		if parseErr != nil {
			return parseErr
		}
		writeErr := writeWithType(wsConn, step, websocket.TextMessage, parsedMessage)
		if writeErr != nil {
			return writeErr
		}
	} else if step.WebSocket.BinaryMessage != nil {
		parsedMessage, parseErr := r.caseRunner.parser.Parse(step.WebSocket.BinaryMessage, stepVariables)
		if parseErr != nil {
			return parseErr
		}
		writeErr := writeWithType(wsConn, step, websocket.BinaryMessage, parsedMessage)
		if writeErr != nil {
			return writeErr
		}
	} else {
		log.Info().Msg("step with empty message")
		err := writeWithAction(wsConn, step, websocket.BinaryMessage, []byte{})
		if err != nil {
			return err
		}
	}
	return nil
}

func writeWithType(c *websocket.Conn, step *StepWebSocket, messageType int, message interface{}) error {
	if message == nil {
		return nil
	}
	if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
		return errors.New("unexpected message type")
	}
	switch msg := message.(type) {
	case []byte:
		return writeWithAction(c, step, messageType, msg)
	case string:
		return writeWithAction(c, step, messageType, []byte(msg))
	case bytes.Buffer:
		return writeWithAction(c, step, messageType, msg.Bytes())
	default:
		msgBytes, _ := json.Marshal(msg)
		return writeWithAction(c, step, messageType, msgBytes)
	}
}

func writeWithAction(c *websocket.Conn, step *StepWebSocket, messageType int, message []byte) error {
	switch step.WebSocket.Type {
	case wsPing:
		return c.WriteControl(websocket.PingMessage, message, time.Now().Add(defaultWriteWait))
	case wsClose:
		closeMessage := websocket.FormatCloseMessage(int(step.WebSocket.GetCloseStatusCode()), string(message))
		return c.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(defaultWriteWait))
	default:
		return c.WriteMessage(messageType, message)
	}
}

func closeWithTimeout(urlString string, r *SessionRunner, step *StepWebSocket, stepVariables map[string]interface{}) (*wsCloseRespObject, error) {
	wsConn := getWsClient(r, urlString)
	if wsConn == nil {
		return nil, errors.New("no connection needs to be closed")
	}
	errorChan := make(chan error)
	go func() {
		err := writeWebSocket(urlString, r, step, stepVariables)
		if err != nil {
			errorChan <- errors.Wrap(err, "send close message failed")
			return
		}
		// discard redundant message left in the connection before close
		var mt int
		var message []byte
		var readErr error
		for readErr == nil {
			mt, message, readErr = wsConn.ReadMessage()
			if readErr == nil {
				log.Info().
					Str("type", MessageType(mt).toString()).
					Str("msg", string(message)).
					Msg("discard redundant message")
				continue
			}
			if e, ok := readErr.(*websocket.CloseError); !ok {
				errorChan <- errors.Wrap(e, "read message failed")
				return
			}
		}
		// r.wsConn.Close() will be called at the end of current session, so no need to Close here
		log.Info().Str("msg", readErr.Error()).Msg("connection closed")
	}()
	timer := time.NewTimer(time.Duration(step.WebSocket.GetTimeout()) * time.Millisecond)
	select {
	case <-timer.C:
		timer.Stop()
		return nil, errors.New("close timeout")
	case err := <-errorChan:
		return nil, err
	case closeResult := <-r.ws.closeResponseChan:
		return closeResult, nil
	}
}

func getResponseObject(t *testing.T, parser *Parser, resp interface{}) (*responseObject, error) {
	// response could be nil for ping and write case
	if resp == nil {
		return nil, nil
	}
	switch r := resp.(type) {
	case *http.Response:
		return newHttpResponseObject(t, parser, r)
	case *wsReadRespObject:
		return newWsReadResponseObject(t, parser, r)
	case *wsCloseRespObject:
		return newWsCloseResponseObject(t, parser, r)
	default:
		return nil, errors.New("unxexpected reponse type")
	}
}

func getContentSize(resp interface{}) int64 {
	switch r := resp.(type) {
	case *http.Response:
		return r.ContentLength
	case *wsReadRespObject:
		return int64(unsafe.Sizeof(r.Message))
	case *wsCloseRespObject:
		return int64(unsafe.Sizeof(r.Text))
	default:
		return -1
	}
}

func (r *SessionRunner) ReleaseResources() {
	// close websocket connections
	for _, wsConn := range r.ws.wsConnMap {
		if wsConn != nil {
			log.Info().Str("testcase", r.caseRunner.Config.Get().Name).Msg("websocket disconnected")
			err := wsConn.Close()
			if err != nil {
				log.Error().Err(err).Msg("websocket disconnection failed")
			}
		}
	}
}
