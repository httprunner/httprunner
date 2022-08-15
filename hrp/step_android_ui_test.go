package hrp

import (
	"fmt"
	"testing"
)

func TestAndroidAction(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("android ui action"),
		TestSteps: []IStep{
			NewStep("launch douyin").
				Android().Serial("xxx").Click("抖音").
				Validate().
				AssertXpathExists("首页", "首页 tab 不存在").
				AssertXpathExists("消息", "消息 tab 不存在"),
			NewStep("swipe up and down").
				Android().Serial("xxx").SwipeUp().SwipeUp().SwipeDown(),
		},
	}
	tCase := testCase.ToTCase()
	fmt.Println(tCase)

	// err := NewRunner(t).Run(testCase)
	// if err != nil {
	// 	t.Fatal(err)
	// }
}
