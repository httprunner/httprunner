package planner

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Error types
var (
	ErrInvalidInput          = fmt.Errorf("invalid input parameters")
	ErrEmptyInstruction      = fmt.Errorf("user instruction is empty")
	ErrNoConversationHistory = fmt.Errorf("conversation history is empty")
	ErrInvalidImageData      = fmt.Errorf("invalid image data")
)

const uiTarsPlanningPrompt = `
You are a GUI agent. You are given a task and your action history, with screenshots. You need to perform the next action to complete the task.

## Output Format
Thought: ...
Action: ...

## Action Space
click(start_box='[x1, y1, x2, y2]')
left_double(start_box='[x1, y1, x2, y2]')
right_single(start_box='[x1, y1, x2, y2]')
drag(start_box='[x1, y1, x2, y2]', end_box='[x3, y3, x4, y4]')
hotkey(key='')
type(content='') #If you want to submit your input, use "\n" at the end of content.
scroll(start_box='[x1, y1, x2, y2]', direction='down or up or right or left')
wait() #Sleep for 5s and take a screenshot to check for any changes.
finished()
call_user() # Submit the task and call the user when the task is unsolvable, or when you need the user's help.

## Note
- Use Chinese in Thought part.
- Write a small plan and finally summarize your next action (with its target element) in one sentence in Thought part.

## User Instruction
`

func NewPlanner(ctx context.Context) (*Planner, error) {
	config, err := GetModelConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI config: %w", err)
	}
	model, err := openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAI model: %w", err)
	}
	return &Planner{
		ctx:   ctx,
		model: model,
	}, nil
}

type Planner struct {
	ctx   context.Context
	model *openai.ChatModel
}

// Start performs UI planning using Vision Language Model
func (p *Planner) Start(opts PlanningOptions) (*PlanningResult, error) {
	log.Info().Str("user_instruction", opts.UserInstruction).Msg("start VLM planning")

	// 1. validate input parameters
	if err := validateInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate input parameters failed")
	}

	// 2. call VLM service
	resp, err := p.callVLMService(opts)
	if err != nil {
		return nil, errors.Wrap(err, "call VLM service failed")
	}

	// 3. process response
	result, err := processVLMResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "process VLM response failed")
	}

	log.Info().
		Interface("summary", result.ActionSummary).
		Interface("actions", result.Actions).
		Msg("VLM planning completed")
	return result, nil
}

func validateInput(opts PlanningOptions) error {
	if opts.UserInstruction == "" {
		return ErrEmptyInstruction
	}

	if len(opts.ConversationHistory) == 0 {
		return ErrNoConversationHistory
	}

	if opts.Size.Width <= 0 || opts.Size.Height <= 0 {
		return ErrInvalidInput
	}

	// ensure at least one image URL
	hasImageURL := false
	for _, msg := range opts.ConversationHistory {
		if msg.Role == "user" {
			// check MultiContent
			if len(msg.MultiContent) > 0 {
				for _, content := range msg.MultiContent {
					if content.Type == "image_url" && content.ImageURL != nil {
						hasImageURL = true
						break
					}
				}
			}
		}
		if hasImageURL {
			break
		}
	}

	if !hasImageURL {
		return ErrInvalidInput
	}

	return nil
}

// callVLMService makes the actual call to the VLM service
func (p *Planner) callVLMService(opts PlanningOptions) (*VLMResponse, error) {
	log.Info().Msg("calling VLM service...")

	// prepare prompt
	systemPrompt := uiTarsPlanningPrompt + opts.UserInstruction
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
	}
	messages = append(messages, opts.ConversationHistory...)

	// generate response
	resp, err := p.model.Generate(p.ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}

	// parse response
	content := resp.Content
	parser := NewActionParser(content, 1000) // 使用与 TypeScript 版本相同的 factor
	actions, err := parser.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse actions: %w", err)
	}

	return &VLMResponse{
		Actions: actions,
	}, nil
}

// processVLMResponse processes the VLM response and converts it to PlanningResult
func processVLMResponse(resp *VLMResponse) (*PlanningResult, error) {
	log.Info().Msg("processing VLM response...")
	if resp.Error != "" {
		return nil, fmt.Errorf("VLM error: %s", resp.Error)
	}

	if len(resp.Actions) == 0 {
		return nil, fmt.Errorf("no actions returned from VLM")
	}

	// 验证和后处理每个动作
	for i := range resp.Actions {
		// 验证动作类型
		switch resp.Actions[i].ActionType {
		case "click", "left_double", "right_single":
			validateCoordinateAction(&resp.Actions[i], "startBox")
		case "drag":
			validateCoordinateAction(&resp.Actions[i], "startBox")
			validateCoordinateAction(&resp.Actions[i], "endBox")
		case "scroll":
			validateCoordinateAction(&resp.Actions[i], "startBox")
			validateScrollDirection(&resp.Actions[i])
		case "type":
			validateTypeContent(&resp.Actions[i])
		case "hotkey":
			validateHotkeyAction(&resp.Actions[i])
		case "wait", "finished", "call_user":
			// 这些动作不需要额外参数
		default:
			log.Printf("警告: 未知的动作类型: %s, 将尝试继续处理", resp.Actions[i].ActionType)
		}
	}

	// 提取动作摘要
	actionSummary := extractActionSummary(resp.Actions)

	// 将ParsedAction转换为接口类型
	var actions []interface{}
	for _, action := range resp.Actions {
		actionMap := map[string]interface{}{
			"actionType":   action.ActionType,
			"actionInputs": action.ActionInputs,
			"thought":      action.Thought,
		}
		actions = append(actions, actionMap)
	}

	return &PlanningResult{
		Actions:       actions,
		RealActions:   resp.Actions,
		ActionSummary: actionSummary,
	}, nil
}

