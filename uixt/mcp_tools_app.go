package uixt

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt/option"
)

// ToolListPackages implements the list_packages tool call.
type ToolListPackages struct {
	// Return data fields - these define the structure of data returned by this tool
	Packages []string `json:"packages" desc:"List of installed app package names on the device"`
	Count    int      `json:"count" desc:"Number of installed packages"`
}

func (t *ToolListPackages) Name() option.ActionName {
	return option.ACTION_ListPackages
}

func (t *ToolListPackages) Description() string {
	return "List all installed apps/packages on the device with their package names."
}

func (t *ToolListPackages) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_ListPackages)
}

func (t *ToolListPackages) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.GetArguments())
		if err != nil {
			return nil, err
		}

		apps, err := driverExt.IDriver.GetDevice().ListPackages()
		if err != nil {
			return NewMCPErrorResponse("Failed to list packages: " + err.Error()), nil
		}

		message := fmt.Sprintf("Found %d installed packages", len(apps))
		returnData := ToolListPackages{
			Packages: apps,
			Count:    len(apps),
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolListPackages) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return BuildMCPCallToolRequest(t.Name(), map[string]any{}, action), nil
}

// ToolLaunchApp implements the launch_app tool call.
type ToolLaunchApp struct {
	// Return data fields - these define the structure of data returned by this tool
	PackageName string `json:"packageName" desc:"Package name of the launched app"`
}

func (t *ToolLaunchApp) Name() option.ActionName {
	return option.ACTION_AppLaunch
}

func (t *ToolLaunchApp) Description() string {
	return "Launch an app on mobile device using its package name. Use list_packages tool first to find the correct package name."
}

func (t *ToolLaunchApp) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AppLaunch)
}

func (t *ToolLaunchApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		if unifiedReq.PackageName == "" {
			return nil, fmt.Errorf("package_name is required")
		}

		// Launch app action logic
		err = driverExt.AppLaunch(unifiedReq.PackageName)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Launch app failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully launched app: %s", unifiedReq.PackageName)
		returnData := ToolLaunchApp{PackageName: unifiedReq.PackageName}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolLaunchApp) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid app launch params: %v", action.Params)
}

// ToolTerminateApp implements the terminate_app tool call.
type ToolTerminateApp struct {
	// Return data fields - these define the structure of data returned by this tool
	PackageName string `json:"packageName" desc:"Package name of the terminated app"`
	WasRunning  bool   `json:"wasRunning" desc:"Whether the app was actually running before termination"`
}

func (t *ToolTerminateApp) Name() option.ActionName {
	return option.ACTION_AppTerminate
}

func (t *ToolTerminateApp) Description() string {
	return "Stop and terminate a running app on mobile device using its package name"
}

func (t *ToolTerminateApp) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AppTerminate)
}

func (t *ToolTerminateApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		if unifiedReq.PackageName == "" {
			return nil, fmt.Errorf("package_name is required")
		}

		// Terminate app action logic
		success, err := driverExt.AppTerminate(unifiedReq.PackageName)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Terminate app failed: %s", err.Error())), err
		}
		if !success {
			log.Warn().Str("packageName", unifiedReq.PackageName).Msg("app was not running")
		}

		message := fmt.Sprintf("Successfully terminated app: %s", unifiedReq.PackageName)
		returnData := ToolTerminateApp{
			PackageName: unifiedReq.PackageName,
			WasRunning:  success,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolTerminateApp) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid app terminate params: %v", action.Params)
}

// ToolAppInstall implements the app_install tool call.
type ToolAppInstall struct {
	// Return data fields - these define the structure of data returned by this tool
	Path string `json:"path" desc:"Path or URL of the installed app"`
}

func (t *ToolAppInstall) Name() option.ActionName {
	return option.ACTION_AppInstall
}

func (t *ToolAppInstall) Description() string {
	return "Install an app on the device from a URL or local file path"
}

func (t *ToolAppInstall) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AppInstall)
}

func (t *ToolAppInstall) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// App install action logic
		err = driverExt.GetDevice().Install(unifiedReq.AppUrl)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("App install failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully installed app from: %s", unifiedReq.AppUrl)
		returnData := ToolAppInstall{Path: unifiedReq.AppUrl}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolAppInstall) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if appUrl, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"appUrl": appUrl,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid app install params: %v", action.Params)
}

// ToolAppUninstall implements the app_uninstall tool call.
type ToolAppUninstall struct {
	// Return data fields - these define the structure of data returned by this tool
	PackageName string `json:"packageName" desc:"Package name of the uninstalled app"`
}

func (t *ToolAppUninstall) Name() option.ActionName {
	return option.ACTION_AppUninstall
}

func (t *ToolAppUninstall) Description() string {
	return "Uninstall an app from the device"
}

func (t *ToolAppUninstall) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AppUninstall)
}

func (t *ToolAppUninstall) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// App uninstall action logic
		err = driverExt.GetDevice().Uninstall(unifiedReq.PackageName)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("App uninstall failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully uninstalled app: %s", unifiedReq.PackageName)
		returnData := ToolAppUninstall{PackageName: unifiedReq.PackageName}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolAppUninstall) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid app uninstall params: %v", action.Params)
}

// ToolAppClear implements the app_clear tool call.
type ToolAppClear struct {
	// Return data fields - these define the structure of data returned by this tool
	PackageName string `json:"packageName" desc:"Package name of the app whose data was cleared"`
}

