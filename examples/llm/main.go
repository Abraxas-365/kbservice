package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Abraxas-365/kbservice/adapters/openai"
	"github.com/Abraxas-365/kbservice/llm"
)

type UserData struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone,omitempty"`
}

func main() {
	ctx := context.Background()

	llmClient := openai.NewOpenAILLM(
		os.Getenv("OPENAI_API_KEY"),
		"gpt-4-turbo-preview",
	)

	messages := []llm.Message{
		{
			Role: llm.RoleSystem,
			Content: "You are a Apple product seller manager, your objective is to lead " +
				"the user to be a lead and get his data, but only ask for the data if the user " +
				"is interesting in buying, or when he is considered a potential lead. Be friendly " +
				"and professional.",
		},
	}

	userDataTool := llm.Function{
		Name:        "send_user_data",
		Description: "Upsert user information",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "User's full name",
				},
				"email": map[string]interface{}{
					"type":        "string",
					"description": "User's email address",
				},
				"phone": map[string]interface{}{
					"type":        "string",
					"description": "User's phone number",
				},
			},
			"required": []string{"name", "email"},
		},
	}

	humanTool := llm.Function{
		Name:        "ask_human",
		Description: "Use this tool to ask information to the user when needed",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"question": map[string]interface{}{
					"type":        "string",
					"description": "The question to ask the user",
				},
			},
			"required": []string{"question"},
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to Apple Store Assistant! Type 'exit' to quit.")
	fmt.Print("You: ")

	for scanner.Scan() {
		userInput := scanner.Text()
		if strings.ToLower(userInput) == "exit" {
			break
		}

		messages = append(messages, llm.Message{
			Role:    llm.RoleUser,
			Content: userInput,
		})

		// Get AI response
		response, err := llmClient.Chat(ctx, messages,
			llm.WithFunctions([]llm.Function{userDataTool, humanTool}),
		)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		// Handle function calls
		if response.FuncCall != nil {
			switch response.FuncCall.Name {
			case "send_user_data":
				var userData UserData
				if err := json.Unmarshal([]byte(response.FuncCall.Arguments), &userData); err != nil {
					log.Printf("Error parsing user data: %v\n", err)
					continue
				}

				messages = append(messages, *response)
				messages = append(messages, llm.Message{
					Role:    llm.RoleFunction,
					Name:    "send_user_data",
					Content: "User data saved successfully",
				})

				fmt.Printf("Assistant: Thank you! I've saved your information.\n")

			case "ask_human":
				var question struct {
					Question string `json:"question"`
				}
				if err := json.Unmarshal([]byte(response.FuncCall.Arguments), &question); err != nil {
					log.Printf("Error parsing question: %v\n", err)
					continue
				}

				messages = append(messages, *response)
				fmt.Printf("Assistant: %s\n", question.Question)
				messages = append(messages, llm.Message{
					Role:    llm.RoleAssistant,
					Content: question.Question,
				})
			}
		} else if response.Content != "" {
			fmt.Printf("Assistant: %s\n", response.Content)
			messages = append(messages, *response)
		}

		fmt.Print("You: ")
	}

	fmt.Println("\nThank you for using Apple Store Assistant!")
}
