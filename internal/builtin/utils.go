package builtin

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	builtinJSON "encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

func Dump2JSON(data interface{}, path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Msg("convert absolute path failed")
		return err
	}
	log.Info().Str("path", path).Msg("dump data to json")

	// Use standard library json encoder with consistent indentation and no HTML escaping
	buffer := new(bytes.Buffer)
	encoder := builtinJSON.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")

	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	// Ensure the JSON content is properly UTF-8 encoded
	// Go's json package already outputs UTF-8, but we explicitly validate it here
	jsonBytes := buffer.Bytes()

	// Create file and write content atomically to prevent corruption
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Error().Err(err).Msg("create json file failed")
		return err
	}
	defer file.Close()

	// Write JSON content directly (Go's json package ensures UTF-8 encoding)
	if _, err := file.Write(jsonBytes); err != nil {
		log.Error().Err(err).Msg("write json content failed")
		return err
	}

	// Ensure data is flushed to disk
	if err := file.Sync(); err != nil {
		log.Error().Err(err).Msg("sync json file failed")
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

	err = os.WriteFile(path, buffer.Bytes(), 0o644)
	if err != nil {
		log.Error().Err(err).Msg("dump yaml path failed")
		return err
	}
	return nil
}

func FormatResponse(raw interface{}) interface{} {
	formattedResponse := make(map[string]interface{})
	for key, value := range raw.(map[string]interface{}) {
		// convert value to json
		if key == "body" {
			b, _ := json.MarshalIndent(&value, "", "    ")
			value = string(b)
		}
		formattedResponse[key] = value
	}
	return formattedResponse
}

func CreateFolder(folderPath string) error {
	log.Info().Str("path", folderPath).Msg("create folder")
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("create folder failed")
		return err
	}
	return nil
}

func CreateFile(filePath string, data string) error {
	log.Info().Str("path", filePath).Msg("create file")
	err := os.WriteFile(filePath, []byte(data), 0o644)
	if err != nil {
		log.Error().Err(err).Msg("create file failed")
		return err
	}
	return nil
}

// IsPathExists returns true if path exists, whether path is file or dir
func IsPathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// IsFilePathExists returns true if path exists and path is file
func IsFilePathExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// path not exists
		return false
	}

	// path exists
	if info.IsDir() {
		// path is dir, not file
		return false
	}
	return true
}

// IsFolderPathExists returns true if path exists and path is folder
func IsFolderPathExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// path not exists
		return false
	}

	// path exists and is dir
	return info.IsDir()
}

func EnsureFolderExists(folderPath string) error {
	if !IsPathExists(folderPath) {
		err := CreateFolder(folderPath)
		return err
	} else if IsFilePathExists(folderPath) {
		return fmt.Errorf("path %v should be directory", folderPath)
	}
	return nil
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GetRandomNumber(min, max int) int {
	if min > max {
		return 0
	}
	r := rand.Intn(max - min + 1)
	return min + r
}

func Interface2Float64(i interface{}) (float64, error) {
	switch v := i.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string: // e.g. "1", "0.5"
		floatVar, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Error().Err(err).Str("value", v).
				Msg("convert string to float64 failed")
			return 0, err
		}
		return floatVar, nil
	}
	// json.Number
	value, ok := i.(builtinJSON.Number)
	if ok {
		return value.Float64()
	}

	// Log error for unsupported types
	log.Error().Interface("value", i).Type("type", i).
		Msg("convert float64 failed")
	return 0, errors.New("failed to convert interface to float64")
}

func TypeNormalization(raw interface{}) interface{} {
	switch v := raw.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int()
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint()
	case float32, float64:
		return reflect.ValueOf(v).Float()
	default:
		return raw
	}
}

func InterfaceType(raw interface{}) string {
	if raw == nil {
		return ""
	}
	return reflect.TypeOf(raw).String()
}

func LoadFile(path string) ([]byte, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("convert absolute path failed")
		return nil, errors.Wrap(code.LoadFileError, err.Error())
	}

	file, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("read file failed")
		return nil, errors.Wrap(code.LoadFileError, err.Error())
	}
	return file, nil
}

func loadFromCSV(path string) []map[string]interface{} {
	log.Info().Str("path", path).Msg("load csv file")
	file, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("read csv file failed")
		os.Exit(code.GetErrorCode(err))
	}

	r := csv.NewReader(strings.NewReader(string(file)))
	content, err := r.ReadAll()
	if err != nil {
		log.Error().Err(err).Msg("parse csv file failed")
		os.Exit(code.GetErrorCode(err))
	}
	firstLine := content[0] // parameter names
	var result []map[string]interface{}
	for i := 1; i < len(content); i++ {
		row := make(map[string]interface{})
		for j := 0; j < len(content[i]); j++ {
			row[firstLine[j]] = content[i][j]
		}
		result = append(result, row)
	}
	return result
}

func loadMessage(path string) []byte {
	log.Info().Str("path", path).Msg("load message file")
	file, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("read message file failed")
		os.Exit(code.GetErrorCode(err))
	}
	return file
}

func GetFileNameWithoutExtension(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return base[0 : len(base)-len(ext)]
}

func sha256HMAC(key []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return []byte(fmt.Sprintf("%x", mac.Sum(nil)))
}

// ver: auth-v1 or auth-v2
func Sign(ver string, ak string, sk string, body []byte) string {
	expiration := 1800
	signKeyInfo := fmt.Sprintf("%s/%s/%d/%d", ver, ak, time.Now().Unix(), expiration)
	signKey := sha256HMAC([]byte(sk), []byte(signKeyInfo))
	signResult := sha256HMAC(signKey, body)
	return fmt.Sprintf("%v/%v", signKeyInfo, string(signResult))
}

