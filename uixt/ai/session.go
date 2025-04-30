package ai

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/rs/zerolog/log"
)

// ConversationHistory represents a sequence of chat messages
type ConversationHistory []*schema.Message

// Append adds a new message to the conversation history
func (h *ConversationHistory) Append(msg *schema.Message) {
	// for user image message:
	// - keep at most 4 user image messages
	// - delete the oldest user image message when the limit is reached
	if msg.Role == schema.User {
		// get all existing user messages
		userImgCount := 0
		firstUserImgIndex := -1

		// calculate the number of user messages and find the index of the first user message
		for i, item := range *h {
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
			*h = append(
				(*h)[:firstUserImgIndex],
				(*h)[firstUserImgIndex+1:]...,
			)
		}
		// add the new user message to the history
		*h = append(*h, msg)
	}

	// for assistant message:
	// - keep at most the last 10 assistant messages
	if msg.Role == schema.Assistant {
		// add the new assistant message to the history
		*h = append(*h, msg)

		// if there are more than 10 assistant messages, remove the oldest ones
		assistantMsgCount := 0
		for i := len(*h) - 1; i >= 0; i-- {
			if (*h)[i].Role == schema.Assistant {
				assistantMsgCount++
				if assistantMsgCount > 10 {
					*h = append((*h)[:i], (*h)[i+1:]...)
				}
			}
		}
	}
}

func logRequest(messages ConversationHistory) {
	msgs := make(ConversationHistory, 0, len(messages))
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
	logger := log.Info().Str("role", string(resp.Role)).
		Str("content", resp.Content)
	if resp.ResponseMeta != nil {
		logger = logger.Interface("response_meta", resp.ResponseMeta)
	}
	if resp.Extra != nil {
		logger = logger.Interface("extra", resp.Extra)
	}
	logger.Msg("log response message")
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
