package convert

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/shlex"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

const (
	originCmdKey = "_origin_cmd_key"
	targetUrlKey = "_target_url_key"
)

var curlOptionAliasMap = map[string]string{
	"-a": "--append",
	"-A": "--user-agent",
	"-b": "--cookie",
	"-B": "--use-ascii",
	"-c": "--cookie-jar",
	"-C": "--continue-at",
	"-d": "--data",
	"-D": "--dump-header",
	"-e": "--referer",
	"-E": "--cert",
	"-f": "--fail",
	"-F": "--form",
	"-g": "--globoff",
	"-G": "--get",
	"-h": "--help",
	"-H": "--header",
	"-i": "--include",
	"-I": "--head",
	"-j": "--junk-session-cookies",
	"-J": "--remote-header-name",
	"-k": "--insecure",
	"-K": "--config",
	"-l": "--list-only",
	"-L": "--location",
	"-m": "--max-time",
	"-M": "--manual",
	"-n": "--netrc",
	"-N": "--no-buffer",
	"-o": "--output",
	"-O": "--remote-name",
	"-p": "--proxytunnel",
	"-P": "--ftp-port",
	"-q": "--disable",
	"-Q": "--quote",
	"-r": "--range",
	"-R": "--remote-time",
	"-s": "--silent",
	"-S": "--show-error",
	"-t": "--telnet-option",
	"-T": "--upload-file",
	"-u": "--user",
	"-U": "--proxy-user",
	"-v": "--verbose",
	"-V": "--version",
	"-w": "--write-out",
	"-x": "--proxy",
	"-X": "--request",
	"-Y": "--speed-limit",
	"-y": "--speed-time",
	"-z": "--time-cond",
	"-Z": "--parallel",
}

var curlOptionWhiteMap = map[string]struct{}{
	"--cookie":  {},
	"--data":    {},
	"--form":    {},
	"--get":     {},
	"--head":    {},
	"--header":  {},
	"--request": {},
}

var curlOptionWhiteList []string

func init() {
	for option := range curlOptionWhiteMap {
		curlOptionWhiteList = append(curlOptionWhiteList, option)
	}
}

// LoadCurlCase loads testcase from one or more curl commands in .txt file
func LoadCurlCase(path string) (*hrp.TCase, error) {
	cmds, err := readFileLines(path)
	if err != nil {
		return nil, err
	}
	tCase := &hrp.TCase{
		Config: &hrp.TConfig{
			Name: "testcase converted from curl command",
		},
	}
	for _, cmd := range cmds {
		tSteps, err := LoadCurlSteps(cmd)
		if err != nil {
			return nil, err
		}
		tCase.TestSteps = append(tCase.TestSteps, tSteps...)
	}
	err = tCase.MakeCompat()
	if err != nil {
		return nil, err
	}
	return tCase, nil
}

func readFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("open file failed")
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "\n" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

// LoadCurlSteps loads one teststep from one curl command
func LoadCurlSteps(cmd string) ([]*hrp.TStep, error) {
	caseCurl, err := loadCaseCurl(cmd)
	if err != nil {
		return nil, err
	}
	return caseCurl.toTSteps()
}

func loadCaseCurl(cmd string) (CaseCurl, error) {
	caseCurl := make(CaseCurl)
	var err error
	caseCurl, err = parseCaseCurl(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "load curl command failed")
	}
	// deal with option alias, turn all options to long form
	if err = caseCurl.toAlias(); err != nil {
		return nil, errors.Wrap(err, "identify curl option alias failed")
	}
	// check if caseCurl contains unsupported args
	if err = caseCurl.checkOptions(); err != nil {
		return nil, errors.Wrap(err, "check curl option failed")
	}
	caseCurl.Set(originCmdKey, cmd)
	return caseCurl, nil
}

// parseCaseCurl parses command string to map, save command keyword and bool option as map key only.
// Otherwise, save option as map key and the following args([]string) as map value
func parseCaseCurl(cmd string) (CaseCurl, error) {
	cmdWords, err := shlex.Split(cmd)
	if err != nil {
		return nil, err
	}

	// parse the command string to map
	res := make(CaseCurl)
	var i int
	if cmdWords[i] != "curl" {
		return nil, errors.New("command not started with curl")
	}
	i++
	for i < len(cmdWords) {
		if !strings.HasPrefix(cmdWords[i], "-") {
			// save target url
			res.Add(targetUrlKey, cmdWords[i])
			i++
			continue
		}
		option := cmdWords[i]
		i++
		if i < len(cmdWords) && !strings.HasPrefix(cmdWords[i], "-") {
			// option with only one following argument
			res.Add(option, cmdWords[i])
			i++
			continue
		}
		// option with no argument, i.e. bool option, save key only
		res[option] = nil
	}
	return res, nil
}

