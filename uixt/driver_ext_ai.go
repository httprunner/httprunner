package uixt

import (
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/pkg/errors"
)

func (dExt *XTDriver) PlanNextAction(text string, opts ...option.ActionOption) (*ai.PlanningResult, error) {
	if dExt.LLMService == nil {
		return nil, errors.New("LLM service is not initialized")
	}

	screenShotBase64, err := dExt.GetScreenShotBase64()
	if err != nil {
		return nil, err
	}

	planningOpts := &ai.PlanningOptions{
		UserInstruction: text,
		ConversationHistory: []*schema.Message{
			{
				Role: schema.User,
				MultiContent: []schema.ChatMessagePart{
					{
						Type: "image_url",
						ImageURL: &schema.ChatMessageImageURL{
							URL: screenShotBase64,
						},
					},
				},
			},
		},
	}

	result, err := dExt.LLMService.Call(planningOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next action from planner")
	}
	return result, nil
}
