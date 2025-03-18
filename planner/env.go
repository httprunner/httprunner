package planner

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/joho/godotenv"
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
	EnvModelName            = "MIDSCENE_MODEL_NAME"
	EnvOpenAIInitConfigJSON = "MIDSCENE_OPENAI_INIT_CONFIG_JSON"
	EnvUseVLMUITars         = "MIDSCENE_USE_VLM_UI_TARS"
)

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
		Timeout: defaultTimeout,
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

func IsUseVLMUITars() bool {
	return GetEnvConfigInBool(EnvUseVLMUITars)
}
