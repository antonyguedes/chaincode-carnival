package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CohereClient struct {
	model  string
	apiKey string
}

func NewCohereClient(model, apiKey string) *CohereClient {
	return &CohereClient{model: model, apiKey: apiKey}
}

func (c *CohereClient) Query(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": c.model,
		"message": prompt,
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.cohere.com/v1/chat", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)

	var res map[string]interface{}
	json.Unmarshal(body, &res)

	text, ok := res["text"].(string)
	if !ok {
		return "", fmt.Errorf("cohere invalid response: %s", string(body))
	}
	
	return text, nil
}

func (c *CohereClient) Name() string {
	return "Cohere (" + c.model + ")"
}
