package builtin

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"mime"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var Functions = map[string]interface{}{
	"get_timestamp":          getTimestamp,    // call without arguments
	"sleep":                  sleep,           // call with one argument
	"gen_random_string":      genRandomString, // call with one argument
	"random_int":             rand.Intn,       // call with one argument
	"random_range":           random_range,    // call with two arguments
	"max":                    math.Max,        // call with two arguments
	"md5":                    MD5,             // call with one argument
	"parameterize":           loadFromCSV,
	"P":                      loadFromCSV,
	"split_by_comma":         splitByComma, // call with one argument
	"environ":                os.Getenv,
	"ENV":                    os.Getenv,
	"load_ws_message":        loadMessage,
	"multipart_encoder":      multipartEncoder,
	"multipart_content_type": multipartContentType,
}

// upload file path must starts with @, like @\"PATH\" or @PATH
var regexUploadFilePath = regexp.MustCompile(`^@(.*)`)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func random_range(a, b float64) float64 {
	return a + rand.Float64()*(b-a)
}

func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func sleep(nSecs int) {
	time.Sleep(time.Duration(nSecs) * time.Second)
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func genRandomString(n int) string {
	lettersLen := len(letters)
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(lettersLen)]
	}
	return string(b)
}

func MD5(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}

type TFormDataWriter struct {
	Writer  *multipart.Writer
	Payload *bytes.Buffer
}

func (w *TFormDataWriter) writeCustomText(formKey, formValue, formType, formFileName string) error {
	if w.Writer == nil {
		return errors.New("form-data writer not initialized")
	}
	h := make(textproto.MIMEHeader)
	// text doesn't have Content-Type by default
	if formType != "" {
		h.Set("Content-Type", formType)
	}
	// text doesn't have filename in Content-Disposition by default
	if formFileName == "" {
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(formKey)))
	} else {
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				escapeQuotes(formKey), escapeQuotes(formFileName)))
	}
	part, err := w.Writer.CreatePart(h)
	if err != nil {
		return err
	}

	_, err = part.Write([]byte(formValue))
	return err
}

func (w *TFormDataWriter) writeCustomFile(formKey, formValue, formType, formFileName string) error {
	if w.Writer == nil {
		return errors.New("form-data writer not initialized")
	}
	fPath, err := filepath.Abs(formValue)
	if err != nil {
		return err
	}
	file, err := os.ReadFile(fPath)
	if err != nil {
		return err
	}

	if formType == "" {
		formType = inferFormType(formValue)
	}
	if formFileName == "" {
		formFileName = filepath.Base(formValue)
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", formType)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(formKey), escapeQuotes(formFileName)))
	part, err := w.Writer.CreatePart(h)
	if err != nil {
		return err
	}

	_, err = part.Write(file)
	return err
}

func inferFormType(formValue string) string {
	extName := filepath.Ext(formValue)
	formType := mime.TypeByExtension(extName)
	if formType == "" {
		// file without extension name
		return "application/octet-stream"
	}
	if strings.HasPrefix(formType, "text") {
		// text/... types have the charset parameter set to "utf-8" by default.
		return strings.TrimSuffix(formType, "; charset=utf-8")
	}
	return formType
}

func multipartEncoder(formMap map[string]interface{}) (*TFormDataWriter, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	tFormWriter := &TFormDataWriter{
		Writer:  writer,
		Payload: payload,
	}
	// e.g. formMap: {"file": "@\"$upload_file\";type=text/foo"}
	for formKey, formData := range formMap {
		formDataString := fmt.Sprintf("%v", formData)
		formItems := strings.Split(formDataString, ";")
		var isFilePath bool
		var formValue, formType, formFileName string
		for _, formItem := range formItems {
			if formItem == "" {
				continue
			}
			equalSignIndex := strings.Index(formItem, "=")
			// parse form value, e.g. @\"$upload_file\"
			if equalSignIndex == -1 {
				matchRes := regexUploadFilePath.FindStringSubmatch(formItem)
				if len(matchRes) > 1 {
					// formItem started with @, regarded as File path
					isFilePath = true
					formValue = strings.Trim(matchRes[1], "\"")
				} else {
					// formItem is not a valid File path, regarded as Text instead
					formValue = strings.TrimSuffix(strings.TrimPrefix(formItem, "\""), "\"")
				}
				continue
			}
			// parse form option, e.g. type=text/plain
			leftPart := strings.TrimSpace(formItem[:equalSignIndex])
			var rightPart string
			if equalSignIndex < len(formItem)-1 {
				rightPart = strings.TrimSpace(formItem[equalSignIndex+1:])
			}
			if (strings.ToLower(leftPart) != "type" && strings.ToLower(leftPart) != "filename") || rightPart == "" {
				formOption := fmt.Sprintf("%s=%s", leftPart, rightPart)
				log.Warn().Msgf("invalid form option: %v, ignore", formOption)
				continue
			}
			if strings.ToLower(leftPart) == "type" {
				formType = rightPart
			}
			if strings.ToLower(leftPart) == "filename" {
				formFileName = rightPart
			}
		}
		if isFilePath {
			if err := tFormWriter.writeCustomFile(formKey, formValue, formType, formFileName); err != nil {
				log.Error().Err(err).Msgf("failed to write file: %v=@\"%v\", exit", formKey, formValue)
				return nil, err
			}
			continue
		}
		if err := tFormWriter.writeCustomText(formKey, formValue, formType, formFileName); err != nil {
			log.Error().Err(err).Msgf("failed to write text: %v=%v, ignore", formKey, formValue)
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close form-data writer")
	}
	return tFormWriter, nil
}

func multipartContentType(w *TFormDataWriter) string {
	if w.Writer == nil {
		return ""
	}
	return w.Writer.FormDataContentType()
}

func splitByComma(s string) []string {
	return strings.Split(s, ",")
}
