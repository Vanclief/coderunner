package llm

type API interface {
	Prompt(prompt string) (string, error)
}
