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

	// Test new bracket format
	text = "Thought: 我需要点击这个按钮\nAction: click(start_box='[100, 200, 150, 250]')"
	result, err = parser.Parse(text, types.Size{Height: 1000, Width: 1000})
	assert.Nil(t, err)
	assert.Equal(t, result.Actions[0].ActionType, "click")
	assert.Contains(t, result.Actions[0].ActionInputs, "start_box")
	coords := result.Actions[0].ActionInputs["start_box"].([]float64)
	assert.Equal(t, 4, len(coords))
	assert.Equal(t, 100.0, coords[0])
	assert.Equal(t, 200.0, coords[1])
	assert.Equal(t, 150.0, coords[2])
	assert.Equal(t, 250.0, coords[3])

	// Test drag operation with both start_box and end_box
	text = "Thought: 我需要拖拽元素\nAction: drag(start_box='[100, 200, 150, 250]', end_box='[300, 400, 350, 450]')"
	result, err = parser.Parse(text, types.Size{Height: 1000, Width: 1000})
	assert.Nil(t, err)
	assert.Equal(t, result.Actions[0].ActionType, "drag")
	assert.Contains(t, result.Actions[0].ActionInputs, "start_box")
	assert.Contains(t, result.Actions[0].ActionInputs, "end_box")
	startCoords := result.Actions[0].ActionInputs["start_box"].([]float64)
	endCoords := result.Actions[0].ActionInputs["end_box"].([]float64)
	assert.Equal(t, 4, len(startCoords))
	assert.Equal(t, 4, len(endCoords))
	assert.Equal(t, 100.0, startCoords[0])
	assert.Equal(t, 300.0, endCoords[0])
}
