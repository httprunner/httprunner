package har2case

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/httprunner/httpboomer"
)

func NewHAR(path string) *HAR {
	return &HAR{
		path: path,
	}
}

type HAR struct {
	path       string
	filterStr  string
	excludeStr string
}

func (h *HAR) GenJSON() (jsonPath string, err error) {
	jsonPath = getFilenameWithoutExtension(h.path) + ".json"

	tCase, err := h.makeTestCase()
	if err != nil {
		return "", err
	}
	err = tCase.Dump2JSON(jsonPath)
	return
}

func (h *HAR) GenYAML() (yamlPath string, err error) {
	yamlPath = getFilenameWithoutExtension(h.path) + ".yaml"

	tCase, err := h.makeTestCase()
	if err != nil {
		return "", err
	}
	err = tCase.Dump2YAML(yamlPath)
	return
}

func (h *HAR) makeTestCase() (*httpboomer.TCase, error) {
	teststeps, err := h.prepareTestSteps()
	if err != nil {
		return nil, err
	}

	tCase := &httpboomer.TCase{
		Config:    *h.prepareConfig(),
		TestSteps: teststeps,
	}
	return tCase, nil
}

func (h *HAR) load() (*Har, error) {
	fp, err := os.Open(h.path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	data, err := ioutil.ReadAll(fp)
	fp.Close()
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	har := &Har{}
	err = json.Unmarshal(data, har)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal error: %w", err)
	}

	return har, nil
}

func (h *HAR) prepareConfig() *httpboomer.TConfig {
	return &httpboomer.TConfig{
		Name:      "testcase description",
		Variables: make(map[string]interface{}),
		Verify:    false,
	}
}

func (h *HAR) prepareTestSteps() ([]*httpboomer.TStep, error) {
	har, err := h.load()
	if err != nil {
		return nil, err
	}

	var steps []*httpboomer.TStep
	for _, entry := range har.Log.Entries {
		step, err := h.prepareTestStep(&entry)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	return steps, nil
}

func (h *HAR) prepareTestStep(entry *Entry) (*httpboomer.TStep, error) {
	tStep := &TStep{
		TStep: httpboomer.TStep{
			Request:    &httpboomer.TRequest{},
			Validators: make([]httpboomer.TValidator, 0),
		},
	}
	if err := tStep.makeRequestMethod(entry); err != nil {
		return nil, err
	}
	if err := tStep.makeRequestURL(entry); err != nil {
		return nil, err
	}
	if err := tStep.makeRequestParams(entry); err != nil {
		return nil, err
	}
	if err := tStep.makeRequestCookies(entry); err != nil {
		return nil, err
	}
	if err := tStep.makeRequestHeaders(entry); err != nil {
		return nil, err
	}
	if err := tStep.makeRequestBody(entry); err != nil {
		return nil, err
	}
	if err := tStep.makeValidate(); err != nil {
		return nil, err
	}
	return &tStep.TStep, nil
}

type TStep struct {
	httpboomer.TStep
}

func (s *TStep) makeRequestMethod(entry *Entry) error {
	s.Request.Method = httpboomer.EnumHTTPMethod(entry.Request.Method)
	return nil
}

func (s *TStep) makeRequestURL(entry *Entry) error {

	u, err := url.Parse(entry.Request.URL)
	if err != nil {
		log.Printf("makeRequestURL error: %v", err)
		return err
	}
	s.Request.URL = fmt.Sprintf("%s://%s", u.Scheme, u.Hostname()+u.Path)
	return nil
}

func (s *TStep) makeRequestParams(entry *Entry) error {
	s.Request.Params = make(map[string]interface{})
	for _, param := range entry.Request.QueryString {
		s.Request.Params[param.Name] = param.Value
	}
	return nil
}

func (s *TStep) makeRequestCookies(entry *Entry) error {
	s.Request.Cookies = make(map[string]string)
	for _, cookie := range entry.Request.Cookies {
		s.Request.Cookies[cookie.Name] = cookie.Value
	}
	return nil
}

func (s *TStep) makeRequestHeaders(entry *Entry) error {
	s.Request.Headers = make(map[string]string)
	for _, header := range entry.Request.Headers {
		if strings.EqualFold(header.Name, "cookie") {
			continue
		}
		s.Request.Headers[header.Name] = header.Value
	}
	return nil
}

func (s *TStep) makeRequestBody(entry *Entry) error {
	mimeType := entry.Request.PostData.MimeType
	if strings.HasPrefix(mimeType, "application/json") {
		// post json
		var data interface{}
		err := json.Unmarshal([]byte(entry.Request.PostData.Text), &data)
		if err != nil {
			log.Printf("makeRequestBody error: %v", err)
			return err
		}
		s.Request.Data = data
	} else if strings.HasPrefix(mimeType, "application/x-www-form-urlencoded") {
		// TODO: post form data
	}
	return nil
}

func (s *TStep) makeValidate() error {
	s.Validators = nil
	return nil
}

func getFilenameWithoutExtension(path string) string {
	ext := filepath.Ext(path)
	return path[0 : len(path)-len(ext)]
}