// extractActionSummary 从动作中提取摘要
func extractActionSummary(actions []ParsedAction) string {
	if len(actions) == 0 {
		return ""
	}

	// 优先使用第一个动作的Thought作为摘要
	if actions[0].Thought != "" {
		return actions[0].Thought
	}

	// 如果没有Thought，则根据动作类型生成摘要
	action := actions[0]
	switch action.ActionType {
	case "click":
		return "点击操作"
	case "drag":
		return "拖拽操作"
	case "left_double":
		return "双击操作"
	case "right_single":
		return "右键点击操作"
	case "scroll":
		direction, _ := action.ActionInputs["direction"].(string)
		return fmt.Sprintf("滚动操作 (%s)", direction)
	case "type":
		content, _ := action.ActionInputs["content"].(string)
		if len(content) > 20 {
			content = content[:20] + "..."
		}
		return fmt.Sprintf("输入文本: %s", content)
	case "hotkey":
		key, _ := action.ActionInputs["key"].(string)
		return fmt.Sprintf("快捷键: %s", key)
	case "wait":
		return "等待操作"
	case "finished":
		return "完成操作"
	case "call_user":
		return "请求用户协助"
	default:
		return fmt.Sprintf("执行 %s 操作", action.ActionType)
	}
}

// validateCoordinateAction 验证坐标类动作
func validateCoordinateAction(action *ParsedAction, boxField string) {
	if box, ok := action.ActionInputs[boxField]; !ok || box == "" {
		// 为空或缺失的坐标设置默认值
		action.ActionInputs[boxField] = "[0.5, 0.5]"
		log.Printf("警告: %s动作缺少%s参数, 已设置默认值", action.ActionType, boxField)
	}
}

// validateScrollDirection 验证滚动方向
func validateScrollDirection(action *ParsedAction) {
	if direction, ok := action.ActionInputs["direction"].(string); !ok || direction == "" {
		// 为空或缺失的方向设置默认值
		action.ActionInputs["direction"] = "down"
		log.Printf("警告: scroll动作缺少direction参数, 已设置默认值")
	} else {
		// 标准化方向
		switch strings.ToLower(direction) {
		case "up", "down", "left", "right":
			// 保持原样
		default:
			// 非标准方向设为默认值
			action.ActionInputs["direction"] = "down"
			log.Printf("警告: 非标准滚动方向: %s, 已设置为down", direction)
		}
	}
}

// validateTypeContent 验证输入文本内容
func validateTypeContent(action *ParsedAction) {
	if content, ok := action.ActionInputs["content"]; !ok || content == "" {
		// 为空或缺失的内容设置默认值
		action.ActionInputs["content"] = ""
		log.Printf("警告: type动作缺少content参数, 已设置为空字符串")
	}
}

// validateHotkeyAction 验证快捷键动作
func validateHotkeyAction(action *ParsedAction) {
	if key, ok := action.ActionInputs["key"]; !ok || key == "" {
		// 为空或缺失的键设置默认值
		action.ActionInputs["key"] = "Enter"
		log.Printf("警告: hotkey动作缺少key参数, 已设置默认值")
	}
}

// SavePositionImg saves an image with position markers
func SavePositionImg(params struct {
	InputImgBase64 string
	Rect           struct {
		X float64
		Y float64
	}
	OutputPath string
}) error {
	// 解码Base64图像
	imgData := params.InputImgBase64
	// 如果包含了数据URL前缀，去掉它
	if strings.HasPrefix(imgData, "data:image/") {
		parts := strings.Split(imgData, ",")
		if len(parts) > 1 {
			imgData = parts[1]
		}
	}

	// 解码Base64
	unbased, err := base64.StdEncoding.DecodeString(imgData)
	if err != nil {
		return fmt.Errorf("无法解码Base64图像: %w", err)
	}

	// 解码图像
	reader := bytes.NewReader(unbased)
	img, _, err := image.Decode(reader)
	if err != nil {
		return fmt.Errorf("无法解码图像数据: %w", err)
	}

	// 创建一个可以在其上绘制的图像
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// 在点击/拖动位置绘制标记
	markRadius := 10
	x, y := int(params.Rect.X), int(params.Rect.Y)

	// 绘制红色圆圈
	for i := -markRadius; i <= markRadius; i++ {
		for j := -markRadius; j <= markRadius; j++ {
			if i*i+j*j <= markRadius*markRadius {
				if x+i >= 0 && x+i < bounds.Max.X && y+j >= 0 && y+j < bounds.Max.Y {
					rgba.Set(x+i, y+j, color.RGBA{255, 0, 0, 255})
				}
			}
		}
	}

	// 保存图像
	outFile, err := os.Create(params.OutputPath)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %w", err)
	}
	defer outFile.Close()

	// 编码为PNG并保存
	if err := png.Encode(outFile, rgba); err != nil {
		return fmt.Errorf("无法编码和保存图像: %w", err)
	}

	return nil
}

// loadImage loads image and returns base64 encoded string
func loadImage(imagePath string) (base64Str string, err error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", err
	}
	base64Str = "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData)
	return
}
