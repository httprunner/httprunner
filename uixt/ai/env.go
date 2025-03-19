package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	defaultTimeout = 60 * time.Second
)

type OpenAIInitConfig struct {
	ReportURL string            `json:"REPORT_SERVER_URL"`
	Headers   map[string]string `json:"defaultHeaders"`
}

const (
	EnvOpenAIBaseURL        = "OPENAI_BASE_URL"
	EnvOpenAIAPIKey         = "OPENAI_API_KEY"
	EnvModelName            = "LLM_MODEL_NAME"
	EnvOpenAIInitConfigJSON = "OPENAI_INIT_CONFIG_JSON"
)

func checkEnvLLM() error {
	openaiBaseURL := os.Getenv("OPENAI_BASE_URL")
	if openaiBaseURL == "" {
		return errors.Wrap(code.LLMEnvMissedError, "OPENAI_BASE_URL missed")
	}
	log.Info().Str("OPENAI_BASE_URL", openaiBaseURL).Msg("get env")
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		return errors.Wrap(code.LLMEnvMissedError, "OPENAI_API_KEY missed")
	}
	log.Info().Str("OPENAI_API_KEY", maskAPIKey(openaiAPIKey)).Msg("get env")
	modelName := os.Getenv("LLM_MODEL_NAME")
	if modelName == "" {
		return errors.Wrap(code.LLMEnvMissedError, "LLM_MODEL_NAME missed")
	}
	log.Info().Str("LLM_MODEL_NAME", modelName).Msg("get env")
	return nil
}

// loadEnv loads environment variables from a file
func loadEnv(envPath string) error {
	err := godotenv.Load(envPath)
	if err != nil {
		return err
	}

	log.Info().Str("path", envPath).Msg("load env success")
	return nil
}

func GetEnvConfig(key string) string {
	return os.Getenv(key)
}

func GetEnvConfigInJSON(key string) (map[string]interface{}, error) {
	value := GetEnvConfig(key)
	if value == "" {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func GetEnvConfigInBool(key string) bool {
	value := GetEnvConfig(key)
	if value == "" {
		return false
	}

	boolValue, _ := strconv.ParseBool(value)
	return boolValue
}

// GetEnvConfigOrDefault get env config or default value
func GetEnvConfigOrDefault(key, defaultValue string) string {
	value := GetEnvConfig(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetEnvConfigInInt(key string, defaultValue int) int {
	value := GetEnvConfig(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// CustomTransport is a custom RoundTripper that adds headers to every request
type CustomTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

// RoundTrip executes a single HTTP transaction and adds custom headers
func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}
	return c.Transport.RoundTrip(req)
}

// GetModelConfig get OpenAI config
func GetModelConfig() (*openai.ChatModelConfig, error) {
	envConfig := &OpenAIInitConfig{
		Headers: make(map[string]string),
	}

	// read from JSON config first
	jsonStr := GetEnvConfig(EnvOpenAIInitConfigJSON)
	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), envConfig); err != nil {
			return nil, err
		}
	}

	config := &openai.ChatModelConfig{
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &CustomTransport{
				Transport: http.DefaultTransport,
				Headers:   envConfig.Headers,
			},
		},
	}

	if baseURL := GetEnvConfig(EnvOpenAIBaseURL); baseURL != "" {
		config.BaseURL = baseURL
	} else {
		return nil, fmt.Errorf("miss env %s", EnvOpenAIBaseURL)
	}

	if apiKey := GetEnvConfig(EnvOpenAIAPIKey); apiKey != "" {
		config.APIKey = apiKey
	} else {
		return nil, fmt.Errorf("miss env %s", EnvOpenAIAPIKey)
	}

	if modelName := GetEnvConfig(EnvModelName); modelName != "" {
		config.Model = modelName
	} else {
		return nil, fmt.Errorf("miss env %s", EnvModelName)
	}

	// log config info
	log.Info().Str("model", config.Model).
		Str("baseURL", config.BaseURL).
		Str("apiKey", maskAPIKey(config.APIKey)).
		Str("timeout", defaultTimeout.String()).
		Msg("get model config")

	return config, nil
}

// maskAPIKey masks the API key
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "******"
	}

	return key[:4] + "******" + key[len(key)-4:]
}
