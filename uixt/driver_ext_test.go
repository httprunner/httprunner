//go:build localtest

package uixt

import (
	"bytes"
	"image"
	"os"
	"testing"
	"time"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriverExt_NewMethod1(t *testing.T) {
	device, err := NewAndroidDevice(option.WithUIA2(true))
	require.Nil(t, err)
	driver, err := device.NewDriver()
	require.Nil(t, err)
	driverExt, err := NewXTDriver(driver,
		option.WithCVService(option.CVServiceTypeVEDEM))
	require.Nil(t, err)
	driverExt.TapByOCR("推荐")
}

func TestDriverExt_NewMethod2(t *testing.T) {
	device, err := NewAndroidDevice()
	require.Nil(t, err)
	driver, err := NewUIA2Driver(device)
	require.Nil(t, err)
	driverExt, err := NewXTDriver(driver,
		option.WithCVService(option.CVServiceTypeVEDEM))
	require.Nil(t, err)
	driverExt.TapByOCR("推荐")
}

func TestDriverExt(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, _ := NewADBDriver(device)
	driverExt, err := NewXTDriver(driver,
		option.WithCVService(option.CVServiceTypeVEDEM))
	require.Nil(t, err)

	// call IDriver methods
	driverExt.TapXY(0.2, 0.5)
	driverExt.Swipe(0.2, 0.5, 0.8, 0.5)
	driverExt.AppLaunch("com.ss.android.ugc.aweme")

	// call AI extended methods
	driverExt.TapByOCR("推荐")
	texts, _ := driverExt.GetScreenTexts()
	t.Log(texts)
	textRect, _ := driverExt.FindScreenText("hello")
	t.Log(textRect)

	err = driverExt.TapByCV(
		option.WithScreenShotUITypes("deepseek_send"),
		option.WithScope(0.8, 0.5, 1, 1))
	assert.Nil(t, err)

	// call IDriver methods
	driverExt.GetDevice().Install("/path/to/app")
	driverExt.GetDevice().GetPackageInfo("com.ss.android.ugc.aweme")

	// get original driver and call its methods
	adbDriver := driverExt.IDriver.(*ADBDriver)
	adbDriver.TapBySelector("hello")
	wdaDriver := driverExt.IDriver.(*WDADriver)
	wdaDriver.GetMjpegClient()
	wdaDriver.Scale()

	// get original device and call its methods
	androidDevice := driver.GetDevice().(*AndroidDevice)
	androidDevice.InstallAPK("/path/to/app.apk")
}

var driverType = "ADB"

func setupDriverExt(t *testing.T) *XTDriver {
	switch driverType {
	case "ADB":
		return setupADBDriverExt(t)
	case "UIA2":
		return setupUIA2DriverExt(t)
	case "WDA":
		return setupWDADriverExt(t)
	case "HDC":
		return setupHDCDriverExt(t)
	default:
		return setupADBDriverExt(t)
	}
}

func TestDriverExt_FindScreenText(t *testing.T) {
	driver := setupDriverExt(t)
	point, err := driver.FindScreenText("首页")
	assert.Nil(t, err)
	t.Log(point)
}

func TestDriverExt_Seek(t *testing.T) {
	driver := setupDriverExt(t)

	textRect, err := driver.FindScreenText("首页")
	assert.Nil(t, err)

	size, err := driver.WindowSize()
	assert.Nil(t, err)
	width := size.Width

	point := textRect.Center()
	y := point.Y - 40
	for i := 0; i < 5; i++ {
		err := driver.Swipe(0.5, 0.8, 0.5, 0.2)
		assert.Nil(t, err)
		time.Sleep(1 * time.Second)
		err = driver.Swipe(20, y, float64(width)*0.6, y)
		assert.Nil(t, err)
		time.Sleep(1 * time.Second)
	}
}

func TestDriverExt_TapByOCR(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.TapByOCR("天气", option.WithScope(0, 0.7, 0.3, 1))
	assert.Nil(t, err)
}

func TestDriverExt_TapByLLM(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.AIAction("点击第一个帖子的作者头像")
	assert.Nil(t, err)

	err = driver.AIAssert("当前在个人介绍页")
	assert.Nil(t, err)
}

func TestDriverExt_StartToGoal(t *testing.T) {
	driver := setupDriverExt(t)

	userInstruction := `连连看是一款经典的益智消除类小游戏，通常以图案或图标为主要元素。以下是连连看的基本规则说明：
	1. 游戏目标: 玩家需要在规定时间内，通过连接相同的图案或图标，将它们从游戏界面中消除。
	2. 连接规则:
	- 两个相同的图案可以通过不超过三条直线连接。
	- 连接线可以水平或垂直，但不能斜线，也不能跨过其他图案。
	- 连接线的转折次数不能超过两次。
	3. 游戏界面:
	- 游戏界面通常是一个矩形区域，内含多个图案或图标，排列成行和列。
	- 图案或图标在未选中状态下背景为白色，选中状态下背景为绿色。
	4. 时间限制: 游戏通常设有时间限制，玩家需要在时间耗尽前完成所有图案的消除。
	5. 得分机制: 每成功连接并消除一对图案，玩家会获得相应的分数。完成游戏后，根据剩余时间和消除效率计算总分。
	6. 关卡设计: 游戏可能包含多个关卡，随着关卡的推进，图案的复杂度和数量会增加。

	注意事项：
	1、当连接错误时，顶部的红心会减少一个，需及时调整策略，避免红心变为0个后游戏失败
	2、不要连续 2 次点击同一个图案
	3、不要犯重复的错误
	`

	userInstruction += "\n\n请严格按照以上游戏规则，开始游戏；注意，请只做点击操作"

	err := driver.StartToGoal(userInstruction)
	assert.Nil(t, err)
}

func TestDriverExt_PlanNextAction(t *testing.T) {
	driver := setupDriverExt(t)
	result, err := driver.PlanNextAction("启动抖音")
	assert.Nil(t, err)
	t.Log(result)
}

func TestDriverExt_prepareSwipeAction(t *testing.T) {
	driver := setupDriverExt(t)

	swipeAction := prepareSwipeAction(driver, "up", option.WithDirection("down"))
	err := swipeAction(driver)
	assert.Nil(t, err)

	swipeAction = prepareSwipeAction(driver, "up", option.WithCustomDirection(0.5, 0.5, 0.5, 0.9))
	err = swipeAction(driver)
	assert.Nil(t, err)
}

func TestDriverExt_SwipeToTapApp(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.SwipeToTapApp("抖音")
	assert.Nil(t, err)
}

func TestDriverExt_SwipeToTapTexts(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.AppLaunch("com.ss.android.ugc.aweme")
	assert.Nil(t, err)

	err = driver.SwipeToTapTexts(
		[]string{"点击进入直播间", "直播中"},
		option.WithDirection("up"),
		option.WithMaxRetryTimes(10))
	assert.Nil(t, err)
}

func TestDriverExt_CheckPopup(t *testing.T) {
	driver := setupADBDriverExt(t)
	popup, err := driver.CheckPopup()
	require.Nil(t, err)
	if popup == nil {
		t.Log("no popup found")
	} else {
		t.Logf("found popup: %v", popup)
	}
}

func TestDriverExt_ClosePopupsHandler(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.ClosePopupsHandler()
	assert.Nil(t, err)
}

func TestDriverExt_Action_Offset(t *testing.T) {
	driver := setupADBDriverExt(t)

	// tap point with constant offset
	err := driver.TapXY(0.5, 0.5, option.WithTapOffset(-10, 10))
	assert.Nil(t, err)

	// tap point with random offset
	err = driver.TapXY(0.5, 0.5, option.WithOffsetRandomRange(-10, 10))
	assert.Nil(t, err)

	// swipe direction with constant offset
	err = driver.Swipe(0.5, 0.5, 0.5, 0.9,
		option.WithSwipeOffset(-50, 50, -50, 50))
	assert.Nil(t, err)

	// swipe direction with random offset
	err = driver.Swipe(0.5, 0.5, 0.5, 0.9,
		option.WithOffsetRandomRange(-50, 50))
	assert.Nil(t, err)

	// drag direction with random offset
	err = driver.Drag(0.5, 0.5, 0.5, 0.9,
		option.WithOffsetRandomRange(-50, 50))
	assert.Nil(t, err)

	// tap random point in ocr text rect
	err = driver.TapByOCR("首页", option.WithTapRandomRect(true))
	assert.Nil(t, err)

	err = driver.TapByCV(
		option.WithScreenShotUITypes("deepseek_send"),
		option.WithTapRandomRect(true))
	assert.Nil(t, err)
}

func TestSaveImageWithCircle(t *testing.T) {
	imgBytes, err := os.ReadFile("ai/testdata/llk_1.png")
	require.NoError(t, err)
	imgBuf := bytes.NewBuffer(imgBytes)

	point := image.Point{X: 500, Y: 500}
	outputPath := "ai/testdata/output.png"

	err = SaveImageWithCircleMarker(imgBuf, point, outputPath)
	require.NoError(t, err)

	defer os.Remove(outputPath)
}

func TestSaveImageWithArrow(t *testing.T) {
	imgBytes, err := os.ReadFile("ai/testdata/llk_1.png")
	require.NoError(t, err)
	imgBuf := bytes.NewBuffer(imgBytes)

	from := image.Point{X: 500, Y: 500}
	to := image.Point{X: 1000, Y: 1000}
	outputPath := "ai/testdata/output.png"

	err = SaveImageWithArrowMarker(imgBuf, from, to, outputPath)
	require.NoError(t, err)

	defer os.Remove(outputPath)
}

func TestMarkOperation(t *testing.T) {
	driver := setupDriverExt(t)

	opts := []option.ActionOption{option.WithMarkOperationEnabled(true)}

	// tap point
	err := driver.TapXY(0.5, 0.5, opts...)
	assert.Nil(t, err)

	err = driver.TapAbsXY(500, 800, opts...)
	assert.Nil(t, err)

	// swipe
	err = driver.Swipe(0.2, 0.5, 0.8, 0.5, opts...)
	assert.Nil(t, err)
	err = driver.Swipe(0.3, 0.7, 0.3, 0.3, opts...)
	assert.Nil(t, err)
}
