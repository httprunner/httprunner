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
	DOUBAO_1_5_UI_TARS_250428             LLMServiceType = "doubao-1.5-ui-tars-250428" // not support function calling and json response
	DOUBAO_1_5_THINKING_VISION_PRO_250428 LLMServiceType = "doubao-1.5-thinking-vision-pro-250428"
)

func WithLLMService(modelType LLMServiceType) AIServiceOption {
	return func(opts *AIServiceOptions) {
		opts.LLMService = modelType
	}
}
