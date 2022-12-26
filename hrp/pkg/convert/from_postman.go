package convert

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

// ==================== model definition starts here ====================

/*
Postman Collection format reference:
https://schema.postman.com/json/collection/v2.0.0/collection.json
https://schema.postman.com/json/collection/v2.1.0/collection.json
*/

// CasePostman represents the postman exported file
type CasePostman struct {
	Info  TInfo   `json:"info"`
	Items []TItem `json:"item"`
}

// TInfo gives information about the collection
type TInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
}

// TItem contains the detail information of request and expected responses
// item could be defined recursively
type TItem struct {
	Items     []TItem     `json:"item"`
	Name      string      `json:"name"`
	Request   TRequest    `json:"request"`
	Responses []TResponse `json:"response"`
}

type TRequest struct {
	Method      string   `json:"method"`
	Headers     []TField `json:"header"`
	Body        TBody    `json:"body"`
	URL         TUrl     `json:"url"`
	Description string   `json:"description"`
}

type TResponse struct {
	Name            string   `json:"name"`
	OriginalRequest TRequest `json:"originalRequest"`
	Status          string   `json:"status"`
	Code            int      `json:"code"`
	Headers         []TField `json:"headers"`
	Body            string   `json:"body"`
}

type TUrl struct {
	Raw         string   `json:"raw"`
	Protocol    string   `json:"protocol"`
	Path        []string `json:"path"`
	Description string   `json:"description"`
	Query       []TField `json:"query"`
	Variable    []TField `json:"variable"`
}

