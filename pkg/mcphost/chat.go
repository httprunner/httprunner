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
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

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

// NewChat creates a new chat session
func (h *MCPHost) NewChat(ctx context.Context, systemPromptFile string) (*Chat, error) {
	// Get model config from environment variables
	modelConfig, err := ai.GetModelConfig(option.LLMServiceTypeGPT)
	if err != nil {
		return nil, err
	}
	model, err := openai.NewChatModel(ctx, modelConfig.ChatModelConfig)
	if err != nil {
		return nil, errors.Wrap(code.LLMPrepareRequestError, err.Error())
	}

	// Create markdown renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(styles.TokyoNightStyle),
		glamour.WithWordWrap(getTerminalWidth()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create markdown renderer")
	}

	// Load system prompt from file if provided
	systemPrompt := "chat to interact with MCP tools"
	if systemPromptFile != "" {
		customPrompt, err := loadSystemPrompt(systemPromptFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load system prompt")
		}
		if customPrompt != "" {
			systemPrompt = customPrompt
		}
	}

	// convert MCP tools to eino tool infos
	einoTools, err := h.GetEinoToolInfos(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get eino tool infos")
	}

	toolCallingModel, err := model.WithTools(einoTools)
	if err != nil {
		return nil, errors.Wrap(code.LLMPrepareRequestError, err.Error())
	}

	return &Chat{
		model:        toolCallingModel,
		systemPrompt: systemPrompt,
		history:      ai.ConversationHistory{},
		renderer:     renderer,
		host:         h,
		tools:        einoTools,
	}, nil
}

// Chat represents a chat session with LLM
type Chat struct {
	model        model.ToolCallingChatModel
	systemPrompt string
	history      ai.ConversationHistory
	renderer     *glamour.TermRenderer
	host         *MCPHost
	tools        []*schema.ToolInfo
}

// Start starts the chat session
func (c *Chat) Start() error {
	// Add system message
	c.history = ai.ConversationHistory{
		{
			Role:    schema.System,
			Content: c.systemPrompt,
		},
	}

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
		if err := c.runPrompt(input); err != nil {
			log.Error().Err(err).Msg("chat error")
		}
	}
}

// runPrompt run prompt with MCP tools
func (c *Chat) runPrompt(prompt string) error {
	fmt.Printf("\n%s\n", promptStyle.Render("You: "+prompt))

	// Create user message
	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
	c.history = append(c.history, userMsg)
	for {
		ctx := context.Background()
		var resp *schema.Message
		var err error
		action := func() {
			resp, err = c.model.Generate(ctx, c.history)
		}
		_ = spinner.New().Title("Thinking...").Action(action).Run()
		if err != nil {
			return err
		}

		// Handle tool calls
		toolCalls := resp.ToolCalls
		if len(toolCalls) > 0 {
			for _, toolCall := range toolCalls {
				parts := strings.SplitN(toolCall.Function.Name, "__", 2)
				if len(parts) != 2 {
					log.Error().Msgf("invalid tool name: %s", toolCall.Function.Name)
					continue
				}
				serverName, toolName := parts[0], parts[1]
				args := toolCall.Function.Arguments

				// Unmarshal tool arguments from JSON string
				var argsMap map[string]interface{}
				if err := sonic.UnmarshalString(args, &argsMap); err != nil {
					log.Error().Err(err).Str("args", args).Msg("failed to unmarshal tool arguments")
					continue
				}

				result, err := c.host.InvokeTool(ctx, serverName, toolName, argsMap)
				if err != nil {
					log.Error().Err(err).Msg("tool call failed")
					continue
				}

				// Format tool result
				resultStr := ""
				if result != nil && len(result.Content) > 0 {
					for _, item := range result.Content {
						resultStr += fmt.Sprintf("%v\n", item)
					}
				} else {
					resultStr = fmt.Sprintf("%+v", result)
				}

				// Add tool result to history
				toolMsg := &schema.Message{
					Role:    schema.Assistant,
					Content: resultStr,
				}
				c.history = append(c.history, toolMsg)
			}
			continue
		}

		// Add assistant's response to history
		c.history = append(c.history, resp)

		// Render and display response
		if rendered, err := c.renderer.Render(resp.Content); err == nil {
			fmt.Printf("\n%s", responseStyle.Render("Assistant: "+rendered))
		} else {
			fmt.Printf("\n%s", errorStyle.Render("Assistant: "+resp.Content))
		}

		return nil
	}
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

- **system-prompt**: %s
- **mcp-config**: %s
`, c.systemPrompt, c.host.config.ConfigPath)

	str, err := c.renderer.Render(markdown)
	if err != nil {
		fmt.Println(markdown)
	} else {
		fmt.Print(str)
	}
}

func (c *Chat) handleCommand(cmd string) error {
	switch cmd {
	case "/help":
		c.showWelcome()
	case "/tools":
		c.showTools()
	case "/history":
		c.showHistory()
	case "/clear":
		c.clearHistory()
	case "/quit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
	return nil
}

func (c *Chat) showHistory() {
	if len(c.history) <= 1 { // Only system message
		fmt.Println("No conversation history yet.")
		return
	}

	fmt.Println("\nConversation History:")
	for _, msg := range c.history {
		if msg.Role == schema.System {
			continue
		}

		role := "You"
		if msg.Role == schema.Assistant {
			role = "Assistant"
		}

		// Render message content as markdown
		rendered, err := c.renderer.Render(msg.Content)
		if err != nil {
			rendered = msg.Content
		}

		fmt.Printf("\n%s: %s\n", role, rendered)
	}
}

func (c *Chat) clearHistory() {
	// Keep only the system message
	systemMsg := c.history[0]
	c.history = ai.ConversationHistory{systemMsg}
	fmt.Println("Conversation history cleared.")
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

// loadSystemPrompt loads the system prompt from a JSON file
func loadSystemPrompt(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("system prompt file does not exist: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading prompt file: %v", err)
	}

	// Read file content directly as prompt
	return string(data), nil
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // Fallback width
	}
	return width - 20
}
