package postman2case

import (
	"bytes"
	"fmt"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

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

const (
	suffixName    = ".converted"
	extensionJSON = ".json"
	extensionYAML = ".yaml"
)

var contentTypeMap = map[string]string{
	"text":       "text/plain",
	"javascript": "application/javascript",
	"json":       "application/json",
	"html":       "text/html",
	"xml":        "application/xml",
}

func NewCollection(path string) *collection {
	return &collection{
		path: path,
	}
}

type collection struct {
	path      string
	outputDir string
}

func (c *collection) SetOutputDir(dir string) {
	log.Info().Str("dir", dir).Msg("set output directory")
	c.outputDir = dir
}

func (c *collection) GenJSON() (jsonPath string, err error) {
	testCase, err := c.makeTestCase()
	if err != nil {
		return "", err
	}
	jsonPath = c.genOutputPath(extensionJSON)
	err = builtin.Dump2JSON(testCase, jsonPath)
	return
}

func (c *collection) GenYAML() (yamlPath string, err error) {
	testCase, err := c.makeTestCase()
	if err != nil {
		return "", err
	}
	yamlPath = c.genOutputPath(extensionYAML)
	err = builtin.Dump2YAML(testCase, yamlPath)
	return
}

func (c *collection) genOutputPath(suffix string) string {
	file := getFilenameWithoutExtension(c.path) + suffix
	if c.outputDir != "" {
		return filepath.Join(c.outputDir, file)
	} else {
		return filepath.Join(filepath.Dir(c.path), file)
	}
}

func getFilenameWithoutExtension(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return base[0:len(base)-len(ext)] + suffixName
}

func (c *collection) makeTestCase() (*hrp.TCase, error) {
	tCollection, err := c.load()
	if err != nil {
		return nil, err
	}
	teststeps, err := c.prepareTestSteps(tCollection)
	if err != nil {
		return nil, err
	}
	tCase := &hrp.TCase{
		Config:    c.prepareConfig(tCollection),
		TestSteps: teststeps,
	}
	return tCase, nil
}

func (c *collection) load() (*TCollection, error) {
	collection := &TCollection{}
	err := builtin.LoadFile(c.path, collection)
	if err != nil {
		return nil, errors.Wrap(err, "load postman collection failed")
	}
	return collection, nil
}

func (c *collection) prepareConfig(tCollection *TCollection) *hrp.TConfig {
	return hrp.NewConfig(tCollection.Info.Name).
		SetVerifySSL(false)
}

func (c *collection) prepareTestSteps(tCollection *TCollection) ([]*hrp.TStep, error) {
	// recursively convert collection items into a list
	var itemList []TItem
	for _, item := range tCollection.Items {
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

func (c *collection) prepareTestStep(item *TItem) (*hrp.TStep, error) {
	log.Info().
		Str("method", item.Request.Method).
		Str("url", item.Request.URL.Raw).
		Msg("convert teststep")

	step := &tStep{
		hrp.TStep{
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
	if err := step.makeRequestHeadersAndCookies(item); err != nil {
		return nil, err
	}
	if err := step.makeRequestBody(item); err != nil {
		return nil, err
	}
	if err := step.makeValidate(item); err != nil {
		return nil, err
	}
	return &step.TStep, nil
}

type tStep struct {
	hrp.TStep
}

// makeRequestName indicates the step name the same as item name
func (s *tStep) makeRequestName(item *TItem) error {
	s.Name = item.Name
	return nil
}

func (s *tStep) makeRequestMethod(item *TItem) error {
	s.Request.Method = hrp.HTTPMethod(item.Request.Method)
	return nil
}

func (s *tStep) makeRequestURL(item *TItem) error {
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

func (s *tStep) makeRequestParams(item *TItem) error {
	s.Request.Params = make(map[string]interface{})
	for _, field := range item.Request.URL.Query {
		if field.Disabled {
			continue
		}
		s.Request.Params[field.Key] = field.Value
	}
	return nil
}

func (s *tStep) makeRequestHeadersAndCookies(item *TItem) error {
	s.Request.Headers = make(map[string]string)
	for _, field := range item.Request.Headers {
		if field.Disabled {
			continue
		}
		if strings.EqualFold(field.Key, "cookie") {
			s.Request.Cookies[field.Key] = field.Value
			continue
		}
		s.Request.Headers[field.Key] = field.Value
	}
	return nil
}

func (s *tStep) makeRequestBody(item *TItem) error {
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
		return errors.New("not supported body type")
	}
	return nil
}

func (s *tStep) makeRequestBodyRaw(item *TItem) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("make request body raw failed: %v", p)
		}
	}()

	// extract language type
	iOptions := item.Request.Body.Options
	iLanguage := iOptions.(map[string]interface{})["raw"]
	languageType := iLanguage.(map[string]interface{})["language"].(string)

	// make request body and indicate Content-Type
	rawBody := item.Request.Body.Raw
	if languageType == "json" {
		var iBody interface{}
		err = json.Unmarshal([]byte(rawBody), &iBody)
		if err != nil {
			return errors.Wrap(err, "make request body raw failed")
		}
		s.Request.Body = iBody
	} else {
		s.Request.Body = rawBody
	}
	s.Request.Headers["Content-Type"] = contentTypeMap[languageType]
	return
}

func (s *tStep) makeRequestBodyFormData(item *TItem) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "make request body form-data failed")
		}
	}()
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	for _, field := range item.Request.Body.FormData {
		if field.Disabled {
			continue
		}
		// form data could be text or file
		if field.Type == enumFieldTypeText {
			err = writer.WriteField(field.Key, field.Value)
			if err != nil {
				return
			}
		} else if field.Type == enumFieldTypeFile {
			err = writeFormDataFile(writer, &field)
			if err != nil {
				return
			}
		}
	}
	err = writer.Close()
	s.Request.Body = payload.String()
	s.Request.Headers["Content-Type"] = writer.FormDataContentType()
	return
}

func writeFormDataFile(writer *multipart.Writer, field *TField) error {
	file, err := os.Open(field.Src)
	if err != nil {
		return err
	}
	defer file.Close()
	formFile, err := writer.CreateFormFile(field.Key, filepath.Base(field.Src))
	if err != nil {
		return err
	}
	_, err = io.Copy(formFile, file)
	return err
}

func (s *tStep) makeRequestBodyUrlEncoded(item *TItem) error {
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

// TODO makeValidate from example response
func (s *tStep) makeValidate(item *TItem) error {
	return nil
}
