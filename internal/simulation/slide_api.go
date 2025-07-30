package simulation

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/httprunner/httprunner/v5/uixt/types"
)

// SlideRequest 滑动请求参数
type SlideRequest struct {
	StartX    float64   `json:"start_x"`   // 起始X坐标
	StartY    float64   `json:"start_y"`   // 起始Y坐标
	Direction Direction `json:"direction"` // 滑动方向
	Distance  float64   `json:"distance"`  // 滑动距离
	DeviceID  int       `json:"device_id"` // 设备ID
	Pressure  float64   `json:"pressure"`  // 压力值
	Size      float64   `json:"size"`      // 按压大小(接触面积)
}

// PointToPointSlideRequest 点对点滑动请求参数
type PointToPointSlideRequest struct {
	StartX   float64 `json:"start_x"`   // 起始X坐标
	StartY   float64 `json:"start_y"`   // 起始Y坐标
	EndX     float64 `json:"end_x"`     // 结束X坐标
	EndY     float64 `json:"end_y"`     // 结束Y坐标
	DeviceID int     `json:"device_id"` // 设备ID
	Pressure float64 `json:"pressure"`  // 压力值
	Size     float64 `json:"size"`      // 按压大小(接触面积)
}

// SlideResponse 滑动响应结果
type SlideResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Points  []SlidePoint `json:"points"`
	Metrics SlideMetrics `json:"metrics"`
}

// SlideMetrics 滑动指标
type SlideMetrics struct {
	TotalDuration   int64   `json:"total_duration_ms"`   // 总持续时间(毫秒)
	PointCount      int     `json:"point_count"`         // 轨迹点数量
	ActualDistance  float64 `json:"actual_distance"`     // 实际滑动距离
	AverageInterval float64 `json:"average_interval_ms"` // 平均采样间隔
}

// SlidePoint 滑动轨迹点
type SlidePoint struct {
	Timestamp int64   `json:"timestamp"`  // 时间戳(毫秒)
	X         float64 `json:"x"`          // X坐标
	Y         float64 `json:"y"`          // Y坐标
	DeviceID  int     `json:"device_id"`  // 设备ID
	Pressure  float64 `json:"pressure"`   // 压力值
	Size      float64 `json:"size"`       // 按压大小(接触面积)
	EventTime int64   `json:"event_time"` // 相对第一个点的时间(ms)，第一个点为0
}

// Direction 滑动方向枚举
type Direction string

const (
	Up    Direction = "up"
	Down  Direction = "down"
	Left  Direction = "left"
	Right Direction = "right"
)

// SlideConfig 滑动配置参数
type SlideConfig struct {
	MinDuration    int64   // 最小持续时间(毫秒)
	MaxDuration    int64   // 最大持续时间(毫秒)
	MinPoints      int     // 最小点数
	MaxPoints      int     // 最大点数
	CurveIntensity float64 // 曲线强度(0-1)
	NoiseLevel     float64 // 噪声级别
}

// DefaultSlideConfig 默认配置
var DefaultSlideConfig = SlideConfig{
	MinDuration:    80,
	MaxDuration:    200,
	MinPoints:      4,
	MaxPoints:      8,
	CurveIntensity: 0.05,
	NoiseLevel:     2.0,
}

// SlideSimulatorAPI 滑动仿真API
type SlideSimulatorAPI struct {
	rand   *rand.Rand
	config SlideConfig
}

// TestCase 测试用例
type TestCase struct {
	Name      string
	StartX    float64
	StartY    float64
	Direction Direction
	Distance  float64
	DeviceID  int
	Pressure  float64
	Size      float64
}

// NewSlideSimulatorAPI 创建新的滑动仿真API
func NewSlideSimulatorAPI(config *SlideConfig) *SlideSimulatorAPI {
	if config == nil {
		config = &DefaultSlideConfig
	}

	return &SlideSimulatorAPI{
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		config: *config,
	}
}

// GenerateSlide 生成滑动轨迹
func (api *SlideSimulatorAPI) GenerateSlide(req SlideRequest) SlideResponse {
	// 验证输入参数
	if err := api.validateRequest(req); err != nil {
		return SlideResponse{
			Success: false,
			Message: err.Error(),
		}
	}

	// 生成滑动轨迹
	points := api.generateSlidePoints(req)

	// 计算指标
	metrics := api.calculateMetrics(points)

	return SlideResponse{
		Success: true,
		Points:  points,
		Metrics: metrics,
	}
}

