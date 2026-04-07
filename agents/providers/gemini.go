package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type GeminiClient struct {
	model  string
	apiKey string
}

func NewGeminiClient(model string, apiKey string) *GeminiClient {
	return &GeminiClient{
		model:  model,
		apiKey: apiKey,
	}
}

func (c *GeminiClient) Query(prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable not set. Please set it in .env")
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.model, c.apiKey)
	
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]interface{}{{"text": prompt}}},
		},
	}
	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	
	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}
	
	// Traverse JSON `res.candidates[0].content.parts[0].text`
	candidates, ok := res["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "", fmt.Errorf("gemini invalid response structure: %s", string(body))
	}
	cand := candidates[0].(map[string]interface{})
	content := cand["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	text := parts[0].(map[string]interface{})["text"].(string)
	
	return text, nil
}

func (c *GeminiClient) Name() string {
	return "Gemini API (" + c.model + ")"
}