func GenNameWithTimestamp(tmpl string) string {
	if !strings.Contains(tmpl, "%d") {
		tmpl = tmpl + "_%d"
	}
	return fmt.Sprintf(tmpl, time.Now().Unix())
}

func IsZeroFloat64(f float64) bool {
	threshold := 1e-9
	return math.Abs(f) < threshold
}

func ConvertToFloat64Slice(val interface{}) ([]float64, error) {
	if paramsSlice, ok := val.([]float64); ok {
		return paramsSlice, nil
	}
	paramsSlice, ok := val.([]interface{})
	if !ok {
		return nil, errors.New("val is not slice")
	}

	var err error
	float64Slice := make([]float64, len(paramsSlice))
	for i, v := range paramsSlice {
		float64Slice[i], err = Interface2Float64(v)
		if err != nil {
			return nil, err
		}
	}
	return float64Slice, nil
}

func ConvertToStringSlice(val interface{}) ([]string, error) {
	paramsSlice, ok := val.([]interface{})
	if !ok {
		return nil, errors.New("val is not slice")
	}

	stringSlice := make([]string, len(paramsSlice))
	for i, v := range paramsSlice {
		stringSlice[i], ok = v.(string)
		if !ok {
			return nil, errors.New("val is not string slice")
		}
	}
	return stringSlice, nil
}

// RoundToOneDecimal rounds a float64 value to 1 decimal place
func RoundToOneDecimal(val float64) float64 {
	return math.Round(val*10) / 10.0
}

func GetFreePort() (int, error) {
	minPort := 20000
	maxPort := 50000
	for i := 0; i < 10; i++ {
		port := rand.Intn(maxPort-minPort+1) + minPort
		addr := fmt.Sprintf("0.0.0.0:%d", port)
		l, err := net.Listen("tcp", addr)
		if err == nil {
			defer l.Close() // 端口成功绑定后立即释放，返回该端口号
			return port, nil
		}
	}
	return 0, errors.New("failed to get available port")
}

func GetCurrentDay() string {
	now := time.Now()
	// 格式化日期为 yyyyMMdd
	formattedDate := now.Format("20060102")
	return formattedDate
}

func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false // 文件不存在
	}
	return err == nil // 文件存在，且没有其他错误
}

func RunCommand(cmdName string, args ...string) error {
	cmd := exec.Command(cmdName, args...)
	log.Info().Str("command", cmd.String()).Msg("exec command")

	// print stderr output
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		log.Error().Err(err).Msg("failed to exec command. msg: " + stderrStr)
		if stderrStr != "" {
			err = errors.Wrap(err, stderrStr)
		}
		return err
	}
	stderrStr := stderr.String()
	log.Error().Msg("failed to exec command. msg: " + stderrStr)
	log.Info().Msg("exec command output: " + stdout.String())
	return nil
}

type LineCallback func(line string) bool

// RunCommandWithCallback 运行命令并根据回调判断是否成功
func RunCommandWithCallback(cmdName string, args []string, callback LineCallback) error {
	cmd := exec.Command(cmdName, args...)
	log.Info().Str("command", cmd.String()).Msg("exec command")

	// 使用管道获取标准输出
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Error().Err(err).Msg("failed to get stdout pipe")
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		log.Error().Err(err).Msg("failed to start command")
		return err
	}

	// 创建一个用于标识成功的通道
	done := make(chan struct{})
	defer close(done)

	// 逐行读取 stdout
	go func() {
		stdoutScanner := bufio.NewScanner(stdoutPipe)
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			log.Info().Msg("stdout: " + line)
			if callback(line) {
				done <- struct{}{}
				return
			}
		}
	}()

	// 等待命令执行完成
	err = cmd.Wait()
	if err != nil {
		log.Error().Msg("failed to exec command. msg: " + stderr.String())
		return err
	}

	// 设置一个1秒的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// 超时，判断失败
		log.Error().Msg("failed to exec command. msg: " + stderr.String())
		err = errors.New("command execution failed: callback failed while exec command")
		log.Error().Err(err).Msg("failed to find keyword in time")
		return err
	}
}

// LoadImage loads image file and returns base64 encoded string and image size
func LoadImage(imagePath string) (base64Str string, size types.Size, err error) {
	// Read the image file
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return "", types.Size{}, fmt.Errorf("failed to open image file: %w", err)
	}
	defer imageFile.Close()

	// Decode the image to get its resolution
	imageData, format, err := image.Decode(imageFile)
	if err != nil {
		return "", types.Size{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get the resolution of the image
	width := imageData.Bounds().Dx()
	height := imageData.Bounds().Dy()
	size = types.Size{Width: width, Height: height}

	// Convert image to base64
	buf := new(bytes.Buffer)
	if format == "jpeg" || format == "jpg" {
		if err := jpeg.Encode(buf, imageData, nil); err != nil {
			return "", types.Size{}, fmt.Errorf("failed to encode image to buffer: %w", err)
		}
	} else {
		// default use png
		if err := png.Encode(buf, imageData); err != nil {
			return "", types.Size{}, fmt.Errorf("failed to encode image to buffer: %w", err)
		}
	}
	base64Str = fmt.Sprintf("data:image/%s;base64,%s", format,
		base64.StdEncoding.EncodeToString(buf.Bytes()))

	return base64Str, size, nil
}
