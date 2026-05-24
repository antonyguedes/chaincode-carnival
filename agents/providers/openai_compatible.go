package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type OpenAICompatibleClient struct {
	name    string
	baseURL string
	model   string
	apiKey  string
}

// NewOpenAICompatibleClient builds a generic client for any API adhering to the OpenAI v1 mapping constraint.
func NewOpenAICompatibleClient(name, baseURL, model, apiKey string) *OpenAICompatibleClient {
	return &OpenAICompatibleClient{
		name:    name,
		baseURL: baseURL,
		model:   model,
		apiKey:  apiKey,
	}
}

func (c *OpenAICompatibleClient) Query(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
		// Ensure enough tokens for structured JSON responses.
		// Small models (llama-3.1-8b) default to very low limits and truncate mid-JSON.
		"max_tokens": 2048,
	}
	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}
	
	choices, ok := res["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("%s invalid response structure: %s", c.name, string(body))
	}
	msg := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := msg["content"].(string)
	
	return content, nil
}

func (c *OpenAICompatibleClient) Name() string {
	return fmt.Sprintf("%s (%s)", c.name, c.model)
}
