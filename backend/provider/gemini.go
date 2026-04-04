package provider

func NewGeminiProvider() *OpenAICompatibleProvider {
	return NewOpenAICompatibleProvider("gemini", "https://generativelanguage.googleapis.com/v1beta/openai", "gemini-2.0-flash")
}