func (t *ToolAppClear) Name() option.ActionName {
	return option.ACTION_AppClear
}

func (t *ToolAppClear) Description() string {
	return "Clear app data and cache for a specific app using its package name"
}

func (t *ToolAppClear) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AppClear)
}

func (t *ToolAppClear) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// App clear action logic
		err = driverExt.AppClear(unifiedReq.PackageName)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("App clear failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully cleared app: %s", unifiedReq.PackageName)
		returnData := ToolAppClear{PackageName: unifiedReq.PackageName}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolAppClear) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid app clear params: %v", action.Params)
}

// ToolGetForegroundApp implements the get_foreground_app tool call.
type ToolGetForegroundApp struct {
	// Return data fields - these define the structure of data returned by this tool
	PackageName string `json:"packageName" desc:"Package name of the foreground app"`
	AppName     string `json:"appName" desc:"Name of the foreground app"`
}

func (t *ToolGetForegroundApp) Name() option.ActionName {
	return option.ACTION_GetForegroundApp
}

func (t *ToolGetForegroundApp) Description() string {
	return "Get information about the currently running foreground app"
}

func (t *ToolGetForegroundApp) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_GetForegroundApp)
}

func (t *ToolGetForegroundApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.GetArguments())
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		// Get foreground app info
		appInfo, err := driverExt.ForegroundInfo()
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Get foreground app failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Current foreground app: %s (%s)", appInfo.AppName, appInfo.PackageName)
		returnData := ToolGetForegroundApp{
			PackageName: appInfo.PackageName,
			AppName:     appInfo.AppName,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolGetForegroundApp) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return BuildMCPCallToolRequest(t.Name(), map[string]any{}, action), nil
}

// ToolOpenApp implements the open_app tool call.
type ToolOpenApp struct {
	// Return data fields - these define the structure of data returned by this tool
	PackageName string `json:"packageName" desc:"Package name of the opened app"`
}

func (t *ToolOpenApp) Name() option.ActionName {
	return option.ACTION_OpenApp
}

func (t *ToolOpenApp) Description() string {
	return "Open an app on mobile device using its package name and wait for the app to load"
}

func (t *ToolOpenApp) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_OpenApp)
}

func (t *ToolOpenApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		if unifiedReq.PackageName == "" {
			return nil, fmt.Errorf("package_name is required")
		}

		// Open app action logic
		err = driverExt.AppLaunch(unifiedReq.PackageName)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Open app failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully opened app: %s", unifiedReq.PackageName)
		returnData := ToolOpenApp{PackageName: unifiedReq.PackageName}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolOpenApp) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid open app params: %v", action.Params)
}

// ToolTerminateAppNew implements the terminal_app tool call.
type ToolTerminateAppNew struct {
	// Return data fields - these define the structure of data returned by this tool
	PackageName string `json:"packageName" desc:"Package name of the terminated app"`
	WasRunning  bool   `json:"wasRunning" desc:"Whether the app was actually running before termination"`
}

func (t *ToolTerminateAppNew) Name() option.ActionName {
	return option.ACTION_TerminateApp
}

func (t *ToolTerminateAppNew) Description() string {
	return "Terminate a running app on mobile device using its package name"
}

func (t *ToolTerminateAppNew) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TerminateApp)
}

func (t *ToolTerminateAppNew) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		if unifiedReq.PackageName == "" {
			return nil, fmt.Errorf("package_name is required")
		}

		// Terminate app action logic
		success, err := driverExt.AppTerminate(unifiedReq.PackageName)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Terminate app failed: %s", err.Error())), err
		}
		if !success {
			log.Warn().Str("packageName", unifiedReq.PackageName).Msg("app was not running")
		}

		message := fmt.Sprintf("Successfully terminated app: %s", unifiedReq.PackageName)
		returnData := ToolTerminateAppNew{
			PackageName: unifiedReq.PackageName,
			WasRunning:  success,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolTerminateAppNew) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid terminate app params: %v", action.Params)
}

// ToolColdLaunch implements the cold_launch tool call.
type ToolColdLaunch struct {
	// Return data fields - these define the structure of data returned by this tool
	PackageName string `json:"packageName" desc:"Package name of the cold launched app"`
}

func (t *ToolColdLaunch) Name() option.ActionName {
	return option.ACTION_ColdLaunch
}

func (t *ToolColdLaunch) Description() string {
	return "Perform a cold launch of an app (terminate first if running, then launch)"
}

func (t *ToolColdLaunch) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_ColdLaunch)
}

func (t *ToolColdLaunch) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		if unifiedReq.PackageName == "" {
			return nil, fmt.Errorf("package_name is required")
		}

		// Cold launch logic: terminate first, then launch
		// First try to terminate the app (ignore errors if app is not running)
		_, err = driverExt.AppTerminate(unifiedReq.PackageName)
		if err != nil {
			log.Warn().Str("packageName", unifiedReq.PackageName).Msg("app was not running")
			return NewMCPErrorResponse(fmt.Sprintf("Cold launch failed, terminate app failed: %s", err.Error())), err
		}
		time.Sleep(3 * time.Second)
		// Then launch the app
		err = driverExt.AppLaunch(unifiedReq.PackageName)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Cold launch failed, launch app failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully cold launched app: %s", unifiedReq.PackageName)
		returnData := ToolColdLaunch{PackageName: unifiedReq.PackageName}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolColdLaunch) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid cold launch params: %v", action.Params)
}
