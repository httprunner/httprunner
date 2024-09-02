package builtin

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	builtinJSON "encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
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

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

func Dump2JSON(data interface{}, path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Msg("convert absolute path failed")
		return err
	}
	log.Info().Str("path", path).Msg("dump data to json")

	// init json encoder
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")

	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, buffer.Bytes(), 0o644)
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
	case string:
		floatVar, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}
		return floatVar, err
	}
	// json.Number
	value, ok := i.(builtinJSON.Number)
	if ok {
		return value.Float64()
	}
	return 0, errors.New("failed to convert interface to float64")
}

func TypeNormalization(raw interface{}) interface{} {
	rawValue := reflect.ValueOf(raw)
	switch rawValue.Kind() {
	case reflect.Int:
		return rawValue.Int()
	case reflect.Int8:
		return rawValue.Int()
	case reflect.Int16:
		return rawValue.Int()
	case reflect.Int32:
		return rawValue.Int()
	case reflect.Float32:
		return rawValue.Float()
	case reflect.Uint:
		return rawValue.Uint()
	case reflect.Uint8:
		return rawValue.Uint()
	case reflect.Uint16:
		return rawValue.Uint()
	case reflect.Uint32:
		return rawValue.Uint()
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

func Bytes2File(data []byte, filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o755)
	defer file.Close()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate file")
	}
	count, err := file.Write(data)
	if err != nil {
		return err
	}
	log.Info().Msg(fmt.Sprintf("write file %s len: %d \n", filename, count))
	return nil
}

func Float32ToByte(v float32) []byte {
	bits := math.Float32bits(v)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

func ByteToFloat32(v []byte) float32 {
	bits := binary.LittleEndian.Uint32(v)
	return math.Float32frombits(bits)
}

func Float64ToByte(v float64) []byte {
	bits := math.Float64bits(v)
	bts := make([]byte, 8)
	binary.LittleEndian.PutUint64(bts, bits)
	return bts
}

func ByteToFloat64(v []byte) float64 {
	bits := binary.LittleEndian.Uint64(v)
	return math.Float64frombits(bits)
}

func Int64ToBytes(n int64) []byte {
	bytesBuf := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuf, binary.BigEndian, n)
	return bytesBuf.Bytes()
}

func BytesToInt64(bys []byte) (data int64) {
	byteBuff := bytes.NewBuffer(bys)
	_ = binary.Read(byteBuff, binary.BigEndian, &data)
	return
}

func SplitInteger(m, n int) (ints []int) {
	quotient := m / n
	remainder := m % n
	if remainder >= 0 {
		for i := 0; i < n-remainder; i++ {
			ints = append(ints, quotient)
		}
		for i := 0; i < remainder; i++ {
			ints = append(ints, quotient+1)
		}
		return
	} else if remainder < 0 {
		for i := 0; i < -remainder; i++ {
			ints = append(ints, quotient-1)
		}
		for i := 0; i < n+remainder; i++ {
			ints = append(ints, quotient)
		}
	}
	return
}

func sha256HMAC(key []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return []byte(fmt.Sprintf("%x", mac.Sum(nil)))
}

// ver: auth-v1or auth-v2
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

func ConvertToFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("invalid type for conversion to float64: %T, value: %+v", val, val)
	}
}

func ConvertToStringSlice(val interface{}) ([]string, error) {
	if valSlice, ok := val.([]interface{}); ok {
		var res []string
		for _, iVal := range valSlice {
			valString, ok := iVal.(string)
			if !ok {
				return nil, fmt.Errorf("invalid type for converting one of the elements to string: %T, value: %v", iVal, iVal)
			}
			res = append(res, valString)
		}
		return res, nil
	}
	return nil, fmt.Errorf("invalid type for conversion to []string")
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, errors.Wrap(err, "resolve tcp addr failed")
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, errors.Wrap(err, "listen tcp addr failed")
	}
	defer func() {
		if err = l.Close(); err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("close addr %s error", l.Addr().String()))
		}
	}()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func GetCurrentDay() string {
	now := time.Now()
	// 格式化日期为 yyyyMMdd
	formattedDate := now.Format("20060102")
	return formattedDate
}

func DownloadFile(filePath string, fileUrl string) error {
	log.Info().Str("filePath", filePath).Str("url", fileUrl).Msg("download file")
	parsedURL, err := url.Parse(fileUrl)
	if err != nil {
		return err
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 创建一个新的 HTTP 请求
	req, err := http.NewRequest("GET", fileUrl, nil)
	if err != nil {
		return err
	}

	if env.EAPI_TOKEN != "" {
		if parsedURL.Host != "gtf-eapi-cn.bytedance.com" && parsedURL.Host != "gtf-eapi-cn.bytedance.net" {
			return errors.New("invalid domain: must be gtf-eapi-cn.bytedance.com")
		}
		// 添加自定义头部
		req.Header.Add("accessKey", "ies.vedem.video")
		req.Header.Add("token", env.EAPI_TOKEN)
	}

	// 创建一个 HTTP 客户端并发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s, download failed", resp.Status)
	}

	// 将响应主体写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
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
