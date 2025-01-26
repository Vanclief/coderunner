package files

import (
	"fmt"
	"os"

	"github.com/vanclief/coderunner/llm/claude"
	"github.com/vanclief/ez"
)

type LLMCallback func(path string, response string) error

// ProcessFilesWithLLM reads files and processes their content through an LLM
func (sm ScopeMap) ProcessFilesWithLLM(prompt string, callback LLMCallback) error {
	const op = "Scanner.ProcessFilesWithLLM"

	paths := make([]string, 0)
	collectPaths(sm, "", &paths)

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return ez.New(op, ez.EINTERNAL, "ANTHROPIC_API_KEY environment variable not set", nil)
	}

	llm := claude.NewAPI(apiKey, "claude-3-5-sonnet-latest", 2000)

	// Process each file through the LLM
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return ez.New(op, ez.EINTERNAL, "Failed to read file: "+path, err)
		}

		if IsBinaryFile(content) {
			continue // Skip binary files
		}

		// Combine the prompt with the file content
		fullPrompt := fmt.Sprintf("%s\n\nFile Content:\n%s", prompt, string(content))

		// Call the LLM with the combined prompt
		fmt.Print("Calling LLM... ")
		response, err := llm.Prompt(fullPrompt)
		if err != nil {
			fmt.Println("Failed", err)
			return ez.New(op, ez.EINTERNAL, "LLM processing failed for file: "+path, err)
		}

		fmt.Println("Ok")

		// Send the result back through the callback
		if err := callback(path, response); err != nil {
			return ez.New(op, ez.EINTERNAL, "LLMCallback failed for file: "+path, err)
		}
	}

	return nil
}
