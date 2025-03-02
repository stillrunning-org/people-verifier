package chatgpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

const apiURL = "https://api.openai.com/v1/chat/completions"

// Define structures to match the API format
type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	FileIDs []string `json:"file_ids,omitempty"`
}

type RequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ResponseChoice struct {
	Message Message `json:"message"`
}

type OpenAIResponse struct {
	Choices []ResponseChoice `json:"choices"`
}

func getApiKey() string {
	var apiKey = os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing OpenAI API key. Set the OPENAI_API_KEY environment variable.")
	}
	return apiKey
}

type FileUploadResponse struct {
	ID string `json:"id"`
}

func UploadFileToOpenAI(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a buffer to store multipart data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the "purpose" field (required by OpenAI API)
	_ = writer.WriteField("purpose", "assistants")

	// Create a file form field
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}

	// Copy the file content into the form file field
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %v", err)
	}

	// Close the multipart writer to finalize the request body
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/files", &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+getApiKey())
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("file upload failed: %s", string(body))
	}

	// Parse JSON response
	var uploadResp FileUploadResponse
	err = json.Unmarshal(body, &uploadResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return uploadResp.ID, nil
}

const (
	openAIAPIURL = "https://api.openai.com/v1/chat/completions"
)

func AskChatGPT(messages []Message) (string, error) {
	requestBody := RequestBody{
		Model:    "gpt-4",
		Messages: messages,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", openAIAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getApiKey())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from OpenAI")
}
