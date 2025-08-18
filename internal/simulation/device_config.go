package simulation

import (
	"math/rand"
	"time"
)

type DeviceConfig struct {
	DeviceID    int
	PressureMin float64
	PressureMax float64
	SizeMin     float64
	SizeMax     float64
}

// DeviceParams 设备参数结构体
type DeviceParams struct {
	DeviceID int
	Pressure float64
	Size     float64
}

// GetRandomDeviceParams 根据设备型号获取随机的设备参数
func GetRandomDeviceParams(deviceModel string) DeviceParams {
	config := getDeviceConfig(deviceModel)

	// 创建随机数生成器
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 在最小值和最大值之间生成随机数
	randomPressure := config.PressureMin + rng.Float64()*(config.PressureMax-config.PressureMin)
	randomSize := config.SizeMin + rng.Float64()*(config.SizeMax-config.SizeMin)

	// 保留合理的精度
	randomPressure = float64(int(randomPressure*100)) / 100 // 保留2位小数
	randomSize = float64(int(randomSize*1000)) / 1000       // 保留3位小数

	return DeviceParams{
		DeviceID: config.DeviceID,
		Pressure: randomPressure,
		Size:     randomSize,
	}
}

// getDeviceConfig returns device-specific configuration based on device model
func getDeviceConfig(deviceModel string) DeviceConfig {
	switch deviceModel {
	// "HUAWEI"
	case "SEA-AL00": // 华为nova5
		return DeviceConfig{
			DeviceID:    1,
			PressureMin: 1.2,
			PressureMax: 1.8,
			SizeMin:     160.0,
			SizeMax:     200.0,
		}
	case "ABR-AL00": // 华为P50
		return DeviceConfig{
			DeviceID:    3,
			PressureMin: 1.4,
			PressureMax: 2.0,
			SizeMin:     170.0,
			SizeMax:     220.0,
		}
	case "SEA-AL10": // 华为nova5Pro
		return DeviceConfig{
			DeviceID:    3,
			PressureMin: 1.3,
			PressureMax: 1.9,
			SizeMin:     165.0,
			SizeMax:     210.0,
		}
	case "ANA-AN00": // 华为P40
		return DeviceConfig{
			DeviceID:    4,
			PressureMin: 1.5,
			PressureMax: 2.2,
			SizeMin:     180.0,
			SizeMax:     230.0,
		}
	case "ELS-AN00": // 华为P40Pro
		return DeviceConfig{
			DeviceID:    5,
			PressureMin: 1.6,
			PressureMax: 2.3,
			SizeMin:     185.0,
			SizeMax:     240.0,
		}
	case "NCO_AL00":
		return DeviceConfig{
			DeviceID:    3,
			PressureMin: 3,
			PressureMax: 7,
			SizeMin:     140.0,
			SizeMax:     200.0,
		}

	// "Xiaomi"
	case "M2007J22C": // RedmiNote9 5G
		return DeviceConfig{
			DeviceID:    3,
			PressureMin: 1.3,
			PressureMax: 1.9,
			SizeMin:     170.0,
			SizeMax:     215.0,
		}
	case "2211133C": // 小米13
		return DeviceConfig{
			DeviceID:    7,
			PressureMin: 1.7,
			PressureMax: 2.4,
			SizeMin:     190.0,
			SizeMax:     250.0,
		}
	case "2206123SC": // 小米12s
		return DeviceConfig{
			DeviceID:    8,
			PressureMin: 1.6,
			PressureMax: 2.3,
			SizeMin:     185.0,
			SizeMax:     245.0,
		}
	case "21091116C":
		return DeviceConfig{
			DeviceID:    5,
			PressureMin: 1,
			PressureMax: 1,
			SizeMin:     0,
			SizeMax:     1,
		}

	// "Google"
	case "Pixel 6 Pro":
		return DeviceConfig{
			DeviceID:    4,
			PressureMin: 1.4,
			PressureMax: 2.1,
			SizeMin:     175.0,
			SizeMax:     225.0,
		}

		// "Google"
	case "iphone":
		return DeviceConfig{
			DeviceID:    2,
			PressureMin: 1,
			PressureMax: 1,
			SizeMin:     0.03,
			SizeMax:     0.04,
		}

	// Default configuration for unknown devices
	default:
		return DeviceConfig{
			DeviceID:    6,
			PressureMin: 1.2,
			PressureMax: 2.0,
			SizeMin:     160.0,
			SizeMax:     220.0,
		}
	}
}
