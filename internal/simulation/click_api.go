package simulation

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/httprunner/httprunner/v5/uixt/types"
)

// ClickRequest 点击请求参数
type ClickRequest struct {
	X        float64 `json:"x"`         // 点击X坐标
	Y        float64 `json:"y"`         // 点击Y坐标
	DeviceID int     `json:"device_id"` // 设备ID
	Pressure float64 `json:"pressure"`  // 压力值
	Size     float64 `json:"size"`      // 接触面积
}

// ClickResponse 点击响应结果
type ClickResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Points  []ClickPoint `json:"points"`
	Metrics ClickMetrics `json:"metrics"`
}

// ClickMetrics 点击指标
type ClickMetrics struct {
	TotalDuration   int64   `json:"total_duration_ms"`   // 总持续时间(毫秒)
	PointCount      int     `json:"point_count"`         // 轨迹点数量
	MaxDeviation    float64 `json:"max_deviation"`       // 最大偏移距离
	AverageInterval float64 `json:"average_interval_ms"` // 平均采样间隔
}

// ClickPoint 点击轨迹点
type ClickPoint struct {
	Timestamp int64   `json:"timestamp"`  // 时间戳(毫秒)
	X         float64 `json:"x"`          // X坐标
	Y         float64 `json:"y"`          // Y坐标
	DeviceID  int     `json:"device_id"`  // 设备ID
	Pressure  float64 `json:"pressure"`   // 压力值
	Size      float64 `json:"size"`       // 接触面积
	Action    int     `json:"action"`     // 动作类型(0=按下,1=抬起,2=移动)
	EventTime int64   `json:"event_time"` // 相对第一个点的时间(ms)，第一个点为0
}

// ClickConfig 点击配置参数
type ClickConfig struct {
	MinDuration  int64   // 最小持续时间(毫秒)
	MaxDuration  int64   // 最大持续时间(毫秒)
	MinPoints    int     // 最小点数
	MaxPoints    int     // 最大点数
	MaxDeviation float64 // 最大坐标偏移(像素)
	NoiseLevel   float64 // 噪声级别
}

// DefaultClickConfig 默认配置
var DefaultClickConfig = ClickConfig{
	MinDuration:  40,
	MaxDuration:  90,
	MinPoints:    4, // 增加最小点数从3到4，确保至少有2个MOVE事件
	MaxPoints:    6,
	MaxDeviation: 2.0,
	NoiseLevel:   0.5,
}

// ClickSimulatorAPI 点击仿真API
type ClickSimulatorAPI struct {
	rand   *rand.Rand
	config ClickConfig
}

// TestCase 测试用例
type ClickTestCase struct {
	Name     string
	X        float64
	Y        float64
	DeviceID int
	Pressure float64
	Size     float64
}

// NewClickSimulatorAPI 创建新的点击仿真API
func NewClickSimulatorAPI(config *ClickConfig) *ClickSimulatorAPI {
	if config == nil {
		config = &DefaultClickConfig
	}

	return &ClickSimulatorAPI{
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		config: *config,
	}
}

// GenerateClick 生成点击轨迹
func (api *ClickSimulatorAPI) GenerateClick(req ClickRequest) ClickResponse {
	// 验证输入参数
	if err := api.validateRequest(req); err != nil {
		return ClickResponse{
			Success: false,
			Message: err.Error(),
		}
	}

	// 生成点击轨迹
	points := api.generateClickPoints(req)

	// 计算指标
	metrics := api.calculateMetrics(points, req.X, req.Y)

	return ClickResponse{
		Success: true,
		Points:  points,
		Metrics: metrics,
	}
}

// validateRequest 验证请求参数
func (api *ClickSimulatorAPI) validateRequest(req ClickRequest) error {
	if req.X < 0 || req.Y < 0 {
		return fmt.Errorf("coordinates must be non-negative")
	}

	if req.DeviceID < 0 {
		return fmt.Errorf("device_id must be non-negative")
	}

	return nil
}

// generateClickPoints 生成点击轨迹点
func (api *ClickSimulatorAPI) generateClickPoints(req ClickRequest) []ClickPoint {
	// 计算点击参数
	duration := api.calculateDuration()
	pointCount := api.calculatePointCount()

	// 生成时间戳序列
	timestamps := api.generateTimestamps(duration, pointCount)

	// 生成轨迹点
	points := make([]ClickPoint, pointCount)

	// 生成size变化曲线（基于真实数据分析）
	sizeValues := api.generateSizeValues(pointCount, req.Size)

	// 生成压力值序列
	pressureValues := api.generatePressureValues(pointCount, req.Pressure)

	// 生成坐标偏移序列
	xOffsets, yOffsets := api.generateCoordinateOffsets(pointCount)

	baseTimestamp := timestamps[0]
	for i := 0; i < pointCount; i++ {
		// 计算当前坐标（添加轻微偏移）
		currentX := req.X + xOffsets[i]
		currentY := req.Y + yOffsets[i]

		// 确定动作类型
		var action int
		if i == 0 {
			action = 0 // 按下
		} else if i == pointCount-1 {
			action = 1 // 抬起
		} else {
			action = 2 // 移动
		}

		eventTime := timestamps[i] - baseTimestamp

		points[i] = ClickPoint{
			Timestamp: timestamps[i],
			X:         currentX,
			Y:         currentY,
			DeviceID:  req.DeviceID,
			Pressure:  pressureValues[i],
			Size:      sizeValues[i],
			Action:    action,
			EventTime: eventTime,
		}
	}

	return points
}