// pressureRiseCurve 压力上升曲线，模拟真实的压力变化
func (api *SlideSimulatorAPI) pressureRiseCurve(t float64) float64 {
	// 使用二次函数模拟压力逐渐增加的过程
	return t*t*0.6 + t*0.4
}

// pressureFallCurve 压力下降曲线，模拟真实的压力变化
func (api *SlideSimulatorAPI) pressureFallCurve(t float64) float64 {
	// 使用指数衰减模拟压力快速下降的过程
	return 1.0 - (1.0-math.Exp(-t*2.0))*0.8
}

// GeneratePointToPointSlide 生成点对点滑动轨迹
func (api *SlideSimulatorAPI) GeneratePointToPointSlide(req PointToPointSlideRequest) SlideResponse {
	// 验证输入参数
	if err := api.validatePointToPointRequest(req); err != nil {
		return SlideResponse{
			Success: false,
			Message: err.Error(),
		}
	}

	// 生成滑动轨迹
	points := api.generatePointToPointSlidePoints(req)

	// 计算指标
	metrics := api.calculateMetrics(points)

	return SlideResponse{
		Success: true,
		Points:  points,
		Metrics: metrics,
	}
}

// validateRequest 验证请求参数
func (api *SlideSimulatorAPI) validateRequest(req SlideRequest) error {
	if req.Distance <= 0 {
		return fmt.Errorf("distance must be positive")
	}

	switch req.Direction {
	case Up, Down, Left, Right:
		// 有效方向
	default:
		return fmt.Errorf("invalid direction: %s", req.Direction)
	}

	return nil
}

// validatePointToPointRequest 验证点对点请求参数
func (api *SlideSimulatorAPI) validatePointToPointRequest(req PointToPointSlideRequest) error {
	// 检查起始点和结束点是否相同
	if req.StartX == req.EndX && req.StartY == req.EndY {
		return fmt.Errorf("start point and end point cannot be the same")
	}

	// 检查距离是否合理
	distance := math.Sqrt((req.EndX-req.StartX)*(req.EndX-req.StartX) + (req.EndY-req.StartY)*(req.EndY-req.StartY))
	if distance < 10 {
		return fmt.Errorf("distance too short: %.2f pixels", distance)
	}

	return nil
}

// generateSlidePoints 生成滑动轨迹点
func (api *SlideSimulatorAPI) generateSlidePoints(req SlideRequest) []SlidePoint {
	// 计算终点坐标
	endX, endY := api.calculateEndPoint(req.StartX, req.StartY, req.Direction, req.Distance)

	// 计算滑动参数
	duration := api.calculateDuration(req.Distance)
	pointCount := api.calculatePointCount(duration)

	// 生成时间戳序列
	timestamps := api.generateTimestamps(duration, pointCount)

	// 生成轨迹点
	points := make([]SlidePoint, pointCount)

	// 计算总偏移趋势（基于真实数据分析）
	var totalOffsetX, totalOffsetY float64
	switch req.Direction {
	case Up:
		// 上滑时倾向于向右偏移，偏移量为距离的15%-35%
		offsetRatio := 0.15 + api.rand.Float64()*0.20
		totalOffsetX = req.Distance * offsetRatio
		totalOffsetY = 0
	case Down:
		// 下滑时可以左右偏移，但偏移较小
		offsetRatio := 0.10 + api.rand.Float64()*0.15
		totalOffsetX = (api.rand.Float64() - 0.5) * req.Distance * offsetRatio
		totalOffsetY = 0
	case Left:
		// 左滑时可能向上或向下偏移
		offsetRatio := 0.05 + api.rand.Float64()*0.20
		totalOffsetX = 0
		totalOffsetY = (api.rand.Float64() - 0.5) * req.Distance * offsetRatio
	case Right:
		// 右滑时偏移相对较小
		offsetRatio := 0.03 + api.rand.Float64()*0.10
		totalOffsetX = 0
		totalOffsetY = (api.rand.Float64() - 0.5) * req.Distance * offsetRatio
	}

	// 生成size变化曲线（基于真实数据分析）
	sizeValues := api.generateSizeValues(pointCount, req.Size)

	// 生成pressure变化曲线（基于真实数据分析）
	pressureValues := api.generatePressureValues(pointCount, req.Pressure, req.Direction)

	baseTimestamp := timestamps[0]
	for i := 0; i < pointCount; i++ {
		progress := float64(i) / float64(pointCount-1)

		// 使用贝塞尔曲线生成基础轨迹
		x, y := api.calculateBezierPoint(req.StartX, req.StartY, endX, endY, progress, req.Direction)

		// 添加渐进式偏移（模拟真实滑动的累积偏移）
		progressiveOffsetX := totalOffsetX * api.getProgressiveOffset(progress)
		progressiveOffsetY := totalOffsetY * api.getProgressiveOffset(progress)

		x += progressiveOffsetX
		y += progressiveOffsetY

		// 添加随机噪声（减小噪声强度，因为主要偏移已经通过渐进式偏移实现）
		x += api.addNoise(api.config.NoiseLevel * 0.5)
		y += api.addNoise(api.config.NoiseLevel * 0.5)

		eventTime := timestamps[i] - baseTimestamp

		points[i] = SlidePoint{
			Timestamp: timestamps[i],
			X:         x,
			Y:         y,
			DeviceID:  req.DeviceID,
			Pressure:  pressureValues[i],
			Size:      sizeValues[i],
			EventTime: eventTime,
		}
	}

	return points
}

