package uixt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func NewXTDriver(driver IDriver, opts ...option.AIServiceOption) (*XTDriver, error) {
	driverExt := &XTDriver{
		IDriver: driver,
		client: &MCPClient4XTDriver{
			Server: NewMCPServer(),
		},
	}

	services := option.NewAIServiceOptions(opts...)

	var err error
	if services.CVService != "" {
		driverExt.CVService, err = ai.NewCVService(services.CVService)
		if err != nil {
			log.Error().Err(err).Msg("init vedem image service failed")
			return nil, err
		}
	}
	if services.LLMService != "" {
		driverExt.LLMService, err = ai.NewLLMService(services.LLMService)
		if err != nil {
			log.Error().Err(err).Msg("init llm service failed")
			return nil, err
		}
	}

	return driverExt, nil
}

// XTDriver = IDriver + AI
type XTDriver struct {
	IDriver
	CVService  ai.ICVService  // OCR/CV
	LLMService ai.ILLMService // LLM

	client *MCPClient4XTDriver // MCP Client
}

// MCPClient4XTDriver is a mock MCP client that only implements the methods used by the host
type MCPClient4XTDriver struct {
	client.MCPClient
	Server *MCPServer4XTDriver
}

func (c *MCPClient4XTDriver) ListTools(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	tools := c.Server.ListTools()
	return &mcp.ListToolsResult{Tools: tools}, nil
}

func (c *MCPClient4XTDriver) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	handler := c.Server.GetHandler(req.Params.Name)
	if handler == nil {
		return mcp.NewToolResultError(fmt.Sprintf("handler for tool %s not found", req.Params.Name)), nil
	}
	return handler(ctx, req)
}

func (c *MCPClient4XTDriver) Initialize(ctx context.Context, req mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	// no need to initialize for local server
	return &mcp.InitializeResult{}, nil
}

func (c *MCPClient4XTDriver) Close() error {
	// no need to close for local server
	return nil
}

func (dExt *XTDriver) ExecuteAction(action MobileAction) (err error) {
	// Convert action to MCP tool call
	req, err := convertActionToCallToolRequest(action)
	if err != nil {
		return fmt.Errorf("failed to convert action to MCP tool call: %w", err)
	}

	// Execute via MCP tool
	result, err := dExt.client.CallTool(context.Background(), req)
	if err != nil {
		return fmt.Errorf("MCP tool call failed: %w", err)
	}

	// Check if the tool execution had business logic errors
	if result.IsError {
		if len(result.Content) > 0 {
			return fmt.Errorf("tool execution failed: %s", result.Content[0])
		}
		return fmt.Errorf("tool execution failed")
	}

	log.Debug().Str("method", string(action.Method)).Msg("executed action via MCP tool")
	return nil
}

func convertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	var arguments map[string]interface{}

	switch action.Method {
	case option.ACTION_WebLoginNoneUI:
		if params, ok := action.Params.([]interface{}); ok && len(params) == 4 {
			arguments = map[string]interface{}{
				"packageName": params[0].(string),
				"phoneNumber": params[1].(string),
				"captcha":     params[2].(string),
				"password":    params[3].(string),
			}
		} else if params, ok := action.Params.([]string); ok && len(params) == 4 {
			arguments = map[string]interface{}{
				"packageName": params[0],
				"phoneNumber": params[1],
				"captcha":     params[2],
				"password":    params[3],
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid web login params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "web_login_none_ui",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_AppInstall:
		if app, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"appUrl": app,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid app install params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "app_install",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_AppUninstall:
		if packageName, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"packageName": packageName,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid app uninstall params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "app_uninstall",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_AppClear:
		if packageName, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"packageName": packageName,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid app clear params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "app_clear",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_AppLaunch:
		if packageName, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"packageName": packageName,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid app launch params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "launch_app",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_SwipeToTapApp:
		if appName, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"appName": appName,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap app params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "swipe_to_tap_app",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_SwipeToTapText:
		if text, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"text": text,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap text params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "swipe_to_tap_text",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_SwipeToTapTexts:
		var texts []string
		if textsSlice, ok := action.Params.([]string); ok {
			texts = textsSlice
		} else if textsInterface, err := builtin.ConvertToStringSlice(action.Params); err == nil {
			texts = textsInterface
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap texts params: %v", action.Params)
		}
		arguments = map[string]interface{}{
			"texts": texts,
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "swipe_to_tap_texts",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_AppTerminate:
		if packageName, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"packageName": packageName,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid app terminate params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "terminate_app",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_Home:
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "home",
				Arguments: map[string]interface{}{},
			},
		}, nil

	case option.ACTION_SecondaryClick:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
			arguments = map[string]interface{}{
				"x": params[0],
				"y": params[1],
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid secondary click params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "secondary_click",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_HoverBySelector:
		if selector, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"selector": selector,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid hover by selector params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "hover_by_selector",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_TapBySelector:
		if selector, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"selector": selector,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid tap by selector params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "tap_by_selector",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_SecondaryClickBySelector:
		if selector, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"selector": selector,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid secondary click by selector params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "secondary_click_by_selector",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_WebCloseTab:
		var tabIndex int
		if param, ok := action.Params.(json.Number); ok {
			paramInt64, _ := param.Int64()
			tabIndex = int(paramInt64)
		} else if param, ok := action.Params.(int64); ok {
			tabIndex = int(param)
		} else if param, ok := action.Params.(int); ok {
			tabIndex = param
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid web close tab params: %v", action.Params)
		}
		arguments = map[string]interface{}{
			"tabIndex": tabIndex,
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "web_close_tab",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_SetIme:
		if ime, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"ime": ime,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid set ime params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "set_ime",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_GetSource:
		if packageName, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"packageName": packageName,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid get source params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "get_source",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_TapXY:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
			x, y := params[0], params[1]
			arguments = map[string]interface{}{
				"x": x,
				"y": y,
			}
			// Add duration if available from action options
			if actionOptions := action.GetOptions(); len(actionOptions) > 0 {
				for _, opt := range actionOptions {
					if opt != nil {
						// Add options like duration
						if duration := action.ActionOptions.Duration; duration > 0 {
							arguments["duration"] = duration
						}
					}
				}
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid tap params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "tap_xy",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_TapAbsXY:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
			x, y := params[0], params[1]
			arguments = map[string]interface{}{
				"x": x,
				"y": y,
			}
			// Add duration if available
			if duration := action.ActionOptions.Duration; duration > 0 {
				arguments["duration"] = duration
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid tap abs params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "tap_abs_xy",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_TapByOCR:
		if text, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"text": text,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid tap by OCR params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "tap_by_ocr",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_TapByCV:
		// For TapByCV, the original action might not have params but relies on options
		arguments = map[string]interface{}{
			"imagePath": "", // Will be handled by the tool based on UI types
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "tap_by_cv",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_DoubleTapXY:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
			x, y := params[0], params[1]
			arguments = map[string]interface{}{
				"x": x,
				"y": y,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid double tap params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "double_tap_xy",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_Swipe:
		// Handle different types of swipe params
		switch params := action.Params.(type) {
		case string:
			// Direction swipe like "up", "down", "left", "right"
			arguments = map[string]interface{}{
				"direction": params,
			}
			// Add duration and press duration from options
			if duration := action.ActionOptions.Duration; duration > 0 {
				arguments["duration"] = duration
			}
			if pressDuration := action.ActionOptions.PressDuration; pressDuration > 0 {
				arguments["pressDuration"] = pressDuration
			}
			return mcp.CallToolRequest{
				Params: struct {
					Name      string         `json:"name"`
					Arguments map[string]any `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name:      "swipe",
					Arguments: arguments,
				},
			}, nil
		default:
			// Advanced swipe with coordinates
			if paramSlice, err := builtin.ConvertToFloat64Slice(params); err == nil && len(paramSlice) == 4 {
				arguments = map[string]interface{}{
					"fromX": paramSlice[0],
					"fromY": paramSlice[1],
					"toX":   paramSlice[2],
					"toY":   paramSlice[3],
				}
				// Add duration and press duration from options
				if duration := action.ActionOptions.Duration; duration > 0 {
					arguments["duration"] = duration
				}
				if pressDuration := action.ActionOptions.PressDuration; pressDuration > 0 {
					arguments["pressDuration"] = pressDuration
				}
				return mcp.CallToolRequest{
					Params: struct {
						Name      string         `json:"name"`
						Arguments map[string]any `json:"arguments,omitempty"`
						Meta      *struct {
							ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
						} `json:"_meta,omitempty"`
					}{
						Name:      "swipe_advanced",
						Arguments: arguments,
					},
				}, nil
			}
		}
		return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe params: %v", action.Params)

	case option.ACTION_Input:
		text := fmt.Sprintf("%v", action.Params)
		arguments = map[string]interface{}{
			"text": text,
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "input",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_Back:
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "back",
				Arguments: map[string]interface{}{},
			},
		}, nil

	case option.ACTION_Sleep:
		arguments = map[string]interface{}{
			"seconds": action.Params,
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "sleep",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_SleepMS:
		var milliseconds int64
		if param, ok := action.Params.(json.Number); ok {
			milliseconds, _ = param.Int64()
		} else if param, ok := action.Params.(int64); ok {
			milliseconds = param
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid sleep ms params: %v", action.Params)
		}
		arguments = map[string]interface{}{
			"milliseconds": milliseconds,
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "sleep_ms",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_SleepRandom:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			arguments = map[string]interface{}{
				"params": params,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid sleep random params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "sleep_random",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_ScreenShot:
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "screenshot",
				Arguments: map[string]interface{}{},
			},
		}, nil

	case option.ACTION_ClosePopups:
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "close_popups",
				Arguments: map[string]interface{}{},
			},
		}, nil

	case option.ACTION_CallFunction:
		if description, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"description": description,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid call function params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "call_function",
				Arguments: arguments,
			},
		}, nil

	case option.ACTION_AIAction:
		if prompt, ok := action.Params.(string); ok {
			arguments = map[string]interface{}{
				"prompt": prompt,
			}
		} else {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid AI action params: %v", action.Params)
		}
		return mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "ai_action",
				Arguments: arguments,
			},
		}, nil

	default:
		return mcp.CallToolRequest{}, fmt.Errorf("unsupported action method: %s", action.Method)
	}
}