// generatePressureValues 生成pressure值序列，基于用户输入的压力值动态仿真点击操作
func (api *ClickSimulatorAPI) generatePressureValues(pointCount int, basePressure float64) []float64 {
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

	// 基于真实点击数据观察的压力变化规律：
	// 点击操作的pressure变化特点：快速上升→短暂保持峰值→快速下降
	// 1. 起始压力：基础压力的95%-105%
	// 2. 峰值压力：基础压力的102%-108% (相对滑动，点击的峰值增幅较小)
	// 3. 结束压力：基础压力的25%-35% (快速下降到较低值)

	startPressureRatio := 0.95 + api.rand.Float64()*0.10 // 95%-105%
	peakPressureRatio := 1.02 + api.rand.Float64()*0.06  // 102%-108%
	endPressureRatio := 0.25 + api.rand.Float64()*0.10   // 25%-35%

	startPressure := baseP * startPressureRatio
	peakPressure := baseP * peakPressureRatio
	endPressure := baseP * endPressureRatio

	// 点击操作的峰值通常出现在早期(第2-3个点)
	var peakIndex int
	if pointCount <= 3 {
		peakIndex = 1 // 对于短序列，峰值在第2个点
	} else {
		peakIndex = 1 + api.rand.Intn(2) // 峰值在第2或第3个点
	}
	if peakIndex >= pointCount {
		peakIndex = pointCount - 2
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

		if i == 0 {
			// 第一个点：起始压力
			pressure = startPressure
		} else if i <= peakIndex {
			// 上升到峰值阶段
			pressure = peakPressure
		} else if i == pointCount-1 {
			// 最后一个点：结束压力
			pressure = endPressure
		} else {
			// 从峰值下降到结束压力的过渡阶段
			t := float64(i-peakIndex) / float64(pointCount-1-peakIndex)
			pressure = peakPressure + (endPressure-peakPressure)*t
		}

		// 添加随机噪声（±3%），点击操作的噪声相对较小
		noiseRange := pressure * 0.03
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
	}

	return pressures
}

// calculateDuration 计算点击持续时间
func (api *ClickSimulatorAPI) calculateDuration() int64 {
	// 基于真实数据的持续时间算法
	baseDuration := float64(api.config.MinDuration+api.config.MaxDuration) / 2
	randomFactor := api.rand.Float64()*float64(api.config.MaxDuration-api.config.MinDuration) -
		float64(api.config.MaxDuration-api.config.MinDuration)/2

	duration := baseDuration + randomFactor

	if duration < float64(api.config.MinDuration) {
		duration = float64(api.config.MinDuration)
	}
	if duration > float64(api.config.MaxDuration) {
		duration = float64(api.config.MaxDuration)
	}

	return int64(duration)
}

// calculatePointCount 计算轨迹点数量
func (api *ClickSimulatorAPI) calculatePointCount() int {
	// 基于真实数据分析，点击通常有3-6个点
	count := api.config.MinPoints + api.rand.Intn(api.config.MaxPoints-api.config.MinPoints+1)
	return count
}

// generateTimestamps 生成时间戳序列
func (api *ClickSimulatorAPI) generateTimestamps(duration int64, pointCount int) []int64 {
	baseTime := time.Now().UnixMilli()
	timestamps := make([]int64, pointCount)

	timestamps[0] = baseTime

	if pointCount == 1 {
		return timestamps
	}

	// 基于真实数据的时间间隔模式
	for i := 1; i < pointCount; i++ {
		// 时间间隔：前期较短，后期可能较长
		var intervalRatio float64
		if i == 1 {
			// 第一个间隔较短 (8-30ms)
			intervalRatio = 0.1 + api.rand.Float64()*0.2 // 10%-30%
		} else if i == pointCount-1 {
			// 最后一个间隔可能较短
			intervalRatio = 0.1 + api.rand.Float64()*0.15 // 10%-25%
		} else {
			// 中间间隔相对均匀
			intervalRatio = 0.15 + api.rand.Float64()*0.25 // 15%-40%
		}

		interval := int64(float64(duration) * intervalRatio)
		timestamps[i] = timestamps[i-1] + interval
	}

	// 确保最后一个时间戳不超过总持续时间
	if timestamps[pointCount-1] > baseTime+duration {
		timestamps[pointCount-1] = baseTime + duration
	}

	return timestamps
}