// generatePointToPointSlidePoints 生成点对点滑动轨迹点
func (api *SlideSimulatorAPI) generatePointToPointSlidePoints(req PointToPointSlideRequest) []SlidePoint {
	// 对起始点和结束点添加随机偏移（正负20以内）
	offsetRange := 20.0

	actualStartX := req.StartX + api.addNoise(offsetRange)
	actualStartY := req.StartY + api.addNoise(offsetRange)
	actualEndX := req.EndX + api.addNoise(offsetRange)
	actualEndY := req.EndY + api.addNoise(offsetRange)

	// 计算实际距离
	distance := math.Sqrt((actualEndX-actualStartX)*(actualEndX-actualStartX) + (actualEndY-actualStartY)*(actualEndY-actualStartY))

	// 计算滑动参数
	duration := api.calculateDuration(distance)
	pointCount := api.calculatePointCount(duration)

	// 生成时间戳序列
	timestamps := api.generateTimestamps(duration, pointCount)

	// 生成轨迹点
	points := make([]SlidePoint, pointCount)

	// 判断主要滑动方向，用于计算偏移
	dx := actualEndX - actualStartX
	dy := actualEndY - actualStartY
	var direction Direction
	if math.Abs(dy) > math.Abs(dx) {
		if dy < 0 {
			direction = Up
		} else {
			direction = Down
		}
	} else {
		if dx < 0 {
			direction = Left
		} else {
			direction = Right
		}
	}

	// 计算总偏移趋势（基于主要方向）
	var totalOffsetX, totalOffsetY float64
	switch direction {
	case Up:
		// 上滑时倾向于向右偏移
		offsetRatio := 0.10 + api.rand.Float64()*0.15
		totalOffsetX = distance * offsetRatio
		totalOffsetY = 0
	case Down:
		// 下滑时可以左右偏移，但偏移较小
		offsetRatio := 0.05 + api.rand.Float64()*0.10
		totalOffsetX = (api.rand.Float64() - 0.5) * distance * offsetRatio
		totalOffsetY = 0
	case Left:
		// 左滑时可能向上或向下偏移
		offsetRatio := 0.03 + api.rand.Float64()*0.15
		totalOffsetX = 0
		totalOffsetY = (api.rand.Float64() - 0.5) * distance * offsetRatio
	case Right:
		// 右滑时偏移相对较小
		offsetRatio := 0.02 + api.rand.Float64()*0.08
		totalOffsetX = 0
		totalOffsetY = (api.rand.Float64() - 0.5) * distance * offsetRatio
	}

	// 生成size变化曲线
	sizeValues := api.generateSizeValues(pointCount, req.Size)

	// 生成pressure变化曲线
	pressureValues := api.generatePressureValues(pointCount, req.Pressure, direction)

	baseTimestamp := timestamps[0]
	for i := 0; i < pointCount; i++ {
		progress := float64(i) / float64(pointCount-1)

		// 使用贝塞尔曲线生成基础轨迹
		x, y := api.calculateBezierPoint(actualStartX, actualStartY, actualEndX, actualEndY, progress, direction)

		// 添加渐进式偏移
		progressiveOffsetX := totalOffsetX * api.getProgressiveOffset(progress)
		progressiveOffsetY := totalOffsetY * api.getProgressiveOffset(progress)

		x += progressiveOffsetX
		y += progressiveOffsetY

		// 添加随机噪声
		x += api.addNoise(api.config.NoiseLevel * 0.5)
		y += api.addNoise(api.config.NoiseLevel * 0.5)

		eventTime := timestamps[i] - baseTimestamp

		points[i] = SlidePoint{
			Timestamp: timestamps[i],
			X:         x,
			Y:         y,
			DeviceID:  req.DeviceID,
			Pressure:  pressureValues[i],
			Size:      sizeValues[i],
			EventTime: eventTime,
		}
	}

	return points
}

