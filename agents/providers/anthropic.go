package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type AnthropicClient struct {
	model  string
	apiKey string
}

func NewAnthropicClient(model, apiKey string) *AnthropicClient {
	return &AnthropicClient{model: model, apiKey: apiKey}
}

func (c *AnthropicClient) Query(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 4096,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)

	var res map[string]interface{}
	json.Unmarshal(body, &res)

	content, ok := res["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("anthropic invalid response: %s", string(body))
	}
	
	text := content[0].(map[string]interface{})["text"].(string)
	return text, nil
}

func (c *AnthropicClient) Name() string {
	return "Anthropic Claude (" + c.model + ")"
}
