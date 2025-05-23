package mcphost

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

// NewChat creates a new chat session
func (h *MCPHost) NewChat(ctx context.Context) (*Chat, error) {
	// Get model config from environment variables
	modelConfig, err := ai.GetModelConfig(option.LLMServiceTypeUITARS)
	if err != nil {
		return nil, err
	}
	planner, err := ai.NewPlanner(ctx, modelConfig)
	if err != nil {
		return nil, err
	}

	// Convert MCP tools to eino tool infos
	einoTools, err := h.GetEinoToolInfos(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get eino tool infos")
	}
	if err := planner.RegisterTools(einoTools); err != nil {
		return nil, err
	}

	// Create markdown renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(styles.TokyoNightStyle),
		glamour.WithWordWrap(getTerminalWidth()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create markdown renderer")
	}

	return &Chat{
		planner:  planner,
		renderer: renderer,
		host:     h,
	}, nil
}

// Chat represents a chat session with LLM
type Chat struct {
	host     *MCPHost
	planner  *ai.Planner
	renderer *glamour.TermRenderer
}

// Start starts the chat session
func (c *Chat) Start(ctx context.Context) error {
	c.showWelcome()

	for {
		var input string
		err := huh.NewForm(huh.NewGroup(huh.NewText().
			Title("Enter your prompt (Type /help for commands, Ctrl+C to quit)").
			Value(&input).
			CharLimit(5000)),
		).WithWidth(getTerminalWidth()).
			WithTheme(huh.ThemeCharm()).
			Run()
		if err != nil {
			// Check if it's a user abort (Ctrl+C)
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println("\nGoodbye!")
				return nil // Exit cleanly
			}
			return err // Return other errors normally
		}

		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			if err := c.handleCommand(input); err != nil {
				log.Error().Err(err).Msg("failed to handle command")
			}
			continue
		}

		// run prompt with MCP tools
		if err := c.runPrompt(ctx, input); err != nil {
			log.Error().Err(err).Msg("run prompt error")
		}
	}
}

// runPrompt run prompt with MCP tools
func (c *Chat) runPrompt(ctx context.Context, prompt string) error {
	fmt.Printf("\n%s\n", promptStyle.Render("You: "+prompt))

	// Create user message
	planningOpts := &ai.PlanningOptions{
		UserInstruction: prompt,
		Message: &schema.Message{
			Role:    schema.User,
			Content: prompt,
		},
	}

	// Call planner to get response
	var result *ai.PlanningResult
	var err error
	_ = spinner.New().Title("Thinking...").Action(func() {
		result, err = c.planner.Call(ctx, planningOpts)
	}).Run()
	if err != nil {
		return err
	}

	// Handle tool calls
	toolCalls := result.ToolCalls
	if len(toolCalls) > 0 {
		return c.handleToolCalls(ctx, toolCalls)
	}

	c.renderContent("Assistant", result.ActionSummary)

	return nil
}

func (c *Chat) handleToolCalls(ctx context.Context, toolCalls []schema.ToolCall) error {
	for _, toolCall := range toolCalls {
		serverToolName := toolCall.Function.Name
		toolArgs := toolCall.Function.Arguments
		log.Debug().Str("name", serverToolName).Str("args", toolArgs).Msg("handle tool call")

		// Parse tool name
		parts := strings.SplitN(serverToolName, "__", 2)
		if len(parts) != 2 {
			log.Error().Str("name", serverToolName).Msg("invalid tool name")
			continue
		}
		serverName, toolName := parts[0], parts[1]

		// Unmarshal tool arguments from JSON string
		var argsMap map[string]any
		if err := sonic.UnmarshalString(toolArgs, &argsMap); err != nil {
			log.Error().Err(err).Str("args", toolArgs).Msg("failed to unmarshal tool arguments")
			continue
		}

		// Invoke tool
		result, err := c.host.InvokeTool(ctx, serverName, toolName, argsMap)
		if err != nil {
			log.Error().Err(err).Msg("invoke tool failed")
			toolMsg := &schema.Message{
				Role:       schema.Tool,
				Content:    fmt.Sprintf("invoke tool %s error: %v", serverToolName, err),
				ToolCallID: toolCall.ID,
			}
			c.planner.History().Append(toolMsg)
			continue
		}

		// Format tool result, append message to history
		renderStr := ""
		if result != nil && len(result.Content) > 0 {
			for _, item := range result.Content {
				if contentMap, ok := item.(mcp.TextContent); ok {
					renderStr += contentMap.Text + "\n"
					toolMsg := &schema.Message{
						Role:       schema.Tool,
						ToolCallID: toolCall.ID,
						Content:    contentMap.Text,
					}
					c.planner.History().Append(toolMsg)
				} else if contentMap, ok := item.(mcp.ImageContent); ok {
					renderStr += "<data:image/base64...>\n" // base64-encoded image data
					toolMsg := &schema.Message{
						Role:       schema.Tool,
						ToolCallID: toolCall.ID,
						MultiContent: []schema.ChatMessagePart{
							{
								Type: schema.ChatMessagePartTypeImageURL,
								ImageURL: &schema.ChatMessageImageURL{
									URL:      contentMap.Data,
									MIMEType: contentMap.MIMEType,
								},
							},
						},
					}
					c.planner.History().Append(toolMsg)
				}
			}
		} else {
			renderStr = fmt.Sprintf("%+v", result)
			toolMsg := &schema.Message{
				Role:       schema.Tool,
				ToolCallID: toolCall.ID,
				Content:    renderStr,
			}
			c.planner.History().Append(toolMsg)
		}
		c.renderContent("Tool Result", renderStr)
	}
	return nil
}

