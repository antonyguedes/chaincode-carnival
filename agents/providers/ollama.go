package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

type OllamaClient struct {
	model string
	host  string
}

func NewOllamaClient(model string) *OllamaClient {
	h := os.Getenv("OLLAMA_HOST")
	if h == "" {
		h = "http://127.0.0.1:11434"
	}
	return &OllamaClient{
		model: model,
		host:  h,
	}
}

func (c *OllamaClient) Query(prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}
	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(fmt.Sprintf("%s/api/generate", c.host), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var oResp OllamaResponse
	if err := json.Unmarshal(body, &oResp); err != nil {
		return "", err
	}
	return oResp.Response, nil
}

func (c *OllamaClient) Name() string {
	return "Ollama (" + c.model + ")"
}
