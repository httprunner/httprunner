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

func IS_UI_TARS(modelType LLMServiceType) bool {
	return modelType == DOUBAO_1_5_UI_TARS_250328 ||
		modelType == DOUBAO_1_5_UI_TARS_250428
}

const (
	DOUBAO_1_5_UI_TARS_250328             LLMServiceType = "doubao-1.5-ui-tars-250328"
	DOUBAO_1_5_UI_TARS_250428             LLMServiceType = "doubao-1.5-ui-tars-250428" // not support function calling and json response
	DOUBAO_1_5_THINKING_VISION_PRO_250428 LLMServiceType = "doubao-1.5-thinking-vision-pro-250428"
	OPENAI_GPT_4O                         LLMServiceType = "openai/gpt-4o"
	DEEPSEEK_R1_250528                    LLMServiceType = "deepseek-r1-250528"
)

func WithLLMService(modelType LLMServiceType) AIServiceOption {
	return func(opts *AIServiceOptions) {
		opts.LLMService = modelType
	}
}
