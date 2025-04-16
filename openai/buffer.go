package openai

import (
	"github.com/charmbracelet/log"
	"github.com/openai/openai-go"
	"github.com/pkoukk/tiktoken-go"
)

type Buffer struct {
	Messages  []openai.ChatCompletionMessageParamUnion
	maxTokens int
}

func NewBuffer(systemPrompt, userPrompt string) *Buffer {
	return &Buffer{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		maxTokens: 128000,
	}
}

func (buffer *Buffer) Add(role, content string) {
	switch role {
	case "system":
		buffer.Messages = append(buffer.Messages, openai.SystemMessage(content))
	case "user":
		buffer.Messages = append(buffer.Messages, openai.UserMessage(content))
	case "assistant":
		buffer.Messages = append(buffer.Messages, openai.AssistantMessage(content))
	}
}

func (buffer *Buffer) AddToolMessage(id, content string) {
	buffer.Messages = append(buffer.Messages, openai.ToolMessage(id, content))
}

/*
Truncate the buffer to the maximum context tokens, making sure to always keep the
first two messages, which are the system prompt and the user message.
*/
func (buffer *Buffer) Truncate() *Buffer {
	// Always include first two messages (system prompt and user message)
	if len(buffer.Messages) < 2 {
		return buffer
	}

	maxTokens := buffer.maxTokens - 500 // Reserve tokens for response
	totalTokens := 0
	var truncatedMessages []openai.ChatCompletionMessageParamUnion

	// Add first two messages
	truncatedMessages = append(truncatedMessages, buffer.Messages[0], buffer.Messages[1])
	totalTokens += buffer.estimateTokens("system", buffer.Messages[0].OfSystem.Content.OfString.String())
	totalTokens += buffer.estimateTokens("user", buffer.Messages[1].OfUser.Content.OfString.String())

	// Start from the most recent message for the rest
	for i := len(buffer.Messages) - 1; i >= 2; i-- {
		msg := buffer.Messages[i]

		var messageTokens int
		if msg.OfSystem != nil {
			messageTokens = buffer.estimateTokens("system", msg.OfSystem.Content.OfString.String())
		} else if msg.OfUser != nil {
			messageTokens = buffer.estimateTokens("user", msg.OfUser.Content.OfString.String())
		} else if msg.OfAssistant != nil {
			messageTokens = buffer.estimateTokens("assistant", msg.OfAssistant.Content.OfString.String())
		} else if msg.OfTool != nil {
			messageTokens = buffer.estimateTokens("tool", msg.OfTool.Content.OfString.String())
		}

		if totalTokens+messageTokens <= maxTokens {
			truncatedMessages = append([]openai.ChatCompletionMessageParamUnion{msg}, truncatedMessages[2:]...)
			truncatedMessages = append(buffer.Messages[0:2], truncatedMessages...)
			totalTokens += messageTokens
		} else {
			break
		}
	}

	buffer.Messages = truncatedMessages
	return buffer
}

func (buffer *Buffer) estimateTokens(role, msg string) int {
	encoding, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		log.Error("Error getting encoding", "error", err)
		return 0
	}

	tokensPerMessage := 4 // As per OpenAI's token estimation guidelines

	numTokens := tokensPerMessage
	numTokens += len(encoding.Encode(msg, nil, nil))
	numTokens += len(encoding.Encode(role, nil, nil))

	return numTokens
}
