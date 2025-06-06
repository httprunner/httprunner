package uixt

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/rs/zerolog/log"
)

var driverCache sync.Map // key is serial, value is *CachedXTDriver

// CachedXTDriver wraps XTDriver with additional cache metadata
type CachedXTDriver struct {
	Platform string
	Serial   string
	Driver   *XTDriver
	RefCount int32 // reference count for resource management
}

// DriverCacheConfig holds configuration for driver creation
type DriverCacheConfig struct {
	Platform   string
	Serial     string
	AIOptions  []option.AIServiceOption
	DeviceOpts *option.DeviceOptions // unified device options
}

// GetOrCreateXTDriver gets an existing driver from cache or creates a new one
func GetOrCreateXTDriver(config DriverCacheConfig) (*XTDriver, error) {
	// If serial is specified, check cache first
	if config.Serial != "" {
		cacheKey := config.Serial
		if cachedItem, ok := driverCache.Load(cacheKey); ok {
			if cached, ok := cachedItem.(*CachedXTDriver); ok {
				log.Info().Str("serial", cached.Serial).Msg("Using cached XTDriver")

				// Increment reference count
				cached.RefCount++
				return cached.Driver, nil
			}
		}
	}

	// If no serial specified, try to find existing driver
	if config.Serial == "" {
		if driver := findCachedDriver(config.Platform); driver != nil {
			return driver, nil
		}
	}

	// Create new driver (will auto-detect serial if empty)
	driverExt, err := createXTDriverWithConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create XTDriver: %w", err)
	}

	// Get actual serial from the created driver
	actualSerial := driverExt.GetDevice().UUID()

	// Check if a driver with this actual serial already exists in cache
	if cachedItem, ok := driverCache.Load(actualSerial); ok {
		if cached, ok := cachedItem.(*CachedXTDriver); ok {
			log.Info().Str("serial", actualSerial).Msg("Found existing cached XTDriver with detected serial")

			// Clean up the newly created driver since we have a cached one
			if err := driverExt.DeleteSession(); err != nil {
				log.Warn().Err(err).Str("serial", actualSerial).Msg("Failed to delete newly created driver session")
			}

			// Increment reference count and return cached driver
			cached.RefCount++
			return cached.Driver, nil
		}
	}

	// Cache the new driver with actual serial
	cached := &CachedXTDriver{
		Platform: config.Platform,
		Driver:   driverExt,
		Serial:   actualSerial,
		RefCount: 1,
	}
	driverCache.Store(actualSerial, cached)

	log.Info().
		Str("platform", config.Platform).
		Str("serial", actualSerial).
		Msg("Created and cached new XTDriver")

	return driverExt, nil
}

