package simulation

import (
	"math/rand"
	"time"
	"unicode"
)

// InputRequest 输入请求参数
type InputRequest struct {
	Text          string `json:"text"`         // 输入文本
	MinSegmentLen int    `json:"min_segment"`  // 最小分割长度
	MaxSegmentLen int    `json:"max_segment"`  // 最大分割长度
	MinDelayMs    int    `json:"min_delay_ms"` // 最小延迟时间(毫秒)
	MaxDelayMs    int    `json:"max_delay_ms"` // 最大延迟时间(毫秒)
}

// InputResponse 输入响应结果
type InputResponse struct {
	Success  bool           `json:"success"`
	Message  string         `json:"message,omitempty"`
	Segments []InputSegment `json:"segments"`
	Metrics  InputMetrics   `json:"metrics"`
}

// InputSegment 输入片段
type InputSegment struct {
	Index   int    `json:"index"`    // 片段索引
	Text    string `json:"text"`     // 片段文本
	DelayMs int    `json:"delay_ms"` // 该片段后的延迟时间(毫秒)
	CharLen int    `json:"char_len"` // 字符长度
}

// InputMetrics 输入指标
type InputMetrics struct {
	TotalSegments   int `json:"total_segments"`    // 总片段数
	TotalDelayMs    int `json:"total_delay_ms"`    // 总延迟时间
	EstimatedTimeMs int `json:"estimated_time_ms"` // 预估总耗时
	OriginalCharLen int `json:"original_char_len"` // 原始字符长度
}

// InputConfig 输入配置参数
type InputConfig struct {
	MinSegmentLen int // 最小分割长度(字符数)
	MaxSegmentLen int // 最大分割长度(字符数)
	MinDelayMs    int // 最小延迟时间(毫秒)
	MaxDelayMs    int // 最大延迟时间(毫秒)
}

// DefaultInputConfig 默认输入配置
var DefaultInputConfig = InputConfig{
	MinSegmentLen: 1,   // 1个字符
	MaxSegmentLen: 4,   // 4个字符
	MinDelayMs:    50,  // 50毫秒
	MaxDelayMs:    200, // 200毫秒
}

// InputSimulatorAPI 输入仿真API
type InputSimulatorAPI struct {
	rand   *rand.Rand
	config InputConfig
}

// NewInputSimulatorAPI 创建新的输入仿真API
func NewInputSimulatorAPI(config *InputConfig) *InputSimulatorAPI {
	if config == nil {
		config = &DefaultInputConfig
	}

	return &InputSimulatorAPI{
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		config: *config,
	}
}

// GenerateInputSegments 生成输入片段序列
func (api *InputSimulatorAPI) GenerateInputSegments(req InputRequest) InputResponse {
	// 验证输入参数
	if err := api.validateRequest(req); err != nil {
		return InputResponse{
			Success: false,
			Message: err.Error(),
		}
	}

	// 如果文本为空，直接返回
	if req.Text == "" {
		return InputResponse{
			Success:  true,
			Segments: []InputSegment{},
			Metrics: InputMetrics{
				TotalSegments:   0,
				TotalDelayMs:    0,
				EstimatedTimeMs: 0,
				OriginalCharLen: 0,
			},
		}
	}

	// 生成分割片段
	segments := api.splitTextIntelligently(req.Text, req.MinSegmentLen, req.MaxSegmentLen)

	// 生成延迟时间
	inputSegments := make([]InputSegment, len(segments))
	totalDelayMs := 0

	for i, segment := range segments {
		var delayMs int
		// 最后一个片段不需要延迟
		if i < len(segments)-1 {
			delayMs = api.generateRandomDelay(req.MinDelayMs, req.MaxDelayMs)
			totalDelayMs += delayMs
		}

		inputSegments[i] = InputSegment{
			Index:   i,
			Text:    segment,
			DelayMs: delayMs,
			CharLen: len([]rune(segment)),
		}
	}

	// 计算指标
	metrics := InputMetrics{
		TotalSegments:   len(segments),
		TotalDelayMs:    totalDelayMs,
		EstimatedTimeMs: totalDelayMs, // 简化计算，实际输入时间可能更长
		OriginalCharLen: len([]rune(req.Text)),
	}

	return InputResponse{
		Success:  true,
		Segments: inputSegments,
		Metrics:  metrics,
	}
}

