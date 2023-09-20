package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestAndroidExpertTest(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("安卓专家用例").
			WithVariables(map[string]interface{}{
				"device":    "${ENV(SerialNumber)}",
				"query":     "${ENV(query)}",
				"bundle_id": "com.ss.android.ugc.aweme",
				"app_name":  "抖音",
			}).
			SetAndroid(
				uixt.WithSerialNumber("$device"),
				uixt.WithAdbLogOn(true),
				uixt.WithUIA2(true),
			),
		TestSteps: []hrp.IStep{
			// 温启动
			hrp.NewStep("app_launch 以及 ui_foreground_app equal 断言").
				Android().
				AppLaunch("$bundle_id").
				Sleep(2).
				Validate().
				AssertAppInForeground("$bundle_id"),
			hrp.NewStep("home 以及 swipe_to_tap_app 默认配置").
				Android().
				Home().
				SwipeToTapApp("$app_name").
				Sleep(10),
			hrp.NewStep("处理弹窗 close_popups 默认配置 以及 ui_ocr exists 断言").
				Android().
				ClosePopups().
				Validate().
				AssertOCRExists("推荐", "进入抖音失败"),
			// 直播赛道
			hrp.NewStep("【直播】feed头像或卡片进房 swipe_to_tap_texts 自定义配置").
				Android().
				SwipeToTapTexts(
					[]string{"直播", "直播中", "点击进入直播间"},
					uixt.WithCustomDirection(0.5, 0.7, 0.5, 0.3),
					uixt.WithScope(0.2, 0.2, 1, 0.8),
					uixt.WithMaxRetryTimes(50),
					uixt.WithWaitTime(1.5),
					uixt.WithIdentifier("click_live"),
				),
			hrp.NewStep("sleep 10s").
				Android().
				Sleep(10),
			hrp.NewStep("【直播】swipe 自定义配置 以及 back").
				Android().
				Swipe(
					0.5, 0.7, 0.5, 0.3,
					uixt.WithIdentifier("slide_in_live"),
				).
				Sleep(5).
				Back().
				Sleep(5),
			// 搜索赛道
			hrp.NewStep("【搜索】点击放大镜 tap_xy 自定义配置").
				Android().
				TapXY(
					0.9, 0.08,
					uixt.WithIdentifier("click_search_in_middle_page"),
				).
				Sleep(5),
			hrp.NewStep("【搜索】输入query词 input").
				Android().
				Input(
					"$query",
					uixt.WithIdentifier("input_query"),
				).
				Sleep(5),
			hrp.NewStep("【搜索】点击搜索按钮 tap_ocr 自定义配置").
				Android().
				TapByOCR(
					"搜索",
					uixt.WithIdentifier("click_search_after_input_query"),
					uixt.WithIndex(0),
				).
				Sleep(5),
			hrp.NewStep("选择直播页签 tap_ocr 默认配置").
				Android().
				TapByOCR("直播").
				Sleep(5),
			// 生活服务赛道
			hrp.NewStep("【生活服务】进入直播间 tap_xy").
				Android().
				TapXY(0.5, 0.5).
				Sleep(5),
			hrp.NewStep("【生活服务】点击货架商品 tap_ocr 自定义配置").
				Android().
				TapByUITypes(
					uixt.WithScreenShotUITypes("dyhouse", "shoppingbag"),
					uixt.WithIdentifier("click_sales_rack"),
				).
				Sleep(5),
			// 冷启动
			hrp.NewStep("app_terminate 以及 ui_foreground_app not_equal 断言").
				Android().
				AppTerminate("$bundle_id").
				Sleep(2).
				Validate().
				AssertAppNotInForeground("$bundle_id"),
			hrp.NewStep("home 以及 swipe_to_tap_app 自定义配置").
				Android().
				Home().
				SwipeToTapApp("$app_name", uixt.WithMaxRetryTimes(5), uixt.WithInterval(1), uixt.WithTapOffset(0, -50)).
				Sleep(10),
			hrp.NewStep("处理弹窗 close_popups 自定义配置 以及 ui_ocr exists 断言").
				Android().
				ClosePopups(
					uixt.WithMaxRetryTimes(3),
					uixt.WithInterval(2),
				).
				Validate().
				AssertOCRExists("推荐", "进入抖音失败"),
			// 点播赛道
			hrp.NewStep("【点播】滑动消费").
				Android().
				VideoCrawler(map[string]interface{}{
					"timeout": 600,
					"feed": map[string]interface{}{
						"target_count": 10,
						"target_labels": []map[string]interface{}{
							{"text": "^广告$", "scope": []float64{0, 0.5, 1, 1}, "regex": true},
							{"text": "^图文$", "scope": []float64{0, 0.5, 1, 1}, "regex": true},
							{"text": `^特效\|`, "scope": []float64{0, 0.5, 1, 1}, "regex": true},
							{"text": `^模板\|`, "scope": []float64{0, 0.5, 1, 1}, "regex": true},
							{"text": `^购物\|`, "scope": []float64{0, 0.5, 1, 1}, "regex": true},
						},
						"sleep_random": []float64{0, 5, 0.6, 5, 15, 0.2, 15, 50, 0.2},
					},
					"live": map[string]interface{}{
						"target_count": 0,
						"sleep_random": []float64{20, 20},
					},
				}),
			// localtime 时间戳界面
			hrp.NewStep("返回主界面，并打开本地时间戳").
				Android().
				Home().
				AppTerminate("$bundle_id").
				Sleep(3).
				SwipeToTapApp("local", uixt.WithMaxRetryTimes(5)).Sleep(10),
			hrp.NewStep("screeshot 以及 sleep_random").
				Loop(3).
				Android().
				ScreenShot().
				SleepRandom(1, 3),
		},
	}

	if err := testCase.Dump2JSON("android_expert_test.json"); err != nil {
		t.Fatal(err)
	}
}

