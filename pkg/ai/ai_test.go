package ai

import "testing"

func TestOption(t *testing.T) {
	options := NewAIService(
		WithCVService(CVServiceTypeOpenCV),
		WithLLMService(LLMServiceTypeDeepSeekV3),
	)
	t.Log(options)
}
