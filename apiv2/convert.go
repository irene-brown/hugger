package apiv2

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Convert2 retrieves data in a specified format from a HuggingFace API endpoint.
func (client *HuggingFaceClient) Convert2(format, what string) ([]byte, error) {
	// Construct the URL using base URL, the type (what), and format.
	url := fmt.Sprintf("%s/api/%s/%s", baseURL, what, format)

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
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
