package option

func NewAIServiceOptions(opts ...AIServiceOption) *AIServiceOptions {
	services := &AIServiceOptions{}
	for _, option := range opts {
		option(services)
	}
	return services
}

type AIServiceOptions struct {
	CVService  CVServiceType
	LLMService LLMServiceType
}

type AIServiceOption func(*AIServiceOptions)

type CVServiceType string

const (
	CVServiceTypeVEDEM  CVServiceType = "vedem"
	CVServiceTypeOpenCV CVServiceType = "opencv"
)

func WithCVService(service CVServiceType) AIServiceOption {
	return func(opts *AIServiceOptions) {
		opts.CVService = service
	}
}

type LLMServiceType string

const (
	LLMServiceTypeUITARS LLMServiceType = "ui-tars"
	LLMServiceTypeGPT    LLMServiceType = "gpt"
	LLMServiceTypeQwenVL LLMServiceType = "qwen-vl"
)

func WithLLMService(modelType LLMServiceType) AIServiceOption {
	return func(opts *AIServiceOptions) {
		opts.LLMService = modelType
	}
}
