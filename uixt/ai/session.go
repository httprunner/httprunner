package ai

import (
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
