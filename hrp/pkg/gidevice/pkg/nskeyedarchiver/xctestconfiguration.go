package nskeyedarchiver

import (
	"reflect"

	"howett.net/plist"
)

type XCTestConfiguration struct {
	internal map[string]interface{}
}

func newXCTestConfiguration(cfg interface{}) *XCTestConfiguration {
	return cfg.(*XCTestConfiguration)
}

func NewXCTestConfiguration(nsuuid *NSUUID, nsurl *NSURL, targetBundleID, targetAppPath string) *XCTestConfiguration {
	contents := map[string]interface{}{
		"aggregateStatisticsBeforeCrash": map[string]interface{}{
			"XCSuiteRecordsKey": map[string]interface{}{},
		},
		"automationFrameworkPath":           "/Developer/Library/PrivateFrameworks/XCTAutomationSupport.framework",
		"baselineFileRelativePath":          nil,
		"baselineFileURL":                   nil,
		"defaultTestExecutionTimeAllowance": nil,
		"disablePerformanceMetrics":         false,
		"emitOSLogs":                        false,
		"formatVersion":                     2,
		"gatherLocalizableStringsData":      false,
		"initializeForUITesting":            true,
		"maximumTestExecutionTimeAllowance": nil,
		"productModuleName":                 "WebDriverAgentRunner", // set to other value is also OK
		"randomExecutionOrderingSeed":       nil,
		"reportActivities":                  true,
		"reportResultsToIDE":                true,
		"systemAttachmentLifetime":          2,
		"targetApplicationArguments":        []interface{}{}, // maybe useless
		"targetApplicationEnvironment":      nil,
		"targetApplicationPath":             targetAppPath,
		"testApplicationDependencies":       map[string]interface{}{},
		"testApplicationUserOverrides":      nil,
		"testBundleRelativePath":            nil,
		"testExecutionOrdering":             0,
		"testTimeoutsEnabled":               false,
		"testsDrivenByIDE":                  false,
		"testsMustRunOnMainThread":          true,
		"testsToRun":                        nil,
		"testsToSkip":                       nil,
		"treatMissingBaselinesAsFailures":   false,
		"userAttachmentLifetime":            1,
		"testBundleURL":                     nsurl,
		"sessionIdentifier":                 nsuuid,
		"targetApplicationBundleID":         targetBundleID,
		// "targetApplicationBundleID":         "",
	}
	return &XCTestConfiguration{internal: contents}
}

func (cfg *XCTestConfiguration) archive(objects []interface{}) []interface{} {
	info := map[string]interface{}{}
	objects = append(objects, info)

	info["$class"] = plist.UID(len(objects))

	cls := map[string]interface{}{
		"$classname": "XCTestConfiguration",
		"$classes":   []interface{}{"XCTestConfiguration", "NSObject"},
	}
	objects = append(objects, cls)

	for k, v := range cfg.internal {
		val := reflect.ValueOf(v)
		if !val.IsValid() {
			info[k] = plist.UID(0)
			continue
		}

		typ := val.Type()

		if k != "formatVersion" && (typ.Kind() == reflect.Bool || typ.Kind() == reflect.Uintptr || typ.Kind() == reflect.Int) {
			info[k] = v
		} else {
			var uid plist.UID
			objects, uid = archive(objects, v)
			info[k] = uid
		}
	}
	return objects
}
