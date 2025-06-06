package ai

import (
	"os"
	"testing"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetServiceEnvPrefix(t *testing.T) {
	tests := []struct {
		name           string
		modelType      option.LLMServiceType
		expectedPrefix string
	}{
		{
			name:           "doubao thinking vision pro",
			modelType:      option.DOUBAO_1_5_THINKING_VISION_PRO_250428,
			expectedPrefix: "DOUBAO_1_5_THINKING_VISION_PRO_250428",
		},
		{
			name:           "doubao ui tars",
			modelType:      option.DOUBAO_1_5_UI_TARS_250428,
			expectedPrefix: "DOUBAO_1_5_UI_TARS_250428",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := getServiceEnvPrefix(tt.modelType)
			assert.Equal(t, tt.expectedPrefix, prefix)
		})
	}
}

func TestGetModelConfigFromEnv_ServiceSpecific(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("DOUBAO_1_5_THINKING_VISION_PRO_250428_BASE_URL")
		os.Unsetenv("DOUBAO_1_5_THINKING_VISION_PRO_250428_API_KEY")
	}()

	// Set service-specific environment variables (no need for MODEL_NAME)
	os.Setenv("DOUBAO_1_5_THINKING_VISION_PRO_250428_BASE_URL", "https://test-base-url.com")
	os.Setenv("DOUBAO_1_5_THINKING_VISION_PRO_250428_API_KEY", "test-api-key")

	baseURL, apiKey, modelName, err := getModelConfigFromEnv(option.DOUBAO_1_5_THINKING_VISION_PRO_250428)

	require.NoError(t, err)
	assert.Equal(t, "https://test-base-url.com", baseURL)
	assert.Equal(t, "test-api-key", apiKey)
	assert.Equal(t, "doubao-1.5-thinking-vision-pro-250428", modelName) // Model name derived from service type
}

func TestGetModelConfigFromEnv_FallbackToDefault(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("OPENAI_BASE_URL")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("LLM_MODEL_NAME")
		// Ensure service-specific vars are not set
		os.Unsetenv("DOUBAO_1_5_THINKING_VISION_PRO_250428_BASE_URL")
		os.Unsetenv("DOUBAO_1_5_THINKING_VISION_PRO_250428_API_KEY")
	}()

	// Set default environment variables
	os.Setenv("OPENAI_BASE_URL", "https://default-base-url.com")
	os.Setenv("OPENAI_API_KEY", "default-api-key")
	os.Setenv("LLM_MODEL_NAME", "default-model-name")

	baseURL, apiKey, modelName, err := getModelConfigFromEnv(option.DOUBAO_1_5_THINKING_VISION_PRO_250428)

	require.NoError(t, err)
	assert.Equal(t, "https://default-base-url.com", baseURL)
	assert.Equal(t, "default-api-key", apiKey)
	assert.Equal(t, "default-model-name", modelName) // Uses default model name when falling back to default config
}

func TestGetModelConfigFromEnv_MixedConfig(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("DOUBAO_1_5_THINKING_VISION_PRO_250428_BASE_URL")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("LLM_MODEL_NAME")
	}()

	// Set mixed configuration: service-specific base URL, default API key
	os.Setenv("DOUBAO_1_5_THINKING_VISION_PRO_250428_BASE_URL", "https://service-specific-url.com")
	os.Setenv("OPENAI_API_KEY", "default-api-key")
	os.Setenv("LLM_MODEL_NAME", "default-model-name")

	baseURL, apiKey, modelName, err := getModelConfigFromEnv(option.DOUBAO_1_5_THINKING_VISION_PRO_250428)

	require.NoError(t, err)
	assert.Equal(t, "https://service-specific-url.com", baseURL)        // Service-specific
	assert.Equal(t, "default-api-key", apiKey)                          // Default fallback
	assert.Equal(t, "doubao-1.5-thinking-vision-pro-250428", modelName) // Service type derived model name
}

func TestGetModelConfigFromEnv_MissingConfig(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("DOUBAO_1_5_THINKING_VISION_PRO_250428_BASE_URL")
		os.Unsetenv("DOUBAO_1_5_THINKING_VISION_PRO_250428_API_KEY")
		os.Unsetenv("OPENAI_BASE_URL")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("LLM_MODEL_NAME")
	}()

	// Test missing base URL
	os.Setenv("OPENAI_API_KEY", "test-api-key")

	_, _, _, err := getModelConfigFromEnv(option.DOUBAO_1_5_THINKING_VISION_PRO_250428)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "BASE_URL")

	// Test missing API key
	os.Unsetenv("OPENAI_API_KEY")
	os.Setenv("OPENAI_BASE_URL", "https://test-url.com")

	_, _, _, err = getModelConfigFromEnv(option.DOUBAO_1_5_THINKING_VISION_PRO_250428)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API_KEY")

	// Test with both base URL and API key present - should succeed
	os.Setenv("OPENAI_API_KEY", "test-api-key")

	baseURL, apiKey, modelName, err := getModelConfigFromEnv(option.DOUBAO_1_5_THINKING_VISION_PRO_250428)
	assert.NoError(t, err)
	assert.Equal(t, "https://test-url.com", baseURL)
	assert.Equal(t, "test-api-key", apiKey)
	assert.Equal(t, "doubao-1.5-thinking-vision-pro-250428", modelName) // Model name derived from service type
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "normal key",
			apiKey:   "sk-1234567890abcdef",
			expected: "sk-1******cdef",
		},
		{
			name:     "short key",
			apiKey:   "short",
			expected: "******",
		},
		{
			name:     "empty key",
			apiKey:   "",
			expected: "******",
		},
		{
			name:     "exactly 8 chars",
			apiKey:   "12345678",
			expected: "******",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskAPIKey(tt.apiKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}