// createXTDriverWithConfig creates a new XTDriver based on configuration
func createXTDriverWithConfig(config DriverCacheConfig) (*XTDriver, error) {
	platform := config.Platform
	if platform == "" {
		log.Warn().Msg("platform is not set, using android as default")
		platform = "android"
	}

	// Create device based on platform and configuration
	var device IDevice
	var err error

	// Create device based on platform and configuration
	if config.DeviceOpts != nil {
		// Use specific device options
		switch strings.ToLower(platform) {
		case "android":
			androidOpts := config.DeviceOpts.ToAndroidOptions().Options()
			device, err = NewAndroidDevice(androidOpts...)
		case "ios":
			iosOpts := config.DeviceOpts.ToIOSOptions().Options()
			device, err = NewIOSDevice(iosOpts...)
		case "harmony":
			harmonyOpts := config.DeviceOpts.ToHarmonyOptions().Options()
			device, err = NewHarmonyDevice(harmonyOpts...)
		case "browser":
			browserOpts := config.DeviceOpts.ToBrowserOptions().Options()
			device, err = NewBrowserDevice(browserOpts...)
		default:
			return nil, fmt.Errorf("unsupported platform: %s", platform)
		}
	} else {
		// Use default options, let NewXXDevice handle serial (empty or specified)
		switch strings.ToLower(platform) {
		case "android":
			if config.Serial != "" {
				device, err = NewAndroidDevice(option.WithSerialNumber(config.Serial))
			} else {
				device, err = NewAndroidDevice()
			}
		case "ios":
			if config.Serial != "" {
				device, err = NewIOSDevice(option.WithUDID(config.Serial))
			} else {
				device, err = NewIOSDevice()
			}
		case "harmony":
			if config.Serial != "" {
				device, err = NewHarmonyDevice(option.WithConnectKey(config.Serial))
			} else {
				device, err = NewHarmonyDevice()
			}
		case "browser":
			if config.Serial != "" {
				device, err = NewBrowserDevice(option.WithBrowserID(config.Serial))
			} else {
				device, err = NewBrowserDevice()
			}
		default:
			return nil, fmt.Errorf("unsupported platform: %s", platform)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	// Create driver
	driver, err := device.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	// Create XTDriver with AI options
	aiOpts := config.AIOptions
	if len(aiOpts) == 0 {
		// Default AI options
		aiOpts = []option.AIServiceOption{
			option.WithCVService(option.CVServiceTypeVEDEM),
		}
	}

	driverExt, err := NewXTDriver(driver, aiOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create XTDriver: %w", err)
	}
	return driverExt, nil
}

// ReleaseXTDriver decrements reference count and removes from cache when count reaches zero
func ReleaseXTDriver(serial string) error {
	if cachedItem, ok := driverCache.Load(serial); ok {
		if cached, ok := cachedItem.(*CachedXTDriver); ok {
			cached.RefCount--
			log.Debug().
				Str("serial", serial).
				Int32("refCount", cached.RefCount).
				Msg("Released XTDriver reference")

			// If no more references, clean up and remove from cache
			if cached.RefCount <= 0 {
				driverCache.Delete(serial)

				// Clean up driver resources if driver has underlying IDriver
				if cached.Driver != nil && cached.Driver.IDriver != nil {
					if err := cached.Driver.DeleteSession(); err != nil {
						log.Warn().Err(err).Str("serial", serial).Msg("Failed to delete driver session")
					}
				}

				log.Info().Str("serial", serial).Msg("Cleaned up XTDriver from cache")
			}
		}
	}
	return nil
}

// CleanupAllDrivers cleans up all cached drivers
func CleanupAllDrivers() {
	driverCache.Range(func(key, value interface{}) bool {
		if serial, ok := key.(string); ok {
			if cached, ok := value.(*CachedXTDriver); ok {
				// Clean up driver resources if driver has underlying IDriver
				if cached.Driver != nil && cached.Driver.IDriver != nil {
					if err := cached.Driver.DeleteSession(); err != nil {
						log.Warn().Err(err).Str("serial", serial).Msg("Failed to delete driver session")
					}
				}
				log.Info().Str("serial", serial).Msg("Cleaned up XTDriver from cache")
			}
			driverCache.Delete(serial)
		}
		return true
	})
}

// ListCachedDrivers returns information about all cached drivers
func ListCachedDrivers() []CachedXTDriver {
	var drivers []CachedXTDriver
	driverCache.Range(func(key, value interface{}) bool {
		if cached, ok := value.(*CachedXTDriver); ok {
			drivers = append(drivers, *cached)
		}
		return true
	})
	return drivers
}

// findCachedDriver searches for a cached driver by platform
// If platform is empty, returns any available driver
func findCachedDriver(platform string) *XTDriver {
	var foundDriver *XTDriver
	driverCache.Range(func(key, value interface{}) bool {
		serial, ok := key.(string)
		if !ok {
			return true // continue iteration
		}

		cached, ok := value.(*CachedXTDriver)
		if !ok {
			return true // continue iteration
		}

		// If platform is specified, match platform; otherwise use any available driver
		if platform == "" || cached.Platform == platform {
			foundDriver = cached.Driver
			cached.RefCount++

			if platform != "" {
				log.Debug().Str("platform", platform).Str("serial", serial).Msg("Using cached XTDriver by platform")
			} else {
				log.Debug().Str("serial", serial).Msg("Using any available cached XTDriver")
			}
			return false // stop iteration
		}

		return true // continue iteration
	})
	return foundDriver
}

// setupXTDriver initializes an XTDriver based on the platform and serial.
// This function is kept for backward compatibility with MCP integration
func setupXTDriver(_ context.Context, args map[string]any) (*XTDriver, error) {
	platform, _ := args["platform"].(string)
	serial, _ := args["serial"].(string)

	// Extract AI service options from arguments if provided
	var aiOpts []option.AIServiceOption

	// Check for LLM service type
	if llmService, ok := args["llm_service"].(string); ok && llmService != "" {
		aiOpts = append(aiOpts, option.WithLLMService(option.LLMServiceType(llmService)))
	}

	// Check for CV service type
	if cvService, ok := args["cv_service"].(string); ok && cvService != "" {
		aiOpts = append(aiOpts, option.WithCVService(option.CVServiceType(cvService)))
	}

	config := DriverCacheConfig{
		Platform:  platform,
		Serial:    serial,
		AIOptions: aiOpts,
	}
	return GetOrCreateXTDriver(config)
}

// RegisterXTDriver registers an externally created XTDriver to the unified cache
func RegisterXTDriver(serial string, driver *XTDriver) error {
	if serial == "" {
		return fmt.Errorf("serial cannot be empty")
	}
	if driver == nil {
		return fmt.Errorf("driver cannot be nil")
	}

	cached := &CachedXTDriver{
		Driver:   driver,
		Serial:   serial,
		RefCount: 1,
	}
	driverCache.Store(serial, cached)

	log.Info().
		Str("serial", serial).
		Msg("Registered external XTDriver to unified cache")

	return nil
}

// getXTDriverFromCache gets XTDriver from cache using device UUID
func getXTDriverFromCache(driver IDriver) *XTDriver {
	// Get device info to find the corresponding XTDriver
	device := driver.GetDevice()
	if device == nil {
		log.Warn().Msg("Cannot get device from driver for MCP hook")
		return nil
	}

	// Get device UUID (serial/udid/connectKey/browserID)
	deviceUUID := device.UUID()
	if deviceUUID == "" {
		log.Warn().Msg("Cannot get device UUID for MCP hook")
		return nil
	}

	// Get XTDriver from cache using device UUID as serial
	cachedDrivers := ListCachedDrivers()
	for _, cached := range cachedDrivers {
		if cached.Serial == deviceUUID {
			return cached.Driver
		}
	}

	log.Warn().Str("uuid", deviceUUID).
		Msg("Cannot find cached XTDriver for MCP hook")
	return nil
}