type CaseCurl map[string][]string

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns the empty string.
func (c CaseCurl) Get(key string, index int) string {
	if c == nil {
		return ""
	}
	vs := c[key]
	if index >= 0 && index < len(vs) {
		return vs[index]
	}
	return ""
}

func (c CaseCurl) Set(key, value string) {
	c[key] = []string{value}
}

func (c CaseCurl) Add(key, value string) {
	c[key] = append(c[key], value)
}

// HaveKey checks key existed or not
func (c CaseCurl) HaveKey(key string) bool {
	if c == nil {
		return false
	}
	_, ok := c[key]
	return ok
}

func (c CaseCurl) toAlias() error {
	for option, args := range c {
		if !strings.HasPrefix(option, "-") || strings.HasPrefix(option, "--") {
			// not a short option like -X, pass
			continue
		}
		longOption, ok := curlOptionAliasMap[option]
		if !ok {
			return errors.Errorf("unexpected curl option: %v", option)
		}
		// FIXME: need to copy args or not?
		c[longOption] = args
		delete(c, option)
	}
	return nil
}

func (c CaseCurl) checkOptions() error {
	for option := range c {
		if option == originCmdKey || option == targetUrlKey {
			continue
		}
		_, ok := curlOptionWhiteMap[option]
		if !ok {
			return errors.Errorf("option %v not supported yet. available options: %v", option, curlOptionWhiteList)
		}
	}
	return nil
}

func (c CaseCurl) toTSteps() ([]*hrp.TStep, error) {
	var tSteps []*hrp.TStep
	for _, rawUrl := range c[targetUrlKey] {
		log.Info().
			Str("url", rawUrl).
			Msg("convert test steps")

		step := &stepFromCurl{
			TStep: &hrp.TStep{
				Request: &hrp.Request{},
			},
		}
		if err := step.makeRequestName(c); err != nil {
			return nil, err
		}
		if err := step.makeRequestMethod(c); err != nil {
			return nil, err
		}
		if err := step.makeRequestURL(rawUrl); err != nil {
			return nil, err
		}
		if err := step.makeRequestParams(rawUrl); err != nil {
			return nil, err
		}
		if err := step.makeRequestHeaders(c); err != nil {
			return nil, err
		}
		if err := step.makeRequestCookies(c); err != nil {
			return nil, err
		}
		if err := step.makeRequestBody(c); err != nil {
			return nil, err
		}
		tSteps = append(tSteps, step.TStep)
	}
	return tSteps, nil
}

type stepFromCurl struct {
	*hrp.TStep
}

func (s *stepFromCurl) makeRequestName(c CaseCurl) error {
	s.Name = c.Get(originCmdKey, 0)
	return nil
}

func (s *stepFromCurl) makeRequestMethod(c CaseCurl) error {
	// default --get
	s.Request.Method = http.MethodGet
	if c.HaveKey("--data") || c.HaveKey("--form") {
		s.Request.Method = http.MethodPost
	}
	if c.HaveKey("--head") {
		s.Request.Method = http.MethodHead
	}
	if c.HaveKey("--request") {
		s.Request.Method = hrp.HTTPMethod(strings.ToUpper(c.Get("--request", 0)))
	}
	return nil
}

func (s *stepFromCurl) makeRequestURL(rawUrl string) error {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return errors.Wrap(err, "parse URL error")
	}
	// default protocol consistent with curl (http)
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	s.Request.URL = fmt.Sprintf("%s://%s", u.Scheme, u.Host+u.Path)
	return nil
}

func (s *stepFromCurl) makeRequestParams(rawUrl string) error {
	s.Request.Params = make(map[string]interface{})
	u, err := url.Parse(rawUrl)
	if err != nil {
		return errors.Wrap(err, "parse URL error")
	}
	s.Request.Params = make(map[string]interface{})
	queryValues := u.Query()
	// query key may correspond to more than one value, get first query key only
	for k := range queryValues {
		s.Request.Params[k] = queryValues.Get(k)
	}
	return nil
}

func (s *stepFromCurl) makeRequestHeaders(c CaseCurl) error {
	s.Request.Headers = make(map[string]string)
	headerList := c["--header"]
	for _, headerExpr := range headerList {
		if err := s.makeRequestHeader(headerExpr); err != nil {
			return err
		}
	}
	return nil
}

