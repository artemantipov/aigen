package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	gogpt "github.com/sashabaranov/go-gpt3"
	"html"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "cli" {
		startCliChat()
	} else {
		startWeb()
	}
}

func startCliChat() {
	fmt.Println("Welcome to the AI assistant!")
	fmt.Println("Enter your request:")
	scanner := bufio.NewScanner(os.Stdin)
	prompt := []gogpt.ChatCompletionMessage{
		{Role: "assistant", Content: "Hello, how can I help you?"},
	}
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			break
		}
		prompt = append(prompt, gogpt.ChatCompletionMessage{Role: "user", Content: line})
		fmt.Println("AI:\n", chatAI(prompt))
		fmt.Println("You:")
	}
}

func chatAI(input []gogpt.ChatCompletionMessage) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	c := gogpt.NewClient(apiKey)
	ctx := context.Background()
	req := gogpt.ChatCompletionRequest{
		Model:       gogpt.GPT3Dot5Turbo,
		MaxTokens:   1000,
		Messages:    input,
		Temperature: 0.7,
	}
	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		return ""
	}
	return resp.Choices[0].Message.Content
}

func startWeb() {
	// Create a gin router
	router := gin.Default()
	err := router.SetTrustedProxies(nil)
	if err != nil {
		return
	}

	// Serve a browser page with a form
	// For user to enter a text
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "main.html", gin.H{})
	})
	router.GET("/summary", func(c *gin.Context) {
		c.HTML(http.StatusOK, "summary.html", gin.H{})
	})
	router.GET("/cover_letter", func(c *gin.Context) {
		c.HTML(http.StatusOK, "cover_letter.html", gin.H{})
	})

	//Post request for the form, receive and print out user input
	router.POST("/cover_letter", func(c *gin.Context) {
		jd := c.PostForm("jd")
		adds := c.PostForm("adds")
		c.HTML(http.StatusOK, "result.html", gin.H{
			"result": html.EscapeString(webPrompt("cover", jd, adds)),
		})
	})

	router.POST("/summary", func(c *gin.Context) {
		jp := c.PostForm("jp")
		adds := c.PostForm("adds")
		c.HTML(http.StatusOK, "result.html", gin.H{
			"result": webPrompt("summary", jp, adds),
		})
	})

	// Run the server on port 8000
	err = router.Run(":8080")
	if err != nil {
		fmt.Println("Error running server: ", err)
	}
}

func webPrompt(inputs ...string) string {
	prompt := []gogpt.ChatCompletionMessage{
		{Role: "system", Content: "You are a helpful assistant to a job seeker. You can help them with writing a cover letter or LinkedIn summary"},
	}
	switch inputs[0] {
	case "cover":
		if inputs[2] != "" {
			prompt = append(prompt, gogpt.ChatCompletionMessage{Role: "user", Content: fmt.Sprintf("Generate cover letter not more than 1000 characters, less formal, mention things like %s\n for job position according to description below \n\n%s", inputs[2], inputs[1])})
		} else {
			prompt = append(prompt, gogpt.ChatCompletionMessage{Role: "user", Content: fmt.Sprintf("Generate cover letter not more than 1000 characters, less formal\n for job position according to description below \n\n%s", inputs[1])})
		}
	case "summary":
		if inputs[2] != "" {
			prompt = append(prompt, gogpt.ChatCompletionMessage{Role: "user", Content: fmt.Sprintf("Generate LinkedIn summary for job position %s with common skills for that position, less formal and not more than 500 characters, mention things like %s", inputs[1], inputs[2])})
		} else {
			prompt = append(prompt, gogpt.ChatCompletionMessage{Role: "user", Content: fmt.Sprintf("Generate LinkedIn summary for job position %s with common skills for that position, less formal and not more than 500 characters", inputs[1])})
		}
	}

	return chatAI(prompt)
}
