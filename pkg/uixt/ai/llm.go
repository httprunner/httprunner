package ai

import "context"

type ILLMService interface {
	Call(ctx context.Context, prompt string) (string, error)
}

func NewGPT4oLLMService() (*openaiLLMService, error) {
	if err := checkEnv(); err != nil {
		return nil, err
	}
	return &openaiLLMService{}, nil
}

type openaiLLMService struct{}

func (s openaiLLMService) Call(ctx context.Context, prompt string) (string, error) {
	return "", nil
}
