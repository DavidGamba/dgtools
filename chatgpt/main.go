/*
Chat GPT Terminal Client

Features:

- Use CUE file for configuration
- Load initial prompts from config
- Load colors from config, default to MAN PAGE | Pager colors.
- Load Key from OPENAI_API_KEY environment variable
- Maintains chat context

Endpoints in use:

- https://api.openai.com/v1/images/generations
- https://api.openai.com/v1/chat/completions
- https://api.openai.com/v1/completions

TODO:
- Add a way to indicate whether or not you want to add previous messages context and how many.

*/

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
)

var closingMessage = "Goodbye!"

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.String("config-file", "", opt.GetEnv("CHATGPT_CONFIG_FILE"))
	opt.SetCommandFn(UIRun)
	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
	}
	Logger.Println(remaining)

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

type customPainter struct{}

func (cp *customPainter) Paint(line []rune, pos int) []rune {
	s := string(line)
	c := color.New(color.FgBlue).SprintFunc()
	return []rune(c(s))
}

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	fmt.Println("Chat GPT Terminal Client")
	fmt.Println("Use .help to see available commands")

	// Operation mode: chat, image
	mode := "chat"

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          color.New(color.FgMagenta, color.Bold).Sprintf("💬 > "),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		Painter:         &customPainter{},
	})

	if err != nil {
		return fmt.Errorf("failed to create readline: %w", err)
	}
	defer rl.Close()

	messageHistory := &[]openai.ChatCompletionMessage{}
	for {
		// check context
		select {
		case <-ctx.Done():
			fmt.Printf("%s\n", closingMessage)
			return nil
		default:
		}

		// Read user input
		color.Set(color.FgBlue, color.Bold)
		input, err := rl.Readline()
		if err != nil {
			return err
		}
		color.Unset()

		if len(input) == 0 {
			continue
		}
		// Get first word from input
		firstWord := strings.Fields(input)[0]
		restOfInput := strings.TrimSpace(strings.TrimPrefix(input, firstWord))

		// Process user input
		switch firstWord {
		case "quit", "exit", ".quit", ".exit":
			fmt.Printf("%s\n", closingMessage)
			return nil
		case "fg", "ls":
			fmt.Printf("No-op!\n")
		case ".reset":
			// Reset chat context
			messageHistory = &[]openai.ChatCompletionMessage{}
		case ".clear", "clear":
			// clear but save contents into scrollback
			fmt.Printf("\033[H\033[22J")
		case ".image":
			mode = "image"
			color.New(color.FgBlue).Println("Changed mode to image")
			// TODO: Add a getoptions parser to get things like:
			// - Image size
			// - Image filename
			Logger.Printf("restOfInput: %s", restOfInput)
		case ".chat":
			mode = "chat"
			color.New(color.FgGreen).Println("Changed mode to chat")
		case ".output":
			color.New(color.FgGreen).Println("Saving context...")
			color.New(color.FgRed).Println("Unimplemented!")
		case ".help":
			color.New(color.FgGreen).Println(`Commands:
.help: Show this help message

.image: Change mode to image
.chat: Change mode to chat

.output: Save context to file
.reset: Reset chat context

fg: no-op

.clear, clear: Clear the terminal

.quit, .exit: Quit the program
`)
		default:
			switch mode {
			case "image":
				// TODO: Return image filename and print it on the terminal with chafa
				err := image(ctx, messageHistory, input, "")
				if err != nil {
					fmt.Printf("Error: %s\n", err)
				}
			default:
				err := chat(ctx, messageHistory, input)
				if err != nil {
					fmt.Printf("Error: %s\n", err)
				}
			}
		}
		err = printMessageHistoryContext(messageHistory)
		if err != nil {
			return fmt.Errorf("failed to print message history: %w", err)
		}

	}
}

func printMessageHistoryContext(messageHistory *[]openai.ChatCompletionMessage) error {
	size := 0
	for _, message := range *messageHistory {
		size += len(message.Content)
	}

	// append message to log file
	logFile, err := os.OpenFile("chatgpt.log.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()
	messageHistoryJSON, err := json.Marshal(*messageHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}
	_, err = logFile.Write(messageHistoryJSON)
	if err != nil {
		return fmt.Errorf("failed to write history: %w", err)
	}
	logFile.WriteString("\n")

	historyContext := color.New(color.FgGreen)
	historyContext.Printf("History Size: %d messages, %d bytes\n", len(*messageHistory), size)
	return nil
}

func image(ctx context.Context, messageHistory *[]openai.ChatCompletionMessage, message string, size string) error {
	// Post message to OpenAI
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// Add new user message
	*messageHistory = append(*messageHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	// Example image as base64
	reqBase64 := openai.ImageRequest{
		Prompt:         message,
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := client.CreateImage(ctx, reqBase64)
	if err != nil {
		return fmt.Errorf("image creation error: %v", err)
	}

	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		return fmt.Errorf("base64 decode error: %v", err)
	}

	r := bytes.NewReader(imgBytes)
	imgData, err := png.Decode(r)
	if err != nil {
		return fmt.Errorf("PNG decode error: %v", err)
	}

	suffix := time.Now().Format("20060102-150405")
	imageName := fmt.Sprintf("chatgpt-generated-%s.png", suffix)

	file, err := os.Create(imageName)
	if err != nil {
		return fmt.Errorf("file creation error: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, imgData); err != nil {
		return fmt.Errorf("PNG encode error: %v", err)
	}

	err = run.CMD("chafa", imageName).Log().Run()
	if err != nil {
		return fmt.Errorf("failed to display image with chafa: %w", err)
	}

	fmt.Printf("The image was saved as %s\n", imageName)

	return nil
}