// generateSizeValues 生成size值序列，基于真实数据分析
func (api *SlideSimulatorAPI) generateSizeValues(pointCount int, baseSize float64) []float64 {
	sizes := make([]float64, pointCount)

	// 如果baseSize为0，使用默认值
	if baseSize == 0 {
		baseSize = 0.04 // 默认size值，基于真实数据平均值
	}

	// 动态计算size范围，基于baseSize的值来适应不同设备
	var minSize, maxSize float64
	if baseSize < 1.0 {
		// 小数值范围（如0.04），使用原有逻辑
		minSize = 0.031
		maxSize = 0.063
		// 确保baseSize在合理范围内
		if baseSize < minSize {
			baseSize = minSize + api.rand.Float64()*(maxSize-minSize)*0.3
		}
		if baseSize > maxSize {
			baseSize = maxSize - api.rand.Float64()*(maxSize-minSize)*0.3
		}
	} else {
		// 大数值范围（如几十或几百），基于baseSize动态计算范围
		// 允许在baseSize的±20%范围内变化
		minSize = baseSize * 0.8
		maxSize = baseSize * 1.2
	}

	for i := 0; i < pointCount; i++ {
		// 基础size值随滑动进度变化
		var sizeModifier float64

		if i == 0 {
			// 开始时：可能较大或较小，有随机性
			sizeModifier = 0.8 + api.rand.Float64()*0.4 // 0.8-1.2倍
		} else if i == pointCount-1 {
			// 结束时：可能增大（手指离开前压力增加）
			if api.rand.Float64() < 0.6 { // 60%概率增大
				sizeModifier = 1.1 + api.rand.Float64()*0.3 // 1.1-1.4倍
			} else {
				sizeModifier = 0.9 + api.rand.Float64()*0.2 // 0.9-1.1倍
			}
		} else {
			// 中间过程：轻微波动
			sizeModifier = 0.85 + api.rand.Float64()*0.3 // 0.85-1.15倍
		}

		// 应用变化
		sizes[i] = baseSize * sizeModifier

		// 确保在合理范围内
		if sizes[i] < minSize {
			sizes[i] = minSize
		}
		if sizes[i] > maxSize {
			sizes[i] = maxSize
		}

		// 添加轻微随机噪声，噪声大小根据baseSize动态调整
		var noiseLevel float64
		if baseSize < 1.0 {
			noiseLevel = 0.003 // 小数值使用固定的小噪声
		} else {
			noiseLevel = baseSize * 0.01 // 大数值使用baseSize的1%作为噪声
		}
		sizes[i] += api.addNoise(noiseLevel)

		// 最终范围检查
		if sizes[i] < minSize {
			sizes[i] = minSize
		}
		if sizes[i] > maxSize {
			sizes[i] = maxSize
		}
	}

	return sizes
}