// handleCommand handles commands
func (c *Chat) handleCommand(cmd string) error {
	switch cmd {
	case "/help":
		c.showWelcome()
	case "/tools":
		c.showTools()
	case "/history":
		c.showHistory()
	case "/clear":
		c.planner.History().Clear()
	case "/quit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
	return nil
}

// showWelcome show welcome and help information
func (c *Chat) showWelcome() {
	markdown := fmt.Sprintf(`# Welcome to HttpRunner MCPHost Chat!

## Available Commands

The following commands are available:

- **/help**: Show this help message
- **/tools**: List all available tools
- **/history**: Display conversation history
- **/clear**: Clear conversation history
- **/quit**: Exit the chat session

You can also press Ctrl+C at any time to quit.

## Configurations

- **mcp-config**: %s
- **system-prompt**: %s
`, c.host.config.ConfigPath, c.planner.SystemPrompt())

	c.renderContent("", markdown)
}

func (c *Chat) showHistory() {
	if len(*c.planner.History()) <= 1 { // Only system message
		fmt.Println("No conversation history yet.")
		return
	}

	fmt.Println("\nConversation History:")
	for _, msg := range *c.planner.History() {
		if msg.Role == schema.System {
			continue
		}

		role := "You"
		if msg.Role == schema.Assistant {
			role = "Assistant"
		} else if msg.Role == schema.Tool {
			role = "Tool Result"
		}
		c.renderContent(role, msg.Content)
	}
}

func (c *Chat) showTools() {
	if c.host == nil {
		fmt.Println("No MCP host loaded.")
		return
	}
	ctx := context.Background()
	results := c.host.GetTools(ctx)
	if len(results) == 0 {
		fmt.Println("No MCP servers loaded.")
		return
	}
	width := getTerminalWidth()
	contentWidth := width - 12
	l := list.New().EnumeratorStyle(lipgloss.NewStyle().Foreground(tokyoPurple).MarginRight(1))
	for _, serverTools := range results {
		serverList := list.New().EnumeratorStyle(lipgloss.NewStyle().Foreground(tokyoCyan).MarginRight(1))
		if serverTools.Err != nil {
			serverList.Item(contentStyle.Render(fmt.Sprintf("Error: %v", serverTools.Err)))
		} else if len(serverTools.Tools) == 0 {
			serverList.Item(contentStyle.Render("No tools available."))
		} else {
			for _, tool := range serverTools.Tools {
				descStyle := lipgloss.NewStyle().Foreground(tokyoFg).Width(contentWidth).Align(lipgloss.Left)
				toolDesc := list.New().EnumeratorStyle(
					lipgloss.NewStyle().Foreground(tokyoGreen).MarginRight(1),
				).Item(descStyle.Render(tool.Description))
				serverList.Item(toolNameStyle.Render(tool.Name)).Item(toolDesc)
			}
		}
		l.Item(serverTools.ServerName).Item(serverList)
	}
	containerStyle := lipgloss.NewStyle().Margin(2).Width(width)
	fmt.Print("\n" + containerStyle.Render(l.String()) + "\n")
}

// Render and display content
func (c *Chat) renderContent(title, content string) {
	output, err := c.renderer.Render(content)
	if err != nil {
		log.Error().Err(err).Msg("render content failed")
		output = content
	}
	if title != "" {
		title = title + ": "
	}
	fmt.Printf("\n%s", responseStyle.Render(title+output))
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // Fallback width
	}
	return width - 20
}

var (
	// Tokyo Night theme colors
	tokyoPurple = lipgloss.Color("99")  // #9d7cd8
	tokyoCyan   = lipgloss.Color("73")  // #7dcfff
	tokyoBlue   = lipgloss.Color("111") // #7aa2f7
	tokyoGreen  = lipgloss.Color("120") // #73daca
	tokyoRed    = lipgloss.Color("203") // #f7768e
	tokyoOrange = lipgloss.Color("215") // #ff9e64
	tokyoFg     = lipgloss.Color("189") // #c0caf5
	tokyoGray   = lipgloss.Color("237") // #3b4261
	tokyoBg     = lipgloss.Color("234") // #1a1b26

	promptStyle = lipgloss.NewStyle().
			Foreground(tokyoBlue).
			PaddingLeft(2)

	responseStyle = lipgloss.NewStyle().
			Foreground(tokyoFg).
			PaddingLeft(2)

	errorStyle = lipgloss.NewStyle().
			Foreground(tokyoRed).
			Bold(true)

	toolNameStyle = lipgloss.NewStyle().
			Foreground(tokyoCyan).
			Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(tokyoFg).
				PaddingBottom(1)

	contentStyle = lipgloss.NewStyle().
			Background(tokyoBg).
			PaddingLeft(4).
			PaddingRight(4)
)
