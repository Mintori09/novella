package provider

func NewGroqProvider() *OpenAICompatibleProvider {
	return NewOpenAICompatibleProvider("groq", "https://api.groq.com/openai/v1", "llama-3.3-70b-versatile")
}