// generatePressureValues 生成pressure值序列，基于用户输入的压力值动态仿真
func (api *SlideSimulatorAPI) generatePressureValues(pointCount int, basePressure float64, direction Direction) []float64 {
	pressures := make([]float64, pointCount)

	// 如果用户没有提供压力值，使用默认值
	if basePressure <= 0 {
		basePressure = 1 // 默认压力值
	}

	// 特殊处理：当压力值为1时，保持恒定不变
	if basePressure == 1 {
		for i := 0; i < pointCount; i++ {
			pressures[i] = 1.0
		}
		return pressures
	}

	// 将整数压力值转换为浮点数
	baseP := float64(basePressure)

	// 基于真实数据观察的压力变化规律：
	// 1. 起始压力：基础压力的70%-90%
	// 2. 峰值压力：基础压力的120%-180%
	// 3. 结束压力：基础压力的30%-60%

	startPressureRatio := 0.7 + api.rand.Float64()*0.2 // 70%-90%
	peakPressureRatio := 1.2 + api.rand.Float64()*0.6  // 120%-180%
	endPressureRatio := 0.3 + api.rand.Float64()*0.3   // 30%-60%

	startPressure := baseP * startPressureRatio
	peakPressure := baseP * peakPressureRatio
	endPressure := baseP * endPressureRatio

	// 峰值出现的位置：通常在滑动过程的20%-70%处
	peakPosition := 0.2 + api.rand.Float64()*0.5
	peakIndex := int(float64(pointCount-1) * peakPosition)
	if peakIndex >= pointCount {
		peakIndex = pointCount - 1
	}

	// 确保压力值在合理范围内（0.5-15.0）
	//if startPressure < 0.5 {
	//	startPressure = 0.5
	//}
	//if peakPressure > 15.0 {
	//	peakPressure = 15.0
	//}
	//if endPressure < 0.5 {
	//	endPressure = 0.5
	//}

	for i := 0; i < pointCount; i++ {
		var pressure float64

		if i <= peakIndex {
			// 上升阶段：从起始到峰值
			if peakIndex == 0 {
				pressure = startPressure
			} else {
				t := float64(i) / float64(peakIndex)
				// 使用非线性插值，模拟真实的压力上升曲线
				t = api.pressureRiseCurve(t)
				pressure = startPressure + (peakPressure-startPressure)*t
			}
		} else {
			// 下降阶段：从峰值到结束
			t := float64(i-peakIndex) / float64(pointCount-1-peakIndex)
			// 使用非线性插值，模拟真实的压力下降曲线
			t = api.pressureFallCurve(t)
			pressure = peakPressure + (endPressure-peakPressure)*t
		}

		// 添加随机噪声（±8%），模拟真实手指压力的微小波动
		noiseRange := pressure * 0.08
		noise := (api.rand.Float64() - 0.5) * noiseRange
		pressure += noise

		// 确保pressure在合理范围内
		//if pressure < 0.5 {
		//	pressure = 0.5 + api.rand.Float64()*0.3
		//}
		//if pressure > 15.0 {
		//	pressure = 14.5 + api.rand.Float64()*0.5
		//}

		// 保留两位小数精度
		pressures[i] = math.Round(pressure*100) / 100

		// 对于最后一个点，可能会有重复（基于真实数据观察）
		if i == pointCount-1 && api.rand.Float64() < 0.25 {
			// 25%概率最后一个点重复前一个点的压力值
			if i > 0 {
				pressures[i] = pressures[i-1]
			}
		}
	}

	return pressures
}

// getProgressiveOffset 获取渐进式偏移系数
func (api *SlideSimulatorAPI) getProgressiveOffset(progress float64) float64 {
	// 使用二次函数让偏移逐渐增加，模拟真实滑动中的累积偏移
	// 开始时偏移较小，中后期偏移逐渐增大
	return progress*progress*0.7 + progress*0.3
}

// calculateEndPoint 计算终点坐标
func (api *SlideSimulatorAPI) calculateEndPoint(startX, startY float64, direction Direction, distance float64) (float64, float64) {
	switch direction {
	case Up:
		return startX, startY - distance
	case Down:
		return startX, startY + distance
	case Left:
		return startX - distance, startY
	case Right:
		return startX + distance, startY
	default:
		return startX, startY
	}
}

