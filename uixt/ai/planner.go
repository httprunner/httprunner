package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Error types
var (
	ErrEmptyInstruction      = fmt.Errorf("user instruction is empty")
	ErrNoConversationHistory = fmt.Errorf("conversation history is empty")
	ErrInvalidImageData      = fmt.Errorf("invalid image data")
)

func NewPlanner(ctx context.Context) (*Planner, error) {
	config, err := GetModelConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI config: %w", err)
	}
	model, err := openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAI model: %w", err)
	}
	parser := NewActionParser(1000)
	return &Planner{
		ctx:    ctx,
		model:  model,
		parser: parser,
	}, nil
}

type Planner struct {
	ctx     context.Context
	model   model.ChatModel
	parser  *ActionParser
	history []*schema.Message // conversation history
}

// Call performs UI planning using Vision Language Model
func (p *Planner) Call(opts *PlanningOptions) (*PlanningResult, error) {
	// validate input parameters
	if err := validateInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate input parameters failed")
	}

	// prepare prompt
	if len(p.history) == 0 {
		// add system message
		systemPrompt := uiTarsPlanningPrompt + opts.UserInstruction
		p.history = []*schema.Message{
			{
				Role:    schema.System,
				Content: systemPrompt,
			},
		}
	}
	// append user image message
	p.appendConversationHistory(opts.Message)

	// call model service, generate response
	logRequest(p.history)
	log.Info().Msg("calling model service...")
	resp, err := p.model.Generate(p.ctx, p.history)
	if err != nil {
		return nil, fmt.Errorf("request model service failed: %w", err)
	}
	logResponse(resp)

	// parse result
	result, err := p.parseResult(resp, opts.Size)
	if err != nil {
		return nil, errors.Wrap(err, "parse result failed")
	}

	// append assistant message
	p.appendConversationHistory(&schema.Message{
		Role:    schema.Assistant,
		Content: result.ActionSummary,
	})

	return result, nil
}

func validateInput(opts *PlanningOptions) error {
	if opts.UserInstruction == "" {
		return ErrEmptyInstruction
	}

	if opts.Message == nil {
		return ErrNoConversationHistory
	}

	if opts.Message.Role == schema.User {
		// check MultiContent
		if len(opts.Message.MultiContent) > 0 {
			for _, content := range opts.Message.MultiContent {
				if content.Type == schema.ChatMessagePartTypeImageURL && content.ImageURL == nil {
					return ErrInvalidImageData
				}
			}
		}
	}

	return nil
}

func logRequest(messages []*schema.Message) {
	msgs := make([]*schema.Message, 0, len(messages))
	for _, message := range messages {
		msg := &schema.Message{
			Role: message.Role,
		}
		if message.Content != "" {
			msg.Content = message.Content
		} else if len(message.MultiContent) > 0 {
			for _, mc := range message.MultiContent {
				switch mc.Type {
				case schema.ChatMessagePartTypeImageURL:
					// Create a copy of the ImageURL to avoid modifying the original message
					imageURLCopy := *mc.ImageURL
					if strings.HasPrefix(imageURLCopy.URL, "data:image/") {
						imageURLCopy.URL = "<data:image/base64...>"
					}
					msg.MultiContent = append(msg.MultiContent, schema.ChatMessagePart{
						Type:     mc.Type,
						ImageURL: &imageURLCopy,
					})
				}
			}
		}
		msgs = append(msgs, msg)
	}
	log.Debug().Interface("messages", msgs).Msg("log request messages")
}

func logResponse(resp *schema.Message) {
	log.Info().Str("role", string(resp.Role)).
		Str("content", resp.Content).Msg("log response message")
}

// appendConversationHistory adds a message to the conversation history
func (p *Planner) appendConversationHistory(msg *schema.Message) {
	// for user image message:
	// - keep at most 4 user image messages
	// - delete the oldest user image message when the limit is reached
	if msg.Role == schema.User {
		// get all existing user messages
		userImgCount := 0
		firstUserImgIndex := -1

		// calculate the number of user messages and find the index of the first user message
		for i, item := range p.history {
			if item.Role == schema.User {
				userImgCount++
				if firstUserImgIndex == -1 {
					firstUserImgIndex = i
				}
			}
		}

		// if there are already 4 user messages, delete the first one before adding the new message
		if userImgCount >= 4 && firstUserImgIndex >= 0 {
			// delete the first user message
			p.history = append(
				p.history[:firstUserImgIndex],
				p.history[firstUserImgIndex+1:]...,
			)
		}
		// add the new user message to the history
		p.history = append(p.history, msg)
	}

	// for assistant message:
	// - keep at most the last 10 assistant messages
	if msg.Role == schema.Assistant {
		// add the new assistant message to the history
		p.history = append(p.history, msg)

		// if there are more than 10 assistant messages, remove the oldest ones
		assistantMsgCount := 0
		for i := len(p.history) - 1; i >= 0; i-- {
			if p.history[i].Role == schema.Assistant {
				assistantMsgCount++
				if assistantMsgCount > 10 {
					p.history = append(p.history[:i], p.history[i+1:]...)
				}
			}
		}
	}
}

