package option

func NewAIServiceOptions(opts ...AIServiceOption) *AIServiceOptions {
	services := &AIServiceOptions{}
	for _, option := range opts {
		option(services)
	}
	return services
}

type AIServiceOptions struct {
	CVService  CVServiceType     `json:"cv_service,omitempty" yaml:"cv_service,omitempty"`
	LLMService LLMServiceType    `json:"llm_service,omitempty" yaml:"llm_service,omitempty"`
	LLMConfig  *LLMServiceConfig `json:"llm_config,omitempty" yaml:"llm_config,omitempty"` // advanced LLM configuration
}

func (opts *AIServiceOptions) Options() []AIServiceOption {
	aiOpts := []AIServiceOption{}
	if opts.CVService != "" {
		aiOpts = append(aiOpts, WithCVService(opts.CVService))
	}
	if opts.LLMService != "" {
		aiOpts = append(aiOpts, WithLLMService(opts.LLMService))
	}
	if opts.LLMConfig != nil {
		aiOpts = append(aiOpts, WithLLMConfig(opts.LLMConfig))
	}
	return aiOpts
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

// UI-TARS do not support function calling and json response
func IS_UI_TARS(modelType LLMServiceType) bool {
	return modelType == DOUBAO_1_5_UI_TARS_250328 ||
		modelType == DOUBAO_1_5_UI_TARS_250428
}

const (
	DOUBAO_1_5_UI_TARS_250328             LLMServiceType = "doubao-1.5-ui-tars-250328"
	DOUBAO_1_5_UI_TARS_250428             LLMServiceType = "doubao-1.5-ui-tars-250428"
	DOUBAO_1_5_THINKING_VISION_PRO_250428 LLMServiceType = "doubao-1.5-thinking-vision-pro-250428"
	DOUBAO_SEED_1_6_250615                LLMServiceType = "doubao-seed-1.6-250615"
	OPENAI_GPT_4O                         LLMServiceType = "openai/gpt-4o"
	DEEPSEEK_R1_250528                    LLMServiceType = "deepseek-r1-250528"
	WINGS_SERVICE                         LLMServiceType = "wings-service"
)

func WithLLMService(modelType LLMServiceType) AIServiceOption {
	return func(opts *AIServiceOptions) {
		opts.LLMService = modelType
	}
}

// LLMServiceConfig defines configuration for different LLM service components
type LLMServiceConfig struct {
	PlannerModel  LLMServiceType `json:"planner_model"`  // Model type for planner component
	AsserterModel LLMServiceType `json:"asserter_model"` // Model type for asserter component
	QuerierModel  LLMServiceType `json:"querier_model"`  // Model type for querier component
}

// NewLLMServiceConfig creates a new LLMServiceConfig with the same model for all components
func NewLLMServiceConfig(modelType LLMServiceType) *LLMServiceConfig {
	return &LLMServiceConfig{
		PlannerModel:  modelType,
		AsserterModel: modelType,
		QuerierModel:  modelType,
	}
}

// WithPlannerModel sets the model type for planner component
func (c *LLMServiceConfig) WithPlannerModel(modelType LLMServiceType) *LLMServiceConfig {
	c.PlannerModel = modelType
	return c
}

// WithAsserterModel sets the model type for asserter component
func (c *LLMServiceConfig) WithAsserterModel(modelType LLMServiceType) *LLMServiceConfig {
	c.AsserterModel = modelType
	return c
}

// WithQuerierModel sets the model type for querier component
func (c *LLMServiceConfig) WithQuerierModel(modelType LLMServiceType) *LLMServiceConfig {
	c.QuerierModel = modelType
	return c
}

// WithLLMConfig sets the advanced LLM configuration
func WithLLMConfig(config *LLMServiceConfig) AIServiceOption {
	return func(opts *AIServiceOptions) {
		opts.LLMConfig = config
	}
}

// RecommendedConfigurations provides some recommended model configurations for different use cases
func RecommendedConfigurations() map[string]*LLMServiceConfig {
	return map[string]*LLMServiceConfig{
		"cost_effective": NewLLMServiceConfig(DOUBAO_1_5_THINKING_VISION_PRO_250428).
			WithPlannerModel(DOUBAO_1_5_UI_TARS_250328).
			WithAsserterModel(DOUBAO_1_5_THINKING_VISION_PRO_250428).
			WithQuerierModel(DOUBAO_1_5_THINKING_VISION_PRO_250428),

		"high_performance": NewLLMServiceConfig(OPENAI_GPT_4O),

		"mixed_optimal": NewLLMServiceConfig(DOUBAO_1_5_THINKING_VISION_PRO_250428).
			WithPlannerModel(DOUBAO_1_5_UI_TARS_250328). // Best for UI understanding
			WithAsserterModel(OPENAI_GPT_4O).            // Best for reasoning
			WithQuerierModel(DEEPSEEK_R1_250528),        // Cost-effective for queries

		"ui_focused": NewLLMServiceConfig(DOUBAO_1_5_UI_TARS_250328),

		"reasoning_focused": NewLLMServiceConfig(DOUBAO_1_5_THINKING_VISION_PRO_250428),
	}
}
