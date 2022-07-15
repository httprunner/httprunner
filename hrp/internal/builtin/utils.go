package builtin

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/csv"
	builtinJSON "encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

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

var python3Executable string = "python3" // system default python3

// EnsurePython3Venv ensures python3 venv with specified packages
// venv should be directory path of target venv
func EnsurePython3Venv(venv string, packages ...string) (python3 string, err error) {
	// priority: specified > $HOME/.hrp/venv
	if venv == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "get user home dir failed")
		}
		venv = filepath.Join(home, ".hrp", "venv")
	}
	python3, err = ensurePython3Venv(venv, packages...)
	if err != nil {
		return "", errors.Wrap(err, "prepare python3 venv failed")
	}
	python3Executable = python3
	log.Info().Str("Python3Executable", python3Executable).Msg("set python3 executable path")
	return python3, nil
}

func ExecPython3Command(cmdName string, args ...string) error {
	args = append([]string{"-m", cmdName}, args...)
	return ExecCommand(python3Executable, args...)
}

func AssertPythonPackage(python3 string, pkgName, pkgVersion string) error {
	out, err := exec.Command(
		python3, "-c", fmt.Sprintf("import %s; print(%s.__version__)", pkgName, pkgName),
	).Output()
	if err != nil {
		return fmt.Errorf("python package %s not found", pkgName)
	}

	// do not check version if pkgVersion is empty
	if pkgVersion == "" {
		log.Info().Str("name", pkgName).Msg("python package is ready")
		return nil
	}

	// check package version equality
	version := strings.TrimSpace(string(out))
	if strings.TrimLeft(version, "v") != strings.TrimLeft(pkgVersion, "v") {
		return fmt.Errorf("python package %s version %s not matched, please upgrade to %s",
			pkgName, version, pkgVersion)
	}

	log.Info().Str("name", pkgName).Str("version", pkgVersion).Msg("python package is ready")
	return nil
}

func InstallPythonPackage(python3 string, pkg string) (err error) {
	var pkgName, pkgVersion string
	if strings.Contains(pkg, "==") {
		// funppy==0.5.0
		pkgInfo := strings.Split(pkg, "==")
		pkgName = pkgInfo[0]
		pkgVersion = pkgInfo[1]
	} else {
		// funppy
		pkgName = pkg
	}

	// check if package installed and version matched
	err = AssertPythonPackage(python3, pkgName, pkgVersion)
	if err == nil {
		return nil
	}

	// check if pip available
	err = ExecCommand(python3, "-m", "pip", "--version")
	if err != nil {
		log.Warn().Msg("pip is not available")
		return errors.Wrap(err, "pip is not available")
	}

	log.Info().Str("pkgName", pkgName).Str("pkgVersion", pkgVersion).Msg("installing python package")

	// install package
	pypiIndexURL := os.Getenv("PYPI_INDEX_URL")
	if pypiIndexURL == "" {
		pypiIndexURL = "https://pypi.org/simple" // default
	}
	err = ExecCommand(python3, "-m", "pip", "install", "--upgrade", pkg,
		"--index-url", pypiIndexURL,
		"--quiet", "--disable-pip-version-check")
	if err != nil {
		return errors.Wrap(err, "pip install package failed")
	}

	return AssertPythonPackage(python3, pkgName, pkgVersion)
}

func ExecCommandInDir(cmd *exec.Cmd, dir string) error {
	log.Info().Str("cmd", cmd.String()).Str("dir", dir).Msg("exec command")
	cmd.Dir = dir

	// print output with colors
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Error().Err(err).Msg("exec command failed")
		return err
	}

	return nil
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
	switch i.(type) {
	case int:
		return float64(i.(int)), nil
	case int32:
		return float64(i.(int32)), nil
	case int64:
		return float64(i.(int64)), nil
	case float32:
		return float64(i.(float32)), nil
	case float64:
		return i.(float64), nil
	case string:
		intVar, err := strconv.Atoi(i.(string))
		if err != nil {
			return 0, err
		}
		return float64(intVar), err
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

var ErrUnsupportedFileExt = fmt.Errorf("unsupported file extension")

// LoadFile loads file content with file extension and assigns to structObj
func LoadFile(path string, structObj interface{}) (err error) {
	log.Info().Str("path", path).Msg("load file")
	file, err := ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "read file failed")
	}
	// remove BOM at the beginning of file
	file = bytes.TrimLeft(file, "\xef\xbb\xbf")
	ext := filepath.Ext(path)
	switch ext {
	case ".json", ".har":
		decoder := json.NewDecoder(bytes.NewReader(file))
		decoder.UseNumber()
		err = decoder.Decode(structObj)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(file, structObj)
	case ".env":
		err = parseEnvContent(file, structObj)
	default:
		err = ErrUnsupportedFileExt
	}
	return err
}

func parseEnvContent(file []byte, obj interface{}) error {
	envMap := obj.(map[string]string)
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			// empty line or comment line
			continue
		}
		var kv []string
		if strings.Contains(line, "=") {
			kv = strings.SplitN(line, "=", 2)
		} else if strings.Contains(line, ":") {
			kv = strings.SplitN(line, ":", 2)
		}
		if len(kv) != 2 {
			return errors.New(".env format error")
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		envMap[key] = value

		// set env
		log.Info().Str("key", key).Msg("set env")
		os.Setenv(key, value)
	}
	return nil
}

func loadFromCSV(path string) []map[string]interface{} {
	log.Info().Str("path", path).Msg("load csv file")
	file, err := ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("read csv file failed")
		os.Exit(1)
	}

	r := csv.NewReader(strings.NewReader(string(file)))
	content, err := r.ReadAll()
	if err != nil {
		log.Error().Err(err).Msg("parse csv file failed")
		os.Exit(1)
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
	file, err := ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("read message file failed")
		os.Exit(1)
	}
	return file
}

func ReadFile(path string) ([]byte, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("convert absolute path failed")
		return nil, err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("read file failed")
		return nil, err
	}
	return file, nil
}

func ReadCmdLines(path string) ([]string, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("convert absolute path failed")
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("open file failed")
		return nil, err
	}
	defer file.Close()

	var line string
	var lines []string
	scanner := bufio.NewScanner(file)
	// FIXME: resize scanner's capacity for lines over 64K
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || text == "\n" {
			continue
		}
		if strings.HasSuffix(text, "\\") {
			line = line + strings.Trim(text, "\\")
			continue
		}
		line = line + text
		lines = append(lines, line)
		line = ""
	}
	return lines, scanner.Err()
}

func GetFileNameWithoutExtension(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return base[0 : len(base)-len(ext)]
}

func Bytes2File(data []byte, filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
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
