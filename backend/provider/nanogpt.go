package provider

func NewNanoGPTProvider() *OpenAICompatibleProvider {
	return NewOpenAICompatibleProvider("nanogpt", "https://nano-gpt.com/api/v1", "zai-org/glm-5")
}
