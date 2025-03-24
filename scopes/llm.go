package scopes

import (
	"fmt"
	"os"

	"github.com/vanclief/coderunner/files"
	"github.com/vanclief/coderunner/llm"
	"github.com/vanclief/coderunner/llm/chatgpt"
	"github.com/vanclief/coderunner/llm/claude"
	"github.com/vanclief/ez"
)

type LLMCallback func(path string, response string) error

func NewLLM(model string) (llm.API, error) {
	const op = "files.NewLLM"

	switch model {
	case "o1":
		fallthrough
	case "o1-mini":
		fallthrough
	case "4o":
		return NewChatGPTAPI(model)

	case "sonnet":
		return NewClaudeAPI("claude-3-5-sonnet-latest")

	default:
		errMsg := fmt.Sprintf("Invalid model: %s", model)
		return nil, ez.New(op, ez.EINVALID, errMsg, nil)
	}
}

func NewClaudeAPI(model string) (llm.API, error) {
	const op = "files.NewClaudeAPI"

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, ez.New(op, ez.EINTERNAL, "ANTHROPIC_API_KEY not set", nil)
	}
	return claude.NewAPI(apiKey, model, 2000)
}

func NewChatGPTAPI(model string) (llm.API, error) {
	const op = "files.NewChatGPTAPI"

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, ez.New(op, ez.EINTERNAL, "OPENAI_API_KEY not set", nil)
	}

	return chatgpt.NewAPI(apiKey, model)
}

// Refactored process function
func (s *Scope) RunPromptOnFiles(api llm.API, prompt string, callback LLMCallback) error {
	const op = "Scanner.RunPromptOnFiles"

	paths := s.GetAllFilePaths()

	return s.processFiles(paths, api, prompt, callback)
}

// Helper function for file processing logic
func (s *Scope) processFiles(paths []string, api llm.API, prompt string, callback LLMCallback) error {
	const op = "Scanner.processFiles"

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return ez.New(op, ez.EINTERNAL, "Failed to read file: "+path, err)
		}

		if files.IsBinaryFile(content) {
			continue
		}

		fullPrompt := fmt.Sprintf("%s\n\nFile Content:\n%s", prompt, string(content))

		fmt.Print("Calling LLM... ")
		response, err := api.Prompt(fullPrompt)
		if err != nil {
			fmt.Println("Failed", err)
			return ez.New(op, ez.EINTERNAL, "LLM processing failed for file: "+path, err)
		}

		fmt.Println("Ok")

		if err := callback(path, response); err != nil {
			return ez.New(op, ez.EINTERNAL, "LLMCallback failed for file: "+path, err)
		}
	}

	return nil
}
