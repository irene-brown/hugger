package apiv2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Statistics represents the statistics for a dataset.
type Statistics struct {
	NumExamples int              `json:"num_examples"`
	Statistics  []map[string]any `json:"statistics"`
	Partial     bool             `json:"partial"`
}

// GetDatasetStatistics fetches the statistics for a specified dataset and split.
func (client *HuggingFaceClient) GetDatasetStatistics(repoName, split string) (*Statistics, error) {
	// Construct the API URL for the specified dataset and split.
	url := fmt.Sprintf("https://datasets-server.huggingface.co/statistics?dataset=%s&config=cola&split=%s", repoName, split)

	// Create a new HTTP GET request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute the request using the client's configured doRequest method.
	res, err := client.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	// Read the response body.
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the JSON data into the Statistics struct.
	var stat Statistics
	if err := json.Unmarshal(data, &stat); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &stat, nil
}
