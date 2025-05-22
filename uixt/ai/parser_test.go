package ai

import (
	"testing"

	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/stretchr/testify/assert"
)

func TestParseAction(t *testing.T) {
	actionStr := "click(point='<point>200 300</point>')"
	result, err := ParseAction(actionStr)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, result.Function, "click")
	assert.Equal(t, result.Args["point"], "<point>200 300</point>")
}

func TestParseActionToStructureOutput(t *testing.T) {
	text := "Thought: test\nAction: click(point='<point>200 300</point>')"
	parser := &UITARSContentParser{}
	result, err := parser.Parse(text, types.Size{Height: 224, Width: 224})
	assert.Nil(t, err)
	assert.Equal(t, result.Actions[0].ActionType, "click")
	assert.Contains(t, result.Actions[0].ActionInputs, "start_box")

	text = "Thought: 我看到页面上有几个帖子，第二个帖子的标题是\"字节四年，头发白了\"。要完成任务，我需要点击这个帖子下方的作者头像，这样就能进入作者的个人主页了。\nAction: click(start_point='<point>550 450 550 450</point>')"
	result, err = parser.Parse(text, types.Size{Height: 2341, Width: 1024})
	assert.Nil(t, err)
	assert.Equal(t, result.Actions[0].ActionType, "click")
	assert.Contains(t, result.Actions[0].ActionInputs, "start_box")
}