type TField struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Src         string `json:"src"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Disabled    bool   `json:"disabled"`
	Enable      bool   `json:"enable"`
}

type TBody struct {
	Mode       string      `json:"mode"`
	FormData   []TField    `json:"formdata"`
	URLEncoded []TField    `json:"urlencoded"`
	Raw        string      `json:"raw"`
	Disabled   bool        `json:"disabled"`
	Options    interface{} `json:"options"`
}

// ==================== model definition ends here ====================

const (
	enumBodyRaw        = "raw"
	enumBodyUrlEncoded = "urlencoded"
	enumBodyFormData   = "formdata"
	enumBodyFile       = "file"
	enumBodyGraphQL    = "graphql"
)

const (
	enumFieldTypeText = "text"
	enumFieldTypeFile = "file"
)

var contentTypeMap = map[string]string{
	"text":       "text/plain",
	"javascript": "application/javascript",
	"json":       "application/json",
	"html":       "text/html",
	"xml":        "application/xml",
}

func LoadPostmanCase(path string) (*hrp.TCase, error) {
	log.Info().Str("path", path).Msg("load postman case file")
	casePostman, err := loadCasePostman(path)
	if err != nil {
		return nil, err
	}

	// convert to TCase format
	return casePostman.ToTCase()
}

func loadCasePostman(path string) (*CasePostman, error) {
	casePostman := new(CasePostman)
	err := builtin.LoadFile(path, casePostman)
	if err != nil {
		return nil, errors.Wrap(err, "load postman file failed")
	}
	if casePostman.Items == nil {
		return nil, errors.New("invalid postman case file, missing items")
	}

	return casePostman, nil
}

func (c *CasePostman) ToTCase() (*hrp.TCase, error) {
	teststeps, err := c.prepareTestSteps()
	if err != nil {
		return nil, err
	}
	tCase := &hrp.TCase{
		Config:    c.prepareConfig(),
		TestSteps: teststeps,
	}
	err = tCase.MakeCompat()
	if err != nil {
		return nil, err
	}
	return tCase, nil
}

func (c *CasePostman) prepareConfig() *hrp.TConfig {
	return hrp.NewConfig(c.Info.Name).
		SetVerifySSL(false)
}

func (c *CasePostman) prepareTestSteps() ([]*hrp.TStep, error) {
	// recursively convert collection items into a list
	var itemList []TItem
	for _, item := range c.Items {
		extractItemList(item, &itemList)
	}

	var steps []*hrp.TStep
	for _, item := range itemList {
		step, err := c.prepareTestStep(&item)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func extractItemList(item TItem, itemList *[]TItem) {
	// current item contains no other items and request is not empty
	if len(item.Items) == 0 {
		if !reflect.DeepEqual(item.Request, TRequest{}) {
			*itemList = append(*itemList, item)
		}
		return
	}

	// look up all items inside
	for _, i := range item.Items {
		// append item name
		i.Name = fmt.Sprintf("%s - %s", item.Name, i.Name)
		extractItemList(i, itemList)
	}
}

func (c *CasePostman) prepareTestStep(item *TItem) (*hrp.TStep, error) {
	log.Info().
		Str("method", item.Request.Method).
		Str("url", item.Request.URL.Raw).
		Msg("convert teststep")

	step := &stepFromPostman{
		TStep: hrp.TStep{
			Request:    &hrp.Request{},
			Validators: make([]interface{}, 0),
		},
	}
	if err := step.makeRequestName(item); err != nil {
		return nil, err
	}
	if err := step.makeRequestMethod(item); err != nil {
		return nil, err
	}
	if err := step.makeRequestURL(item); err != nil {
		return nil, err
	}
	if err := step.makeRequestParams(item); err != nil {
		return nil, err
	}
	if err := step.makeRequestHeaders(item); err != nil {
		return nil, err
	}
	if err := step.makeRequestCookies(item); err != nil {
		return nil, err
	}
	if err := step.makeRequestBody(item); err != nil {
		return nil, err
	}
	return &step.TStep, nil
}

type stepFromPostman struct {
	hrp.TStep
}

// makeRequestName indicates the step name the same as item name
func (s *stepFromPostman) makeRequestName(item *TItem) error {
	s.Name = item.Name
	return nil
}

func (s *stepFromPostman) makeRequestMethod(item *TItem) error {
	s.Request.Method = hrp.HTTPMethod(item.Request.Method)
	return nil
}

func (s *stepFromPostman) makeRequestURL(item *TItem) error {
	rawUrl := item.Request.URL.Raw
	// parse path variables like ":path" in https://postman-echo.com/:path?k1=v1&k2=v2
	for _, field := range item.Request.URL.Variable {
		pathVar := ":" + field.Key
		rawUrl = strings.Replace(rawUrl, pathVar, field.Value, -1)
	}
	u, err := url.Parse(rawUrl)
	if err != nil {
		return errors.Wrap(err, "parse URL error")
	}
	s.Request.URL = fmt.Sprintf("%s://%s", u.Scheme, u.Host+u.Path)
	return nil
}

func (s *stepFromPostman) makeRequestParams(item *TItem) error {
	s.Request.Params = make(map[string]interface{})
	for _, field := range item.Request.URL.Query {
		if field.Disabled {
			continue
		}
		s.Request.Params[field.Key] = field.Value
	}
	return nil
}

func (s *stepFromPostman) makeRequestHeaders(item *TItem) error {
	// headers defined in postman collection
	s.Request.Headers = make(map[string]string)
	for _, field := range item.Request.Headers {
		if field.Disabled || strings.EqualFold(field.Key, "cookie") {
			continue
		}
		s.Request.Headers[field.Key] = field.Value
	}
	return nil
}

func (s *stepFromPostman) makeRequestCookies(item *TItem) error {
	// cookies defined in postman collection
	s.Request.Cookies = make(map[string]string)
	for _, field := range item.Request.Headers {
		if field.Disabled || !strings.EqualFold(field.Key, "cookie") {
			continue
		}
		s.parseRequestCookiesMap(field.Value)
	}
	return nil
}

func (s *stepFromPostman) parseRequestCookiesMap(cookies string) {
	for _, cookie := range strings.Split(cookies, ";") {
		cookie = strings.TrimSpace(cookie)
		index := strings.Index(cookie, "=")
		if index == -1 {
			log.Warn().Str("cookie", cookie).Msg("cookie format invalid")
			continue
		}
		s.Request.Cookies[cookie[:index]] = cookie[index+1:]
	}
}

func (s *stepFromPostman) makeRequestBody(item *TItem) error {
	mode := item.Request.Body.Mode
	if mode == "" {
		return nil
	}
	switch mode {
	case enumBodyRaw:
		return s.makeRequestBodyRaw(item)
	case enumBodyFormData:
		return s.makeRequestBodyFormData(item)
	case enumBodyUrlEncoded:
		return s.makeRequestBodyUrlEncoded(item)
	case enumBodyFile, enumBodyGraphQL:
		return errors.Errorf("unsupported body type: %v", mode)
	}
	return nil
}

func (s *stepFromPostman) makeRequestBodyRaw(item *TItem) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("make request body (raw) failed: %v", p)
		}
	}()

	languageType := "text"
	iOptions := item.Request.Body.Options
	if iOptions != nil {
		iLanguage := iOptions.(map[string]interface{})["raw"]
		if iLanguage != nil {
			languageType = iLanguage.(map[string]interface{})["language"].(string)
		}
	}

	s.Request.Body = item.Request.Body.Raw
	contentType := s.Request.Headers["Content-Type"]
	if strings.Contains(contentType, "application/json") || languageType == "json" {
		var iBody interface{}
		err = json.Unmarshal([]byte(item.Request.Body.Raw), &iBody)
		if err != nil {
			return errors.Wrap(err, "make request body (raw -> json) failed")
		}
		s.Request.Body = iBody
	}

	if contentType == "" {
		s.Request.Headers["Content-Type"] = contentTypeMap[languageType]
	}
	return
}

func (s *stepFromPostman) makeRequestBodyFormData(item *TItem) error {
	s.Request.Upload = make(map[string]interface{})
	for _, field := range item.Request.Body.FormData {
		if field.Disabled {
			continue
		}
		// form data could be text or file
		if field.Type == enumFieldTypeText {
			s.Request.Upload[field.Key] = field.Value
		} else if field.Type == enumFieldTypeFile {
			s.Request.Upload[field.Key] = field.Src
		} else {
			return errors.Errorf("make request body form data failed: unexpect field type: %v", field.Type)
		}
	}
	return nil
}

func (s *stepFromPostman) makeRequestBodyUrlEncoded(item *TItem) error {
	payloadMap := make(map[string]string)
	for _, field := range item.Request.Body.URLEncoded {
		if field.Disabled {
			continue
		}
		payloadMap[field.Key] = field.Value
	}
	s.Request.Body = payloadMap
	s.Request.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	return nil
}

// TODO makeValidate from test scripts
func (s *stepFromPostman) makeValidate(item *TItem) error {
	return nil
}
