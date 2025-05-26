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
	cacheKey := config.Serial
	if cacheKey == "" {
		return nil, fmt.Errorf("serial cannot be empty")
	}

	// Check if driver exists in cache
	if cachedItem, ok := driverCache.Load(cacheKey); ok {
		if cached, ok := cachedItem.(*CachedXTDriver); ok {
			log.Info().Str("serial", cached.Serial).Msg("Using cached XTDriver")

			// Increment reference count
			cached.RefCount++
			return cached.Driver, nil
		}
	}

	// Create new driver
	driverExt, err := createXTDriverWithConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create XTDriver: %w", err)
	}

	// Cache the driver
	cached := &CachedXTDriver{
		Platform: config.Platform,
		Driver:   driverExt,
		Serial:   config.Serial,
		RefCount: 1,
	}
	driverCache.Store(cacheKey, cached)

	log.Info().
		Str("platform", config.Platform).
		Str("serial", config.Serial).
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

	if config.Serial == "" {
		return nil, fmt.Errorf("serial is empty")
	}

	// Create device based on platform and configuration
	var device IDevice
	var err error

	// Try to create device with specific options first
	if config.DeviceOpts != nil {
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
		}
	} else {
		device, err = NewDeviceWithDefault(platform, config.Serial)
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

				// Clean up driver resources
				if err := cached.Driver.DeleteSession(); err != nil {
					log.Warn().Err(err).Str("serial", serial).Msg("Failed to delete driver session")
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
				// Clean up driver resources
				if err := cached.Driver.DeleteSession(); err != nil {
					log.Warn().Err(err).Str("serial", serial).Msg("Failed to delete driver session")
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

// setupXTDriver initializes an XTDriver based on the platform and serial.
// This function is kept for backward compatibility with MCP integration
func setupXTDriver(_ context.Context, args map[string]any) (*XTDriver, error) {
	platform, _ := args["platform"].(string)
	serial, _ := args["serial"].(string)

	config := DriverCacheConfig{
		Platform: platform,
		Serial:   serial,
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
