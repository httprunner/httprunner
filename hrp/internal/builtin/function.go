package builtin

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var Functions = map[string]interface{}{
	"get_timestamp":          getTimestamp,    // call without arguments
	"sleep":                  sleep,           // call with one argument
	"gen_random_string":      genRandomString, // call with one argument
	"max":                    math.Max,        // call with two arguments
	"md5":                    MD5,             // call with one argument
	"parameterize":           loadFromCSV,
	"P":                      loadFromCSV,
	"environ":                os.Getenv,
	"ENV":                    os.Getenv,
	"load_ws_message":        loadMessage,
	"multipart_encoder":      multipartEncoder,
	"multipart_content_type": multipartContentType,
}

func init() {
	rand.Seed(time.Now().UnixNano())
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

type TFormWriter struct {
	Writer  *multipart.Writer
	Payload *bytes.Buffer
}

func multipartEncoder(formMap map[string]interface{}) *TFormWriter {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	for formKey, formValue := range formMap {
		formValueString := fmt.Sprintf("%v", formValue)
		if err := writeFormDataFile(writer, formKey, formValueString); err == nil {
			// form value is a file path
			continue
		}
		// form value is not a file path, write as raw string
		if err := writer.WriteField(formKey, formValueString); err != nil {
			log.Info().Err(err).Msgf("failed to write field: %v=%v, ignore", formKey, formValue)
		}
	}
	if err := writer.Close(); err != nil {
	}
	return &TFormWriter{
		Writer:  writer,
		Payload: payload,
	}
}

func writeFormDataFile(writer *multipart.Writer, fName, fPath string) error {
	var err error
	fPath, err = filepath.Abs(fPath)
	if err != nil {
		log.Error().Err(err).Str("path", fPath).Msg("convert absolute path failed")
		return err
	}
	if !IsFilePathExists(fPath) {
		return errors.Errorf("file %v not existed", fPath)
	}
	file, err := os.ReadFile(fPath)
	if err != nil {
		log.Error().Err(err).Str("path", fPath).Msg("read file failed")
		return err
	}
	formFile, err := writer.CreateFormFile(fName, filepath.Base(fPath))
	if err != nil {
		return err
	}
	_, err = formFile.Write(file)
	return err
}

func multipartContentType(w *TFormWriter) string {
	if w.Writer == nil {
		return ""
	}
	return w.Writer.FormDataContentType()
}