func (p *Planner) parseResult(msg *schema.Message, size types.Size) (*PlanningResult, error) {
	// parse response
	parseActions, err := p.parser.Parse(msg.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse actions: %w", err)
	}

	// process response
	result, err := processVLMResponse(parseActions, size)
	if err != nil {
		return nil, errors.Wrap(err, "process VLM response failed")
	}

	log.Info().
		Interface("summary", result.ActionSummary).
		Interface("actions", result.NextActions).
		Msg("get VLM planning result")
	return result, nil
}

// processVLMResponse processes the VLM response and converts it to PlanningResult
func processVLMResponse(actions []ParsedAction, size types.Size) (*PlanningResult, error) {
	log.Info().Msg("processing VLM response...")

	if len(actions) == 0 {
		return nil, fmt.Errorf("no actions returned from VLM")
	}

	// validate and post-process each action
	for i := range actions {
		// validate action type
		switch actions[i].ActionType {
		case "click":
			if err := convertCoordinateAction(&actions[i], "startBox", size); err != nil {
				return nil, errors.Wrap(err, "convert coordinate action failed")
			}
		case "drag":
			if err := convertCoordinateAction(&actions[i], "startBox", size); err != nil {
				return nil, errors.Wrap(err, "convert coordinate action failed")
			}
			if err := convertCoordinateAction(&actions[i], "endBox", size); err != nil {
				return nil, errors.Wrap(err, "convert coordinate action failed")
			}
		case "type":
			validateTypeContent(&actions[i])
		case "wait", "finished", "call_user":
			// these actions do not need extra parameters
		default:
			log.Printf("warning: unknown action type: %s, will try to continue processing", actions[i].ActionType)
		}
	}

	// extract action summary
	actionSummary := extractActionSummary(actions)

	return &PlanningResult{
		NextActions:   actions,
		ActionSummary: actionSummary,
	}, nil
}

// extractActionSummary extracts the summary from the actions
func extractActionSummary(actions []ParsedAction) string {
	if len(actions) == 0 {
		return ""
	}

	// use the Thought of the first action as summary
	if actions[0].Thought != "" {
		return actions[0].Thought
	}

	// if no Thought, generate summary from action type
	action := actions[0]
	switch action.ActionType {
	case "click":
		return "点击操作"
	case "drag":
		return "拖拽操作"
	case "type":
		content, _ := action.ActionInputs["content"].(string)
		if len(content) > 20 {
			content = content[:20] + "..."
		}
		return fmt.Sprintf("输入文本: %s", content)
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

func convertCoordinateAction(action *ParsedAction, boxField string, size types.Size) error {
	// The model generates a 2D coordinate output that represents relative positions.
	// To convert these values to image-relative coordinates, divide each component by 1000 to obtain values in the range [0,1].
	// The absolute coordinates required by the Action can be calculated by:
	// - X absolute = X relative × image width / 1000
	// - Y absolute = Y relative × image height / 1000

	// get image width and height
	imageWidth := size.Width
	imageHeight := size.Height

	box := action.ActionInputs[boxField]
	coords, ok := box.([]float64)
	if !ok {
		log.Error().Interface("inputs", action.ActionInputs).Msg("invalid action inputs")
		return fmt.Errorf("invalid action inputs")
	}

	if len(coords) == 2 {
		coords[0] = math.Round((coords[0]/1000*float64(imageWidth))*10) / 10
		coords[1] = math.Round((coords[1]/1000*float64(imageHeight))*10) / 10
	} else if len(coords) == 4 {
		coords[0] = math.Round((coords[0]/1000*float64(imageWidth))*10) / 10
		coords[1] = math.Round((coords[1]/1000*float64(imageHeight))*10) / 10
		coords[2] = math.Round((coords[2]/1000*float64(imageWidth))*10) / 10
		coords[3] = math.Round((coords[3]/1000*float64(imageHeight))*10) / 10
	} else {
		log.Error().Interface("inputs", action.ActionInputs).Msg("invalid action inputs")
		return fmt.Errorf("invalid action inputs")
	}

	return nil
}

// validateTypeContent 验证输入文本内容
func validateTypeContent(action *ParsedAction) {
	if content, ok := action.ActionInputs["content"]; !ok || content == "" {
		// default to empty string
		action.ActionInputs["content"] = ""
		log.Warn().Msg("type action missing content parameter, set to default")
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
func loadImage(imagePath string) (base64Str string, size types.Size, err error) {
	// Read the image file
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return "", types.Size{}, fmt.Errorf("failed to open image file: %w", err)
	}
	defer imageFile.Close()

	// Decode the image to get its resolution
	imageData, format, err := image.Decode(imageFile)
	if err != nil {
		return "", types.Size{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get the resolution of the image
	width := imageData.Bounds().Dx()
	height := imageData.Bounds().Dy()
	size = types.Size{Width: width, Height: height}

	// Convert image to base64
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, imageData); err != nil {
		return "", types.Size{}, fmt.Errorf("failed to encode image to buffer: %w", err)
	}
	base64Str = fmt.Sprintf("data:image/%s;base64,%s", format,
		base64.StdEncoding.EncodeToString(buf.Bytes()))

	return base64Str, size, nil
}
