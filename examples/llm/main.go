package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Abraxas-365/kbservice/adapters/inmemory"
	"github.com/Abraxas-365/kbservice/adapters/openai"
	"github.com/Abraxas-365/kbservice/chathistory"
	"github.com/Abraxas-365/kbservice/llm"
)

type UserData struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone,omitempty"`
}

func main() {
	ctx := context.Background()

	// Initialize LLM client
	llmClient := openai.NewOpenAILLM(
		os.Getenv("OPENAI_API_KEY"),
		"gpt-4-turbo-preview",
	)

	// Initialize chat history
	repo := inmemory.NewInMemoryRepository()
	memory := chathistory.New(repo,
		chathistory.WithSystemPrompt("You are a Apple product seller manager, your objective is to lead "+
			"the user to be a lead and get his data, but only ask for the data if the user "+
			"is interesting in buying, or when he is considered a potential lead. Be friendly "+
			"and professional."),
		chathistory.WithMaxMessages(100),
		chathistory.WithReturnLimit(50),
	)

	// Create a new conversation
	conv, err := memory.CreateConversation(ctx, map[string]any{
		"type": "sales_conversation",
	})
	if err != nil {
		log.Fatalf("Failed to create conversation: %v", err)
	}

	// Define tools
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

		// Add user message to history
		err = memory.AddMessage(ctx, conv.ID, llm.Message{
			Role:    llm.RoleUser,
			Content: userInput,
		})
		if err != nil {
			log.Printf("Error adding message: %v\n", err)
			continue
		}

		// Get conversation history
		messages, err := memory.GetMessages(ctx, conv.ID, 0)
		if err != nil {
			log.Printf("Error getting messages: %v\n", err)
			continue
		}

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

				// Update conversation metadata with user data
				err = memory.UpdateConversationMetadata(ctx, conv.ID, map[string]any{
					"user_data": userData,
				})
				if err != nil {
					log.Printf("Error updating metadata: %v\n", err)
				}

				// Add assistant's function call to history
				err = memory.AddMessage(ctx, conv.ID, *response)
				if err != nil {
					log.Printf("Error adding message: %v\n", err)
				}

				// Add function result to history
				err = memory.AddMessage(ctx, conv.ID, llm.Message{
					Role:    llm.RoleFunction,
					Name:    "send_user_data",
					Content: "User data saved successfully",
				})
				if err != nil {
					log.Printf("Error adding message: %v\n", err)
				}

				fmt.Printf("Assistant: Thank you! I've saved your information.\n")

			case "ask_human":
				var question struct {
					Question string `json:"question"`
				}
				if err := json.Unmarshal([]byte(response.FuncCall.Arguments), &question); err != nil {
					log.Printf("Error parsing question: %v\n", err)
					continue
				}

				// Add assistant's function call to history
				err = memory.AddMessage(ctx, conv.ID, *response)
				if err != nil {
					log.Printf("Error adding message: %v\n", err)
				}

				fmt.Printf("Assistant: %s\n", question.Question)

				// Add question as assistant message
				err = memory.AddMessage(ctx, conv.ID, llm.Message{
					Role:    llm.RoleAssistant,
					Content: question.Question,
				})
				if err != nil {
					log.Printf("Error adding message: %v\n", err)
				}
			}
		} else if response.Content != "" {
			fmt.Printf("Assistant: %s\n", response.Content)

			// Add assistant's response to history
			err = memory.AddMessage(ctx, conv.ID, *response)
			if err != nil {
				log.Printf("Error adding message: %v\n", err)
			}
		}

		fmt.Print("You: ")
	}

	// Get final conversation state
	finalConv, err := memory.GetConversation(ctx, conv.ID)
	if err != nil {
		log.Printf("Error getting final conversation: %v\n", err)
	} else {
		// Log conversation summary
		messageCount, _ := memory.GetMessageCount(ctx, conv.ID, chathistory.Filter{})
		fmt.Printf("\nConversation Summary:\n")
		fmt.Printf("ID: %s\n", finalConv.ID)
		fmt.Printf("Messages: %d\n", messageCount)
		fmt.Printf("Duration: %v\n", finalConv.UpdatedAt.Sub(finalConv.CreatedAt))
		if userData, ok := finalConv.Metadata["user_data"]; ok {
			fmt.Printf("User Data: %+v\n", userData)
		}
	}

	fmt.Println("\nThank you for using Apple Store Assistant!")
}
