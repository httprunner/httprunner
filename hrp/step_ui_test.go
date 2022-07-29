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

	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSSettingsAction(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios ui action on Settings"),
		TestSteps: []IStep{
			NewStep("launch Settings").
				IOS().Home().Click("//*[@label='设置']").
				Validate().
				AssertNameExists("飞行模式", "「飞行模式」不存在").
				AssertNameNotExists("飞行模式2", "「飞行模式2」不存在"),
			NewStep("swipe up and down").
				IOS().SwipeUp().SwipeUp().SwipeDown(),
		},
	}

	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSSearchApp(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios ui action on Search App 资源库"),
		TestSteps: []IStep{
			NewStep("进入 App 资源库 搜索框").
				IOS().Home().SwipeLeft().Times(2).Click("dewey-search-field").
				Validate().
				AssertNameExists("取消", "「取消」不存在"),
			NewStep("搜索抖音").
				IOS().Input("抖音\n"),
		},
	}

	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSWeixin(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios ui action on 微信"),
		TestSteps: []IStep{
			NewStep("启动微信").
				IOS().Home().Click("微信").
				Validate().
				AssertNameExists("通讯录", "微信启动失败，「通讯录」不存在"),
			NewStep("进入直播页").
				IOS().Click("发现").Click([]float64{0.5, 0.3}).
				Validate().
				AssertNameExists("直播", "「直播」不存在"),
			NewStep("向上滑动 5 次").
				IOS().SwipeUp().Times(5),
		},
	}

	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSDouyinAction(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios ui action on 抖音"),
		TestSteps: []IStep{
			NewStep("launch douyin").
				IOS().Home().Click("//*[@label='抖音']").
				Validate().
				AssertNameExists("首页", "首页 tab 不存在").
				AssertNameExists("消息", "消息 tab 不存在"),
			NewStep("swipe up and down").
				IOS().SwipeUp().Times(3).SwipeDown(),
		},
	}

	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
