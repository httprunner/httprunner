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

func TestIOSAppLaunch(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("启动 & 关闭 App"),
		TestSteps: []IStep{
			NewStep("终止今日头条").
				IOS().AppTerminate("com.ss.iphone.article.News"),
			NewStep("启动今日头条").
				IOS().AppLaunch("com.ss.iphone.article.News"),
			NewStep("终止今日头条").
				IOS().AppTerminate("com.ss.iphone.article.News"),
			NewStep("启动今日头条").
				IOS().AppLaunchUnattached("com.ss.iphone.article.News"),
		},
	}

	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSWeixinLive(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios ui action on 微信直播"),
		TestSteps: []IStep{
			NewStep("启动微信").
				IOS().
				Home().
				AppTerminate("com.tencent.xin"). // 关闭已运行的微信，确保启动微信后在「微信」首页
				Click("微信").
				Validate().
				AssertNameExists("通讯录", "微信启动失败，「通讯录」不存在"),
			NewStep("进入直播页").
				IOS().
				Click("发现").Sleep(5).       // 进入「发现页」；等待 5 秒确保加载完成
				Click([]float64{0.5, 0.3}). // 基于坐标位置点击「直播」；TODO：通过 OCR 识别「直播」
				Validate().
				AssertNameExists("直播", "「直播」不存在"),
			NewStep("向上滑动 5 次").
				IOS().
				SwipeUp().Times(3).ScreenShot(). // 上划 3 次，截图保存
				SwipeUp().Times(2).ScreenShot(), // 再上划 2 次，截图保存
		},
	}

	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSCameraPhotoCapture(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios camera photo capture"),
		TestSteps: []IStep{
			NewStep("launch camera").
				IOS().Home().
				StopCamera().
				StartCamera().
				Validate().
				AssertNameExists("PhotoCapture", "拍照按钮不存在"),
			NewStep("start recording").
				IOS().Click("PhotoCapture"),
		},
	}

	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSCameraVideoCapture(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios camera video capture"),
		TestSteps: []IStep{
			NewStep("launch camera").
				IOS().Home().
				StopCamera().
				StartCamera().
				Validate().
				AssertNameExists("PhotoCapture", "录像按钮不存在"),
			NewStep("switch to video capture").
				IOS().
				SwipeRight().
				Validate().
				AssertNameExists("VideoCapture", "拍摄按钮不存在"),
			NewStep("start recording").
				IOS().
				Click("VideoCapture"). // 开始录像
				Sleep(5).
				Click("VideoCapture"), // 停止录像
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
