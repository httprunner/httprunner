package uixt

import (
	"testing"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to clean up cache before each test
func setupTest() {
	CleanupAllDrivers()
}

func TestGetOrCreateXTDriver_EmptySerial_AutoDetect(t *testing.T) {
	setupTest()

	config := DriverCacheConfig{
		Platform: "android",
		Serial:   "", // Empty serial will be auto-detected by NewAndroidDevice
	}

	driver, err := GetOrCreateXTDriver(config)
	// Auto-detection may succeed or fail depending on test environment
	if err != nil {
		// If device creation fails (no devices or multiple devices)
		assert.Nil(t, driver)
		assert.Contains(t, err.Error(), "failed to create XTDriver")
	} else {
		// If device creation succeeds (exactly one device connected)
		assert.NotNil(t, driver)
		// Verify that a driver was created and cached with actual serial
		drivers := ListCachedDrivers()
		assert.Len(t, drivers, 1)
		assert.NotEmpty(t, drivers[0].Serial) // Serial should be populated with actual device serial
	}
}

func TestGetOrCreateXTDriver_EmptySerial_DefaultPlatform(t *testing.T) {
	setupTest()

	config := DriverCacheConfig{
		Platform: "", // Empty platform should default to android in createXTDriverWithConfig
		Serial:   "", // Empty serial will be auto-detected by NewAndroidDevice
	}

	driver, err := GetOrCreateXTDriver(config)
	// Device creation may succeed or fail depending on test environment
	if err != nil {
		// If device creation fails (no devices or multiple devices)
		assert.Nil(t, driver)
		assert.Contains(t, err.Error(), "failed to create XTDriver")
	} else {
		// If device creation succeeds (exactly one device connected)
		assert.NotNil(t, driver)
		// Verify that a driver was created and cached with actual serial
		drivers := ListCachedDrivers()
		assert.Len(t, drivers, 1)
		assert.NotEmpty(t, drivers[0].Serial) // Serial should be populated with actual device serial
	}
}

func TestGetOrCreateXTDriver_WithUnifiedDeviceOptions(t *testing.T) {
	setupTest()

	// Test creating driver config with unified DeviceOptions
	deviceOpts := option.NewDeviceOptions(
		option.WithPlatform("android"),
		option.WithDeviceSerialNumber("test_device_001"),
		option.WithDeviceUIA2(true),
	)

	config := DriverCacheConfig{
		Platform:   deviceOpts.Platform,
		Serial:     deviceOpts.GetSerial(),
		DeviceOpts: deviceOpts,
		AIOptions: []option.AIServiceOption{
			option.WithCVService(option.CVServiceTypeVEDEM),
		},
	}

	// Verify config is properly constructed
	assert.Equal(t, "android", config.Platform)
	assert.Equal(t, "test_device_001", config.Serial)
	assert.NotNil(t, config.DeviceOpts)
	assert.Equal(t, "android", config.DeviceOpts.Platform)
	assert.Equal(t, "test_device_001", config.DeviceOpts.GetSerial())
}

func TestGetOrCreateXTDriver_DifferentPlatformConfigs(t *testing.T) {
	setupTest()

	// Test Android config
	androidOpts := option.NewDeviceOptions(
		option.WithDeviceSerialNumber("android_001"),
		option.WithDeviceUIA2(true),
	)
	androidConfig := DriverCacheConfig{
		Platform:   "android",
		Serial:     "android_001",
		DeviceOpts: androidOpts,
	}
	assert.Equal(t, "android", androidConfig.DeviceOpts.Platform)

	// Test iOS config
	iosOpts := option.NewDeviceOptions(
		option.WithDeviceUDID("ios_001"),
		option.WithDeviceWDAPort(8100),
	)
	iosConfig := DriverCacheConfig{
		Platform:   "ios",
		Serial:     "ios_001",
		DeviceOpts: iosOpts,
	}
	assert.Equal(t, "ios", iosConfig.DeviceOpts.Platform)

	// Test Harmony config
	harmonyOpts := option.NewDeviceOptions(
		option.WithDeviceConnectKey("harmony_001"),
	)
	harmonyConfig := DriverCacheConfig{
		Platform:   "harmony",
		Serial:     "harmony_001",
		DeviceOpts: harmonyOpts,
	}
	assert.Equal(t, "harmony", harmonyConfig.DeviceOpts.Platform)

	// Test Browser config
	browserOpts := option.NewDeviceOptions(
		option.WithDeviceBrowserID("browser_001"),
		option.WithDeviceBrowserPageSize(1920, 1080),
	)
	browserConfig := DriverCacheConfig{
		Platform:   "browser",
		Serial:     "browser_001",
		DeviceOpts: browserOpts,
	}
	assert.Equal(t, "browser", browserConfig.DeviceOpts.Platform)
}

func TestRegisterXTDriver_EmptySerial(t *testing.T) {
	setupTest()

	err := RegisterXTDriver("", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serial cannot be empty")
}

func TestRegisterXTDriver_NilDriver(t *testing.T) {
	setupTest()

	err := RegisterXTDriver("test_serial", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver cannot be nil")
}

func TestRegisterXTDriver_Success(t *testing.T) {
	setupTest()

	// Create a minimal XTDriver for testing
	xtDriver := &XTDriver{}

	// Register external driver
	err := RegisterXTDriver("external_001", xtDriver)
	require.NoError(t, err)

	// Verify driver is cached
	drivers := ListCachedDrivers()
	assert.Len(t, drivers, 1)
	assert.Equal(t, "external_001", drivers[0].Serial)
	assert.Equal(t, int32(1), drivers[0].RefCount)
	assert.Equal(t, xtDriver, drivers[0].Driver)
}

func TestReleaseXTDriver_NonExistentSerial(t *testing.T) {
	setupTest()

	// Release non-existent driver should not error
	err := ReleaseXTDriver("non_existent")
	assert.NoError(t, err)
}

func TestReleaseXTDriver_CleanupWhenZero(t *testing.T) {
	setupTest()

	// Register driver
	xtDriver := &XTDriver{}
	err := RegisterXTDriver("cleanup_test", xtDriver)
	require.NoError(t, err)

	// Verify driver is cached
	drivers := ListCachedDrivers()
	assert.Len(t, drivers, 1)

	// Release driver (ref count goes to 0)
	err = ReleaseXTDriver("cleanup_test")
	require.NoError(t, err)

	// Verify driver is removed from cache
	drivers = ListCachedDrivers()
	assert.Len(t, drivers, 0)
}

func TestCleanupAllDrivers(t *testing.T) {
	setupTest()

	// Create multiple drivers
	xtDriver1 := &XTDriver{}
	xtDriver2 := &XTDriver{}
	xtDriver3 := &XTDriver{}

	err := RegisterXTDriver("cleanup_all_1", xtDriver1)
	require.NoError(t, err)
	err = RegisterXTDriver("cleanup_all_2", xtDriver2)
	require.NoError(t, err)
	err = RegisterXTDriver("cleanup_all_3", xtDriver3)
	require.NoError(t, err)

	// Verify all drivers are cached
	drivers := ListCachedDrivers()
	assert.Len(t, drivers, 3)

	// Cleanup all drivers
	CleanupAllDrivers()

	// Verify cache is empty
	drivers = ListCachedDrivers()
	assert.Len(t, drivers, 0)
}

func TestListCachedDrivers_Empty(t *testing.T) {
	setupTest()

	drivers := ListCachedDrivers()
	assert.Len(t, drivers, 0)
}

func TestListCachedDrivers_Multiple(t *testing.T) {
	setupTest()

	// Register multiple drivers
	xtDriver1 := &XTDriver{}
	xtDriver2 := &XTDriver{}

	err := RegisterXTDriver("list_test_1", xtDriver1)
	require.NoError(t, err)
	err = RegisterXTDriver("list_test_2", xtDriver2)
	require.NoError(t, err)

	// List drivers
	drivers := ListCachedDrivers()
	assert.Len(t, drivers, 2)

	// Verify driver information
	serials := make(map[string]bool)
	for _, cached := range drivers {
		serials[cached.Serial] = true
		assert.Equal(t, int32(1), cached.RefCount)
		assert.NotNil(t, cached.Driver)
	}
	assert.True(t, serials["list_test_1"])
	assert.True(t, serials["list_test_2"])
}

func TestDriverCacheConfig_WithoutDeviceOpts(t *testing.T) {
	setupTest()

	// Test creating config without DeviceOpts
	config := DriverCacheConfig{
		Platform: "android",
		Serial:   "default_test",
		// DeviceOpts is nil
	}

	// Verify config structure
	assert.Equal(t, "android", config.Platform)
	assert.Equal(t, "default_test", config.Serial)
	assert.Nil(t, config.DeviceOpts)
}

func TestDriverCacheConfig_DefaultAIOptions(t *testing.T) {
	setupTest()

	deviceOpts := option.NewDeviceOptions(
		option.WithPlatform("android"),
		option.WithDeviceSerialNumber("ai_test"),
	)

	config := DriverCacheConfig{
		Platform:   deviceOpts.Platform,
		Serial:     deviceOpts.GetSerial(),
		DeviceOpts: deviceOpts,
		// AIOptions is empty, should use default
	}

	// Verify config structure
	assert.Equal(t, "android", config.Platform)
	assert.Equal(t, "ai_test", config.Serial)
	assert.NotNil(t, config.DeviceOpts)
	assert.Len(t, config.AIOptions, 0) // Empty AI options
}

func TestConcurrentAccess(t *testing.T) {
	setupTest()

	// Test concurrent access to cache with GetOrCreateXTDriver
	const numGoroutines = 10
	const serial = "concurrent_test"

	deviceOpts := option.NewDeviceOptions(
		option.WithPlatform("android"),
		option.WithDeviceSerialNumber(serial),
	)
	config := DriverCacheConfig{
		Platform:   deviceOpts.Platform,
		Serial:     deviceOpts.GetSerial(),
		DeviceOpts: deviceOpts,
	}

	// Create drivers concurrently - this tests the cache's ability to handle concurrent access
	results := make(chan *XTDriver, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			driver, err := GetOrCreateXTDriver(config)
			results <- driver
			errors <- err
		}(i)
	}

	// Collect results
	var drivers []*XTDriver
	var errorCount int
	for i := 0; i < numGoroutines; i++ {
		driver := <-results
		err := <-errors
		if err != nil {
			errorCount++
		} else {
			drivers = append(drivers, driver)
		}
	}

	// All operations should succeed (or all fail if device creation fails)
	if errorCount == 0 {
		// If device creation succeeds, all drivers should be the same instance
		assert.Len(t, drivers, numGoroutines)
		firstDriver := drivers[0]
		for _, driver := range drivers[1:] {
			assert.Equal(t, firstDriver, driver)
		}

		// Verify ref count
		cachedDrivers := ListCachedDrivers()
		assert.Len(t, cachedDrivers, 1)
		assert.Equal(t, int32(numGoroutines), cachedDrivers[0].RefCount)
	} else {
		// If device creation fails (expected in test environment), all should fail
		assert.Equal(t, numGoroutines, errorCount)
		assert.Len(t, drivers, 0)
	}
}

func TestIntegrationExample_BasicUsage(t *testing.T) {
	setupTest()

	// Example 1: Basic external driver registration using unified DeviceOptions
	deviceOpts := option.NewDeviceOptions(
		option.WithPlatform("android"),
		option.WithDeviceSerialNumber("integration_001"),
		option.WithDeviceUIA2(true),
	)

	config := DriverCacheConfig{
		Platform:   deviceOpts.Platform,
		Serial:     deviceOpts.GetSerial(),
		DeviceOpts: deviceOpts,
		AIOptions: []option.AIServiceOption{
			option.WithCVService(option.CVServiceTypeVEDEM),
		},
	}

	// Verify config is properly constructed
	assert.Equal(t, "android", config.Platform)
	assert.Equal(t, "integration_001", config.Serial)
	assert.NotNil(t, config.DeviceOpts)
	assert.Len(t, config.AIOptions, 1)
}

func TestIntegrationExample_TraditionalWay(t *testing.T) {
	setupTest()

	// Example 1b: Traditional way (still supported)
	xtDriver := &XTDriver{}

	// Register using cache API directly
	err := RegisterXTDriver("integration_002", xtDriver)
	require.NoError(t, err)

	// Verify registration
	drivers := ListCachedDrivers()
	assert.Len(t, drivers, 1)
	assert.Equal(t, "integration_002", drivers[0].Serial)

	// Clean up
	err = ReleaseXTDriver("integration_002")
	require.NoError(t, err)
}

func TestIntegrationExample_MultipleDevices(t *testing.T) {
	setupTest()

	// Test multiple devices like in external_driver_example.go
	devices := []struct {
		platform string
		serial   string
		opts     *option.DeviceOptions
	}{
		{
			platform: "android",
			serial:   "multi_android_001",
			opts: option.NewDeviceOptions(
				option.WithDeviceSerialNumber("multi_android_001"),
				option.WithDeviceUIA2(true),
			),
		},
		{
			platform: "ios",
			serial:   "multi_ios_001",
			opts: option.NewDeviceOptions(
				option.WithDeviceUDID("multi_ios_001"),
				option.WithDeviceWDAPort(8100),
			),
		},
		{
			platform: "harmony",
			serial:   "multi_harmony_001",
			opts: option.NewDeviceOptions(
				option.WithDeviceConnectKey("multi_harmony_001"),
			),
		},
		{
			platform: "browser",
			serial:   "multi_browser_001",
			opts: option.NewDeviceOptions(
				option.WithDeviceBrowserID("multi_browser_001"),
				option.WithDeviceBrowserPageSize(1920, 1080),
			),
		},
	}

	// Create configs for all devices
	var configs []DriverCacheConfig
	for _, device := range devices {
		config := DriverCacheConfig{
			Platform:   device.platform,
			Serial:     device.serial,
			DeviceOpts: device.opts,
		}
		configs = append(configs, config)
	}

	// Verify all configs are properly constructed
	assert.Len(t, configs, len(devices))

	// Verify each device config
	for i, config := range configs {
		assert.Equal(t, devices[i].platform, config.Platform)
		assert.Equal(t, devices[i].serial, config.Serial)
		assert.NotNil(t, config.DeviceOpts)
		assert.Equal(t, devices[i].platform, config.DeviceOpts.Platform)
	}
}

func TestDeviceOptionsIntegration(t *testing.T) {
	setupTest()

	// Test unified DeviceOptions with different platforms
	testCases := []struct {
		name     string
		platform string
		opts     []option.DeviceOption
		expected string
	}{
		{
			name:     "Android with auto-detection",
			platform: "",
			opts: []option.DeviceOption{
				option.WithDeviceSerialNumber("android_auto"),
				option.WithDeviceUIA2(true),
			},
			expected: "android",
		},
		{
			name:     "iOS with auto-detection",
			platform: "",
			opts: []option.DeviceOption{
				option.WithDeviceUDID("ios_auto"),
				option.WithDeviceWDAPort(8100),
			},
			expected: "ios",
		},
		{
			name:     "Harmony with auto-detection",
			platform: "",
			opts: []option.DeviceOption{
				option.WithDeviceConnectKey("harmony_auto"),
			},
			expected: "harmony",
		},
		{
			name:     "Browser with auto-detection",
			platform: "",
			opts: []option.DeviceOption{
				option.WithDeviceBrowserID("browser_auto"),
				option.WithDeviceBrowserPageSize(1920, 1080),
			},
			expected: "browser",
		},
		{
			name:     "Explicit platform setting",
			platform: "android",
			opts: []option.DeviceOption{
				option.WithPlatform("android"),
				option.WithDeviceSerialNumber("explicit_android"),
			},
			expected: "android",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deviceOpts := option.NewDeviceOptions(tc.opts...)
			assert.Equal(t, tc.expected, deviceOpts.Platform)
			assert.NotEmpty(t, deviceOpts.GetSerial())
		})
	}
}

func TestCacheReferenceCountManagement(t *testing.T) {
	setupTest()

	// Test reference count increment and decrement
	xtDriver := &XTDriver{}
	serial := "ref_count_test"

	// Register driver
	err := RegisterXTDriver(serial, xtDriver)
	require.NoError(t, err)

	// Verify initial ref count
	drivers := ListCachedDrivers()
	assert.Len(t, drivers, 1)
	assert.Equal(t, int32(1), drivers[0].RefCount)

	// Simulate multiple references by manually incrementing
	if cachedItem, ok := driverCache.Load(serial); ok {
		if cached, ok := cachedItem.(*CachedXTDriver); ok {
			cached.RefCount++
		}
	}

	// Verify ref count increased
	drivers = ListCachedDrivers()
	assert.Len(t, drivers, 1)
	assert.Equal(t, int32(2), drivers[0].RefCount)

	// Release once
	err = ReleaseXTDriver(serial)
	require.NoError(t, err)

	// Verify ref count decreased but driver still cached
	drivers = ListCachedDrivers()
	assert.Len(t, drivers, 1)
	assert.Equal(t, int32(1), drivers[0].RefCount)

	// Release again
	err = ReleaseXTDriver(serial)
	require.NoError(t, err)

	// Verify driver removed from cache
	drivers = ListCachedDrivers()
	assert.Len(t, drivers, 0)
}