func (s *stepFromCurl) makeRequestHeader(headerExpr string) error {
	headerExpr = strings.TrimSpace(headerExpr)
	if strings.HasPrefix(headerExpr, "@") {
		return errors.Errorf("loading header from file not supported: %v", headerExpr)
	}
	if strings.TrimSpace(headerExpr) == ";" || strings.HasPrefix(strings.TrimSpace(headerExpr), ":") {
		return errors.Errorf("invalid curl header format: %v", headerExpr)
	}
	if s.Request.Headers == nil {
		s.Request.Headers = make(map[string]string)
	}
	if i := strings.Index(headerExpr, ":"); i != -1 {
		headerKey := strings.TrimSpace(headerExpr[:i])
		var headerValue string
		if i < len(headerExpr)-1 {
			headerValue = strings.TrimSpace(headerExpr[i+1:])
		}
		if strings.ToLower(headerKey) == "host" {
			// headerExpr modifying internal header like "Host:"
			log.Warn().Str("--header", headerExpr).Msg("modifying internal header not supported")
			return nil
		}
		if headerValue != "" {
			// normal headerExpr like "User-Agent: httprunner"
			s.Request.Headers[headerKey] = headerValue
			return nil
		}
	}
	if i := strings.Index(headerExpr, ";"); i != -1 {
		// headerExpr terminated with a semicolon like "X-Custom-Header;"
		headerKey := strings.TrimSpace(headerExpr[:i])
		if strings.ToLower(headerKey) == "host" {
			log.Warn().Str("--header", headerExpr).Msg("modifying internal header not supported")
			return nil
		}
		s.Request.Headers[headerKey] = ""
		return nil
	}
	log.Warn().Str("--header", headerExpr).Msg("pass meaningless curl header expression")
	return nil
}

func (s *stepFromCurl) makeRequestCookies(c CaseCurl) error {
	s.Request.Cookies = make(map[string]string)
	cookieList := c["--cookie"]
	for _, cookieExpr := range cookieList {
		if err := s.makeRequestCookie(cookieExpr); err != nil {
			return err
		}
	}
	return nil
}

func (s *stepFromCurl) makeRequestCookie(cookieExpr string) error {
	if !strings.Contains(cookieExpr, "=") {
		return errors.Errorf("loading cookie from file not supported: %v", cookieExpr)
	}
	if s.Request.Cookies == nil {
		s.Request.Cookies = make(map[string]string)
	}
	// deal with cookieExpr like "name1=value1;   name2 =  value2"
	cookies := strings.Split(cookieExpr, ";")
	for _, cookie := range cookies {
		i := strings.Index(cookie, "=")
		if i == -1 {
			log.Warn().Str("--cookie", cookie).Msg("pass meaningless curl cookie expression")
			continue
		}
		cookieKey := strings.TrimSpace(cookie[:i])
		var cookieValue string
		if i < len(cookie)-1 {
			cookieValue = strings.TrimSpace(cookie[i+1:])
		}
		s.Request.Cookies[cookieKey] = cookieValue
	}
	return nil
}

func (s *stepFromCurl) makeRequestBody(c CaseCurl) error {
	// check priority: --data > --form
	dataList, dataExisted := c["--data"]
	formList, formExisted := c["--form"]
	if dataExisted {
		if err := s.makeRequestData(dataList); err != nil {
			return err
		}
	} else if formExisted {
		if err := s.makeRequestForm(formList); err != nil {
			return err
		}
	}
	return nil
}

func (s *stepFromCurl) makeRequestData(dataList []string) error {
	dataMap := make(map[string]interface{})
	for _, dataExpr := range dataList {
		if strings.HasPrefix(dataExpr, "@") {
			return errors.Errorf("loading data from file not supported: %v", dataExpr)
		}
		var m map[string]interface{}
		// --data may be json string, try to unmarshal to map first
		err := json.Unmarshal([]byte(dataExpr), &m)
		if err == nil {
			for k, v := range m {
				dataMap[k] = v
			}
			continue
		}
		dataValues, err := url.ParseQuery(dataExpr)
		if err != nil {
			return err
		}
		for dataKey := range dataValues {
			dataMap[dataKey] = strings.Trim(dataValues.Get(dataKey), "\"'")
		}
	}
	s.Request.Body = dataMap
	return nil
}

func (s *stepFromCurl) makeRequestForm(formList []string) error {
	if s.Request.Upload == nil {
		s.Request.Upload = make(map[string]interface{})
	}
	for _, formExpr := range formList {
		if !strings.Contains(formExpr, "=") {
			return errors.Errorf("option --form: is badly used: %v", formExpr)
		}
		if i := strings.Index(formExpr, "="); i != -1 {
			formKey := strings.TrimSpace(formExpr[:i])
			var formValue string
			if i < len(formExpr)-1 {
				formValue = strings.TrimSpace(formExpr[i+1:])
			}
			s.Request.Upload[formKey] = strings.Trim(formValue, "\"")
		}
	}
	return nil
}
