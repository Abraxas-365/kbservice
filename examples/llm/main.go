// examples/llm/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Abraxas-365/kbservice/adapters/openai"
	"github.com/Abraxas-365/kbservice/llm"
)

func main() {
	ctx := context.Background()

	// Initialize OpenAI LLM
	llmClient := openai.NewOpenAILLM(
		os.Getenv("OPENAI_API_KEY"),
		"", // Use default model (gpt-4-turbo-preview)
	)

	// Example 1: Simple completion
	fmt.Println("=== Simple Completion ===")
	response, err := llmClient.Complete(ctx, "What is the capital of France?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n\n", response)

	// Example 2: Chat with multiple messages
	fmt.Println("=== Chat Conversation ===")
	messages := []llm.Message{
		{
			Role:    llm.RoleSystem,
			Content: "You are a helpful assistant that provides concise answers.",
		},
		{
			Role:    llm.RoleUser,
			Content: "What are the main benefits of using Go for backend development?",
		},
	}

	chatResp, err := llmClient.Chat(ctx, messages)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Assistant: %s\n\n", chatResp.Content)

	// Example 3: Streaming response
	fmt.Println("=== Streaming Response ===")
	streamMessages := []llm.Message{
		{
			Role:    llm.RoleUser,
			Content: "Write a short poem about coding.",
		},
	}

	stream, err := llmClient.ChatStream(ctx, streamMessages)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Assistant: ")
	for response := range stream {
		if response.Error != nil {
			log.Fatal(response.Error)
		}
		fmt.Print(response.Message.Content)
		if response.Done {
			break
		}
	}
	fmt.Println("\n")

	// Example 4: Function calling
	fmt.Println("=== Function Calling ===")
	functionsMessages := []llm.Message{
		{
			Role:    llm.RoleUser,
			Content: "What's the weather like in Paris?",
		},
	}

	weatherFunction := llm.Function{
		Name:        "get_weather",
		Description: "Get the current weather in a location",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"location": map[string]interface{}{
					"type":        "string",
					"description": "The city and country",
				},
				"unit": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"celsius", "fahrenheit"},
					"description": "The temperature unit",
				},
			},
			"required": []string{"location"},
		},
	}

	functionResp, err := llmClient.Chat(ctx, functionsMessages,
		llm.WithFunctions([]llm.Function{weatherFunction}),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Assistant: %s\n", functionResp.FuncCall.Arguments)
}