// validateRequest 验证请求参数
func (api *InputSimulatorAPI) validateRequest(req InputRequest) error {
	// 使用配置中的默认值填充请求参数
	if req.MinSegmentLen <= 0 {
		req.MinSegmentLen = api.config.MinSegmentLen
	}
	if req.MaxSegmentLen <= 0 {
		req.MaxSegmentLen = api.config.MaxSegmentLen
	}
	if req.MinDelayMs < 0 {
		req.MinDelayMs = api.config.MinDelayMs
	}
	if req.MaxDelayMs < 0 {
		req.MaxDelayMs = api.config.MaxDelayMs
	}

	return nil
}

// splitTextIntelligently 智能分割文本
// 规则：
// 1. 先分解成基础单元：中文每个字符一个单元，英文/数字连续的作为一个单元，其他字符各自一个单元
// 2. 按MinSegmentLen到MaxSegmentLen的随机值组合基础单元
func (api *InputSimulatorAPI) splitTextIntelligently(text string, minLen, maxLen int) []string {
	if minLen <= 0 {
		minLen = api.config.MinSegmentLen
	}
	if maxLen <= 0 {
		maxLen = api.config.MaxSegmentLen
	}
	if maxLen < minLen {
		maxLen = minLen
	}

	// 第一步：分解成基础单元
	baseUnits := api.splitIntoBaseUnits(text)

	// 第二步：按随机数组合基础单元
	var segments []string
	i := 0

	for i < len(baseUnits) {
		remainingUnits := len(baseUnits) - i

		var unitCount int
		// 如果剩余单元数少于minLen，就把剩余的全部作为一个片段
		if remainingUnits < minLen {
			unitCount = remainingUnits
		} else {
			// 随机决定本次要组合的单元数量（在minLen到maxLen之间）
			unitCount = minLen
			if maxLen > minLen {
				// 确保unitCount不超过剩余单元数
				maxPossibleCount := maxLen
				if maxPossibleCount > remainingUnits {
					maxPossibleCount = remainingUnits
				}
				unitCount = minLen + api.rand.Intn(maxPossibleCount-minLen+1)
			}
		}

		// 组合unitCount个基础单元成一个片段
		segment := ""
		for j := 0; j < unitCount; j++ {
			segment += baseUnits[i+j]
		}
		segments = append(segments, segment)
		i += unitCount
	}

	return segments
}

// splitIntoBaseUnits 将文本分解成基础单元
func (api *InputSimulatorAPI) splitIntoBaseUnits(text string) []string {
	var units []string
	runes := []rune(text)
	i := 0

	for i < len(runes) {
		// 处理中文字符：每个字符一个单元
		if api.isChinese(runes[i]) {
			units = append(units, string(runes[i]))
			i++
			continue
		}

		// 处理连续英文字母：作为一个单元
		if unicode.IsLetter(runes[i]) && runes[i] <= 127 {
			start := i
			for i < len(runes) && unicode.IsLetter(runes[i]) && runes[i] <= 127 {
				i++
			}
			word := string(runes[start:i])
			units = append(units, word)
			continue
		}

		// 处理连续数字：作为一个单元
		if unicode.IsDigit(runes[i]) {
			start := i
			for i < len(runes) && unicode.IsDigit(runes[i]) {
				i++
			}
			number := string(runes[start:i])
			units = append(units, number)
			continue
		}

		// 处理其他字符（空格、标点等）：每个字符一个单元
		units = append(units, string(runes[i]))
		i++
	}

	return units
}

// isChinese 判断字符是否为中文
func (api *InputSimulatorAPI) isChinese(r rune) bool {
	return unicode.Is(unicode.Scripts["Han"], r)
}

// generateRandomDelay 生成随机延迟时间
func (api *InputSimulatorAPI) generateRandomDelay(minDelayMs, maxDelayMs int) int {
	if minDelayMs < 0 {
		minDelayMs = api.config.MinDelayMs
	}
	if maxDelayMs < 0 {
		maxDelayMs = api.config.MaxDelayMs
	}
	if maxDelayMs < minDelayMs {
		maxDelayMs = minDelayMs
	}

	if maxDelayMs == minDelayMs {
		return minDelayMs
	}

	return minDelayMs + api.rand.Intn(maxDelayMs-minDelayMs+1)
}

// SplitText 公开的文本分割函数（使用智能分割）
func (api *InputSimulatorAPI) SplitText(text string) []string {
	return api.splitTextIntelligently(text, api.config.MinSegmentLen, api.config.MaxSegmentLen)
}

// GenerateDelay 公开的延迟生成函数
func (api *InputSimulatorAPI) GenerateDelay() int {
	return api.generateRandomDelay(api.config.MinDelayMs, api.config.MaxDelayMs)
}