// calculateDuration 计算滑动持续时间
func (api *SlideSimulatorAPI) calculateDuration(distance float64) int64 {
	// 基于真实数据的持续时间算法
	baseDuration := 120.0
	variableDuration := distance * 0.05
	randomFactor := api.rand.Float64()*40 - 20

	duration := baseDuration + variableDuration + randomFactor

	if duration < float64(api.config.MinDuration) {
		duration = float64(api.config.MinDuration)
	}
	if duration > float64(api.config.MaxDuration) {
		duration = float64(api.config.MaxDuration)
	}

	return int64(duration)
}

// calculatePointCount 计算轨迹点数量
func (api *SlideSimulatorAPI) calculatePointCount(duration int64) int {
	avgInterval := 20.0 + api.rand.Float64()*10
	count := int(float64(duration)/avgInterval) + 1

	if count < api.config.MinPoints {
		count = api.config.MinPoints
	}
	if count > api.config.MaxPoints {
		count = api.config.MaxPoints
	}

	return count
}

// generateTimestamps 生成时间戳序列
func (api *SlideSimulatorAPI) generateTimestamps(duration int64, pointCount int) []int64 {
	baseTime := time.Now().UnixMilli()
	timestamps := make([]int64, pointCount)

	timestamps[0] = baseTime

	for i := 1; i < pointCount; i++ {
		progress := float64(i) / float64(pointCount-1)
		timeProgress := api.speedCurve(progress)
		timestamps[i] = baseTime + int64(timeProgress*float64(duration))
	}

	return timestamps
}

// speedCurve 速度曲线函数
func (api *SlideSimulatorAPI) speedCurve(progress float64) float64 {
	// 模拟真实滑动的速度变化
	if progress <= 0.5 {
		return 0.8*progress*progress + 0.2*progress
	} else {
		return 0.2 + 0.8*(2*progress-1)
	}
}

// calculateBezierPoint 计算贝塞尔曲线点
func (api *SlideSimulatorAPI) calculateBezierPoint(startX, startY, endX, endY, progress float64, direction Direction) (float64, float64) {
	controlX, controlY := api.calculateControlPoint(startX, startY, endX, endY, direction)

	t := progress
	oneMinusT := 1 - t

	x := oneMinusT*oneMinusT*startX + 2*oneMinusT*t*controlX + t*t*endX
	y := oneMinusT*oneMinusT*startY + 2*oneMinusT*t*controlY + t*t*endY

	return x, y
}

// calculateControlPoint 计算控制点
func (api *SlideSimulatorAPI) calculateControlPoint(startX, startY, endX, endY float64, direction Direction) (float64, float64) {
	midX := (startX + endX) / 2
	midY := (startY + endY) / 2

	distance := math.Sqrt((endX-startX)*(endX-startX) + (endY-startY)*(endY-startY))

	var offsetX, offsetY float64

	switch direction {
	case Up, Down:
		// 垂直滑动时的X轴偏移：根据真实数据分析，平均偏移比例为25.8%
		// 偏移范围：距离的15%-35%
		offsetRatio := 0.15 + api.rand.Float64()*0.20 // 15%-35%
		maxOffsetX := distance * offsetRatio

		// 上滑时倾向于向右偏移，下滑时可以任意方向
		if direction == Up {
			offsetX = api.rand.Float64() * maxOffsetX // 0到最大偏移（向右）
		} else {
			offsetX = (api.rand.Float64() - 0.5) * maxOffsetX // 左右偏移
		}
		offsetY = 0

	case Left, Right:
		// 水平滑动时的Y轴偏移：根据真实数据分析，平均偏移比例为12.5%
		// 偏移范围：距离的5%-25%
		offsetRatio := 0.05 + api.rand.Float64()*0.20 // 5%-25%
		maxOffsetY := distance * offsetRatio

		offsetX = 0
		// 左滑时可能向上或向下偏移，右滑时偏移较小
		if direction == Left {
			offsetY = (api.rand.Float64() - 0.5) * maxOffsetY
		} else {
			// 右滑时偏移相对较小
			offsetY = (api.rand.Float64() - 0.5) * maxOffsetY * 0.7
		}
	}

	return midX + offsetX, midY + offsetY
}