func TestIOSExpertTest(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("iOS 专家用例").
			WithVariables(map[string]interface{}{
				"device":    "${ENV(UDID)}",
				"query":     "${ENV(query)}",
				"bundle_id": "com.ss.iphone.ugc.Aweme",
				"app_name":  "抖音",
			}).
			SetIOS(
				uixt.WithUDID("$device"),
				uixt.WithWDALogOn(true),
				uixt.WithWDAPort(8700),
				uixt.WithWDAMjpegPort(8800),
			),
		TestSteps: []hrp.IStep{
			// 温启动
			// iOS 不支持前台 App 断言操作
			hrp.NewStep("启动应用程序 app_launch").
				IOS().
				AppLaunch("$bundle_id").
				Sleep(2),
			hrp.NewStep("home 以及 swipe_to_tap_app 默认配置").
				IOS().
				Home().
				SwipeToTapApp("$app_name").
				Sleep(10),
			hrp.NewStep("处理弹窗 close_popups 默认配置 以及 ui_ocr exists 断言").
				IOS().
				ClosePopups().
				Validate().
				AssertOCRExists("推荐", "进入抖音失败"),
			// 直播赛道
			hrp.NewStep("【直播】feed头像或卡片进房 swipe_to_tap_texts 自定义配置").
				IOS().
				SwipeToTapTexts(
					[]string{"直播", "直播中", "点击进入直播间"},
					uixt.WithCustomDirection(0.5, 0.7, 0.5, 0.3),
					uixt.WithScope(0.2, 0.2, 1, 0.8),
					uixt.WithMaxRetryTimes(50),
					uixt.WithWaitTime(1.5),
					uixt.WithIdentifier("click_live"),
				),
			hrp.NewStep("sleep 10s").
				IOS().
				Sleep(10),
			hrp.NewStep("【直播】swipe 自定义配置 以及 back").
				IOS().
				Swipe(
					0.5, 0.7, 0.5, 0.3,
					uixt.WithIdentifier("slide_in_live"),
				).
				Sleep(5).
				Back().
				Sleep(5),
			// 搜索赛道
			hrp.NewStep("【搜索】点击放大镜 tap_xy 自定义配置").
				IOS().
				TapXY(
					0.9, 0.075,
					uixt.WithIdentifier("click_search_in_middle_page"),
				).
				Sleep(5),
			hrp.NewStep("【搜索】输入query词 input").
				IOS().
				Input(
					"$query",
					uixt.WithIdentifier("input_query"),
				).
				Sleep(5),
			hrp.NewStep("【搜索】点击搜索按钮 tap_ocr 自定义配置").
				IOS().
				TapByOCR(
					"搜索",
					uixt.WithIdentifier("click_search_after_input_query"),
					uixt.WithIndex(0),
				).
				Sleep(5),
			// 生活服务赛道
			hrp.NewStep("选择直播页签 tap_ocr 默认配置").
				IOS().
				TapByOCR("直播").
				Sleep(5),
			hrp.NewStep("【生活服务】进入直播间 tap_xy").
				IOS().
				TapXY(0.5, 0.5).
				Sleep(5),
			hrp.NewStep("【生活服务】点击货架商品 tap_ocr 自定义配置").
				IOS().
				TapByUITypes(
					uixt.WithScreenShotUITypes("dyhouse", "shoppingbag"),
					uixt.WithIdentifier("click_sales_rack"),
				).
				Sleep(5),
			// 冷启动
			// iOS 不支持前台 App 断言操作
			hrp.NewStep("终止应用程序 app_terminate").
				IOS().
				AppTerminate("$bundle_id").
				Sleep(2),
			hrp.NewStep("home 以及 swipe_to_tap_app 自定义配置").
				IOS().
				Home().
				SwipeToTapApp("$app_name", uixt.WithMaxRetryTimes(5), uixt.WithInterval(1), uixt.WithTapOffset(0, -50)).
				Sleep(10),
			hrp.NewStep("处理弹窗 close_popups 自定义配置 以及 ui_ocr exists 断言").
				IOS().
				ClosePopups(
					uixt.WithMaxRetryTimes(3),
					uixt.WithInterval(2),
				).
				Validate().
				AssertOCRExists("推荐", "进入抖音失败"),
			// 点播赛道
			hrp.NewStep("【点播】滑动消费").
				IOS().
				VideoCrawler(map[string]interface{}{
					"timeout": 600,
					"feed": map[string]interface{}{
						"target_count": 10,
						"target_labels": []map[string]interface{}{
							{"text": "^广告$", "scope": []float64{0, 0.5, 1, 1}, "regex": true},
							{"text": "^图文$", "scope": []float64{0, 0.5, 1, 1}, "regex": true},
							{"text": `^特效\|`, "scope": []float64{0, 0.5, 1, 1}, "regex": true},
							{"text": `^模板\|`, "scope": []float64{0, 0.5, 1, 1}, "regex": true},
							{"text": `^购物\|`, "scope": []float64{0, 0.5, 1, 1}, "regex": true},
						},
						"sleep_random": []float64{0, 5, 0.6, 5, 15, 0.2, 15, 50, 0.2},
					},
					"live": map[string]interface{}{
						"target_count": 0,
						"sleep_random": []float64{20, 20},
					},
				}),
			// localtime 时间戳界面
			hrp.NewStep("返回主界面，并打开本地时间戳").
				IOS().
				Home().
				AppTerminate("$bundle_id").
				Sleep(3).
				SwipeToTapApp("local", uixt.WithMaxRetryTimes(5)).Sleep(10),
			hrp.NewStep("screeshot 以及 sleep_random").
				Loop(3).
				IOS().
				ScreenShot().
				SleepRandom(1, 3),
		},
	}

	if err := testCase.Dump2JSON("ios_expert_test.json"); err != nil {
		t.Fatal(err)
	}
}
