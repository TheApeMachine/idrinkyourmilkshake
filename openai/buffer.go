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
	totalTokens += buffer.estimateTokens(buffer.Messages[0].(openai.ChatCompletionSystemMessageParam).Content.String(), "system")
	totalTokens += buffer.estimateTokens(buffer.Messages[1].(openai.ChatCompletionUserMessageParam).Content.String(), "user")

	// Start from the most recent message for the rest
	for i := len(buffer.Messages) - 1; i >= 2; i-- {
		msg := buffer.Messages[i]

		var messageTokens int
		switch msg := msg.(type) {
		case openai.ChatCompletionSystemMessageParam:
			messageTokens = buffer.estimateTokens(msg.Content.String(), "system")
		case openai.ChatCompletionUserMessageParam:
			messageTokens = buffer.estimateTokens(msg.Content.String(), "user")
		case openai.ChatCompletionAssistantMessageParam:
			messageTokens = buffer.estimateTokens(msg.Content.String(), "assistant")
		case openai.ChatCompletionToolMessageParam:
			messageTokens = buffer.estimateTokens(msg.Content.String(), "tool")
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