// generateSizeValues 生成size值序列，基于真实数据分析
func (api *ClickSimulatorAPI) generateSizeValues(pointCount int, baseSize float64) []float64 {
	sizes := make([]float64, pointCount)

	// 如果baseSize为0，使用默认值
	if baseSize == 0 {
		baseSize = 0.043 // 默认size值，基于真实数据平均值
	}

	// 动态计算size范围，基于baseSize的值来适应不同设备
	var minSize, maxSize float64
	if baseSize < 1.0 {
		// 小数值范围（如0.043），使用原有逻辑
		minSize = 0.035
		maxSize = 0.051
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
		// 基础size值随点击进度变化
		var sizeModifier float64

		if i == 0 {
			// 开始时：可能较小
			sizeModifier = 0.85 + api.rand.Float64()*0.3 // 0.85-1.15倍
		} else if i == pointCount-1 {
			// 结束时：可能减小（手指抬起）
			sizeModifier = 0.8 + api.rand.Float64()*0.25 // 0.8-1.05倍
		} else {
			// 中间过程：可能增大（压力增加）
			sizeModifier = 0.95 + api.rand.Float64()*0.25 // 0.95-1.2倍
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
			noiseLevel = 0.002 // 小数值使用固定的小噪声
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

// generateCoordinateOffsets 生成坐标偏移序列
func (api *ClickSimulatorAPI) generateCoordinateOffsets(pointCount int) ([]float64, []float64) {
	xOffsets := make([]float64, pointCount)
	yOffsets := make([]float64, pointCount)

	// 第一个点不偏移
	xOffsets[0] = 0
	yOffsets[0] = 0

	if pointCount == 1 {
		return xOffsets, yOffsets
	}

	// 基于真实数据分析，点击时会有轻微的移动
	for i := 1; i < pointCount; i++ {
		// 累积偏移，模拟手指的轻微移动
		maxOffset := api.config.MaxDeviation * float64(i) / float64(pointCount-1)

		// 添加随机偏移
		xOffsets[i] = xOffsets[i-1] + api.addNoise(maxOffset*0.5)
		yOffsets[i] = yOffsets[i-1] + api.addNoise(maxOffset*0.5)

		// 限制总偏移量
		if math.Abs(xOffsets[i]) > api.config.MaxDeviation {
			if xOffsets[i] > 0 {
				xOffsets[i] = api.config.MaxDeviation
			} else {
				xOffsets[i] = -api.config.MaxDeviation
			}
		}
		if math.Abs(yOffsets[i]) > api.config.MaxDeviation {
			if yOffsets[i] > 0 {
				yOffsets[i] = api.config.MaxDeviation
			} else {
				yOffsets[i] = -api.config.MaxDeviation
			}
		}
	}

	return xOffsets, yOffsets
}

// addNoise 添加随机噪声
func (api *ClickSimulatorAPI) addNoise(maxNoise float64) float64 {
	return (api.rand.Float64() - 0.5) * maxNoise * 2
}

// calculateMetrics 计算点击指标
func (api *ClickSimulatorAPI) calculateMetrics(points []ClickPoint, originX, originY float64) ClickMetrics {
	if len(points) == 0 {
		return ClickMetrics{}
	}

	totalDuration := points[len(points)-1].Timestamp - points[0].Timestamp

	// 计算最大偏移距离
	var maxDeviation float64
	for _, point := range points {
		deviation := math.Sqrt((point.X-originX)*(point.X-originX) + (point.Y-originY)*(point.Y-originY))
		if deviation > maxDeviation {
			maxDeviation = deviation
		}
	}

	// 计算平均间隔
	var averageInterval float64
	if len(points) > 1 {
		averageInterval = float64(totalDuration) / float64(len(points)-1)
	}

	return ClickMetrics{
		TotalDuration:   totalDuration,
		PointCount:      len(points),
		MaxDeviation:    maxDeviation,
		AverageInterval: averageInterval,
	}
}

// ToJSON 将结果转换为JSON
func (resp ClickResponse) ToJSON() (string, error) {
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ConvertClickToTouchEvents 将ClickPoint切片转换为TouchEvent切片
func (api *ClickSimulatorAPI) ConvertClickToTouchEvents(points []ClickPoint) []types.TouchEvent {
	if len(points) == 0 {
		return nil
	}

	events := make([]types.TouchEvent, len(points))
	baseDownTime := points[0].Timestamp

	for i, point := range points {
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
			ToolType:  1,            // TOOL_TYPE_FINGER
			Flag:      0,            // 默认flag
			Action:    point.Action, // 直接使用point的action
		}
	}

	return events
}

// GenerateClickEvents 生成点击的TouchEvent序列
func (api *ClickSimulatorAPI) GenerateClickEvents(x, y float64, deviceID int, pressure float64, size float64) ([]types.TouchEvent, error) {
	// 验证输入参数
	if x < 0 || y < 0 {
		return nil, fmt.Errorf("coordinates must be non-negative: x=%.2f, y=%.2f", x, y)
	}

	// 构建点击请求
	req := ClickRequest{
		X:        x,
		Y:        y,
		DeviceID: deviceID,
		Pressure: pressure,
		Size:     size,
	}

	// 生成点击轨迹
	response := api.GenerateClick(req)
	if !response.Success {
		return nil, fmt.Errorf("generate click failed: %s", response.Message)
	}

	// 转换为TouchEvent
	events := api.ConvertClickToTouchEvents(response.Points)
	return events, nil
}
