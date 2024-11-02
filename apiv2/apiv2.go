package apiv2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// MetadataResponse represents the unified metadata structure
type MetadataResponse struct {
	// Standard metadata fields
	Name        string            `json:"name"`
	Type        string            `json:"@type,omitempty"`
	Description string            `json:"description"`
	Creator     map[string]string `json:"creator"`
	URL         string            `json:"url"`
	License     string            `json:"license"`
	Keywords    []string          `json:"keywords"`

	// Croissant-specific fields
	Context       map[string]any   `json:"@context,omitempty"`
	Distribution  []map[string]any `json:"distribution,omitempty"`
	RecordSet     []map[string]any `json:"recordSet,omitempty"`
	ConformsTo    string           `json:"conformsTo,omitempty"`
	AlternateName []string         `json:"alternateName,omitempty"`
}

func (client *HuggingFaceClient) GetMetadata(repoType, repoID string) (*MetadataResponse, error) {
	// Try Croissant endpoint first
	croissantURL := fmt.Sprintf("%s/api/%s/%s/croissant", baseURL, repoType+"s", repoID)
	metadata, err := client.fetchMetadata(croissantURL)
	if err == nil {
		return metadata, nil
	}

	// Fall back to standard endpoint
	standardURL := fmt.Sprintf("%s/api/%s/%s", baseURL, repoType+"s", repoID)
	return client.fetchMetadata(standardURL)
}

func (client *HuggingFaceClient) fetchMetadata(url string) (*MetadataResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	res, err := client.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var metadata MetadataResponse
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &metadata, nil
}
