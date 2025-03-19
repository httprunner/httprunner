package ai

import "testing"

func TestOption(t *testing.T) {
	options := NewAIService(
		WithCVService(CVServiceTypeOpenCV),
		WithLLMService(LLMServiceTypeUITARS),
	)
	t.Log(options)
}
