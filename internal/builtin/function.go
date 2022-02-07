package builtin

import (
	"bytes"
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math"
	"math/rand"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/rs/zerolog/log"
)

var Functions = map[string]interface{}{
	"get_timestamp":     getTimestamp,    // call without arguments
	"sleep":             sleep,           // call with one argument
	"gen_random_string": genRandomString, // call with one argument
	"max":               math.Max,        // call with two arguments
	"md5":               MD5,             // call with one argument
	"parameterize":      loadFromCSV,
	"P":                 loadFromCSV,
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

func loadFromCSV(path string) []map[string]interface{} {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Str("path", path).Err(err).Msg("convert absolute path failed")
		panic(err)
	}
	log.Info().Str("path", path).Msg("load csv file")

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("load csv file failed")
		panic(err)
	}
	r := csv.NewReader(strings.NewReader(string(file)))
	content, err := r.ReadAll()
	if err != nil {
		log.Error().Err(err).Msg("parse csv file failed")
		panic(err)
	}
	var result []map[string]interface{}
	for i := 1; i < len(content); i++ {
		row := make(map[string]interface{})
		for j := 0; j < len(content[i]); j++ {
			row[content[0][j]] = content[i][j]
		}
		result = append(result, row)
	}
	return result
}

func Dump2JSON(data interface{}, path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Msg("convert absolute path failed")
		return err
	}
	log.Info().Str("path", path).Msg("dump data to json")
	file, _ := json.MarshalIndent(data, "", "    ")
	err = ioutil.WriteFile(path, file, 0644)
	if err != nil {
		log.Error().Err(err).Msg("dump json path failed")
		return err
	}
	return nil
}

func Dump2YAML(data interface{}, path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Msg("convert absolute path failed")
		return err
	}
	log.Info().Str("path", path).Msg("dump data to yaml")

	// init yaml encoder
	buffer := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buffer)
	encoder.SetIndent(4)

	// encode
	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, buffer.Bytes(), 0644)
	if err != nil {
		log.Error().Err(err).Msg("dump yaml path failed")
		return err
	}
	return nil
}

func formatValue(raw interface{}) interface{} {
	rawValue := reflect.ValueOf(raw)
	switch rawValue.Kind() {
	case reflect.Map:
		m := make(map[string]interface{})
		for key, value := range rawValue.Interface().(map[string]interface{}) {
			b, _ := json.MarshalIndent(&value, "", "    ")
			m[key] = string(b)
		}
		return m
	case reflect.Slice:
		b, _ := json.MarshalIndent(&raw, "", "    ")
		return string(b)
	default:
		return raw
	}
}

func FormatResponse(raw interface{}) interface{} {
	formattedResponse := make(map[string]interface{})
	for key, value := range raw.(map[string]interface{}) {
		// convert value to json
		formattedResponse[key] = formatValue(value)
	}
	return formattedResponse
}
