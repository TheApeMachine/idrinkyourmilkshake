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

func (b *Buffer) Add(role string, content string) {
	switch role {
	case "system":
		b.Messages = append(b.Messages, openai.SystemMessage(content))
	case "user":
		b.Messages = append(b.Messages, openai.UserMessage(content))
	case "assistant":
		b.Messages = append(b.Messages, openai.AssistantMessage(content))
	default:
		// For other roles, we'll just log an error as the API only accepts specific roles
		log.Error("Unsupported role", "role", role)
		// Fall back to user message
		b.Messages = append(b.Messages, openai.UserMessage(content))
	}
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
	totalTokens += buffer.estimateTokens("system", buffer.Messages[0].(openai.ChatCompletionSystemMessageParam).Content.String())
	totalTokens += buffer.estimateTokens("user", buffer.Messages[1].(openai.ChatCompletionUserMessageParam).Content.String())

	// Start from the most recent message for the rest
	for i := len(buffer.Messages) - 1; i >= 2; i-- {
		msg := buffer.Messages[i]

		switch msg := msg.(type) {
		case openai.ChatCompletionAssistantMessageParam:
			totalTokens += buffer.estimateTokens("assistant", msg.Content.String())
		case openai.ChatCompletionUserMessageParam:
			totalTokens += buffer.estimateTokens("user", msg.Content.String())
		case openai.ChatCompletionSystemMessageParam:
			totalTokens += buffer.estimateTokens("system", msg.Content.String())
		case openai.ChatCompletionToolMessageParam:
			totalTokens += buffer.estimateTokens("tool", msg.Content.String())
		default:
		}
		if totalTokens <= maxTokens {
			truncatedMessages = append([]openai.ChatCompletionMessageParamUnion{msg}, truncatedMessages[2:]...)
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