// addNoise 添加随机噪声
func (api *SlideSimulatorAPI) addNoise(maxNoise float64) float64 {
	return (api.rand.Float64() - 0.5) * maxNoise
}

// calculateMetrics 计算滑动指标
func (api *SlideSimulatorAPI) calculateMetrics(points []SlidePoint) SlideMetrics {
	if len(points) == 0 {
		return SlideMetrics{}
	}

	totalDuration := points[len(points)-1].Timestamp - points[0].Timestamp

	// 计算实际距离
	var actualDistance float64
	for i := 1; i < len(points); i++ {
		dx := points[i].X - points[i-1].X
		dy := points[i].Y - points[i-1].Y
		actualDistance += math.Sqrt(dx*dx + dy*dy)
	}

	// 计算平均间隔
	var averageInterval float64
	if len(points) > 1 {
		averageInterval = float64(totalDuration) / float64(len(points)-1)
	}

	return SlideMetrics{
		TotalDuration:   totalDuration,
		PointCount:      len(points),
		ActualDistance:  actualDistance,
		AverageInterval: averageInterval,
	}
}

// ToJSON 将结果转换为JSON
func (resp SlideResponse) ToJSON() (string, error) {
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ConvertToTouchEvents 将SlidePoint切片转换为TouchEvent切片
func (api *SlideSimulatorAPI) ConvertToTouchEvents(points []SlidePoint) []types.TouchEvent {
	if len(points) == 0 {
		return nil
	}

	events := make([]types.TouchEvent, len(points))
	baseDownTime := points[0].Timestamp

	for i, point := range points {
		var action int
		if i == 0 {
			action = 0 // ACTION_DOWN
		} else if i == len(points)-1 {
			action = 1 // ACTION_UP
		} else {
			action = 2 // ACTION_MOVE
		}

		events[i] = types.TouchEvent{
			X:         point.X,
			Y:         point.Y,
			DeviceID:  point.DeviceID,
			Pressure:  float64(point.Pressure),
			Size:      point.Size,
			RawX:      point.X,      // 使用相同的X坐标
			RawY:      point.Y,      // 使用相同的Y坐标
			DownTime:  baseDownTime, // 第一个事件的时间戳作为DownTime
			EventTime: point.Timestamp,
			ToolType:  1, // TOOL_TYPE_FINGER
			Flag:      0, // 默认flag
			Action:    action,
		}
	}

	return events
}

// GenerateSlideWithRandomDistance 生成指定方向和随机距离的滑动轨迹
func (api *SlideSimulatorAPI) GenerateSlideWithRandomDistance(startX, startY float64, direction Direction, minDistance, maxDistance float64, deviceID int, pressure float64, size float64) ([]types.TouchEvent, error) {
	// 验证输入参数
	if minDistance <= 0 || maxDistance < minDistance {
		return nil, fmt.Errorf("invalid distance range: minDistance=%.2f, maxDistance=%.2f", minDistance, maxDistance)
	}

	// 计算实际滑动距离
	var actualDistance float64
	if minDistance == maxDistance {
		actualDistance = minDistance
	} else {
		actualDistance = minDistance + api.rand.Float64()*(maxDistance-minDistance)
	}

	// 构建滑动请求
	req := SlideRequest{
		StartX:    startX,
		StartY:    startY,
		Direction: direction,
		Distance:  actualDistance,
		DeviceID:  deviceID,
		Pressure:  pressure,
		Size:      size,
	}

	// 生成滑动轨迹
	response := api.GenerateSlide(req)
	if !response.Success {
		return nil, fmt.Errorf("generate slide failed: %s", response.Message)
	}

	// 转换为TouchEvent
	events := api.ConvertToTouchEvents(response.Points)
	return events, nil
}

// GenerateSlideInArea 在指定区域内生成滑动轨迹
func (api *SlideSimulatorAPI) GenerateSlideInArea(areaStartX, areaStartY, areaEndX, areaEndY float64, direction Direction, minDistance, maxDistance float64, deviceID int, pressure float64, size float64) ([]types.TouchEvent, error) {
	// 验证输入参数
	if minDistance <= 0 || maxDistance < minDistance {
		return nil, fmt.Errorf("invalid distance range: minDistance=%.2f, maxDistance=%.2f", minDistance, maxDistance)
	}

	// 验证区域参数（允许start和end相等，表示单点区域）
	if areaStartX > areaEndX || areaStartY > areaEndY {
		return nil, fmt.Errorf("invalid area: start point (%.2f, %.2f) should be less than or equal to end point (%.2f, %.2f)",
			areaStartX, areaStartY, areaEndX, areaEndY)
	}

	// 在区域内随机选择起始点（如果start和end相等，则使用固定点）
	var randomStartX, randomStartY float64

	if areaStartX == areaEndX {
		randomStartX = areaStartX // 单点X坐标
	} else {
		areaWidth := areaEndX - areaStartX
		randomStartX = areaStartX + api.rand.Float64()*areaWidth
	}

	if areaStartY == areaEndY {
		randomStartY = areaStartY // 单点Y坐标
	} else {
		areaHeight := areaEndY - areaStartY
		randomStartY = areaStartY + api.rand.Float64()*areaHeight
	}

	// 计算实际滑动距离
	var actualDistance float64
	if minDistance == maxDistance {
		actualDistance = minDistance
	} else {
		actualDistance = minDistance + api.rand.Float64()*(maxDistance-minDistance)
	}

	// 验证滑动后的点是否会超出屏幕边界(这里做简单检查)
	// 可以根据实际需要调整边界检查逻辑
	endX, endY := api.calculateEndPoint(randomStartX, randomStartY, direction, actualDistance)

	// 如果滑动后超出合理范围，调整起始点位置
	const marginBuffer = 50.0 // 边界缓冲区
	switch direction {
	case Up:
		if endY < marginBuffer {
			randomStartY = math.Min(areaEndY-marginBuffer, randomStartY+actualDistance)
		}
	case Down:
		// 这里假设屏幕高度最大为2400，可以根据实际需要调整
		if endY > 2400-marginBuffer {
			randomStartY = math.Max(areaStartY+marginBuffer, randomStartY-actualDistance)
		}
	case Left:
		if endX < marginBuffer {
			randomStartX = math.Min(areaEndX-marginBuffer, randomStartX+actualDistance)
		}
	case Right:
		// 这里假设屏幕宽度最大为1800，可以根据实际需要调整
		if endX > 1800-marginBuffer {
			randomStartX = math.Max(areaStartX+marginBuffer, randomStartX-actualDistance)
		}
	}

	// 构建滑动请求
	req := SlideRequest{
		StartX:    randomStartX,
		StartY:    randomStartY,
		Direction: direction,
		Distance:  actualDistance,
		DeviceID:  deviceID,
		Pressure:  pressure,
		Size:      size,
	}

	// 生成滑动轨迹
	response := api.GenerateSlide(req)
	if !response.Success {
		return nil, fmt.Errorf("generate slide failed: %s", response.Message)
	}

	// 转换为TouchEvent
	events := api.ConvertToTouchEvents(response.Points)
	return events, nil
}

// GeneratePointToPointSlideEvents 生成点对点滑动的TouchEvent序列
func (api *SlideSimulatorAPI) GeneratePointToPointSlideEvents(startX, startY, endX, endY float64, deviceID int, pressure float64, size float64) ([]types.TouchEvent, error) {
	// 验证输入参数
	if startX == endX && startY == endY {
		return nil, fmt.Errorf("start point (%.2f, %.2f) and end point (%.2f, %.2f) cannot be the same", startX, startY, endX, endY)
	}

	// 计算距离
	distance := math.Sqrt((endX-startX)*(endX-startX) + (endY-startY)*(endY-startY))
	if distance < 10 {
		return nil, fmt.Errorf("distance too short: %.2f pixels", distance)
	}

	// 构建点对点滑动请求
	req := PointToPointSlideRequest{
		StartX:   startX,
		StartY:   startY,
		EndX:     endX,
		EndY:     endY,
		DeviceID: deviceID,
		Pressure: pressure,
		Size:     size,
	}

	// 生成滑动轨迹
	response := api.GeneratePointToPointSlide(req)
	if !response.Success {
		return nil, fmt.Errorf("generate point to point slide failed: %s", response.Message)
	}

	// 转换为TouchEvent
	events := api.ConvertToTouchEvents(response.Points)
	return events, nil
}
