package apiv2

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

type Statistics struct {
	NumExamples	int			`json:"num_examples"`
	Statistics	[]map[string]any	`json:"statistics"`
	Partial		bool			`json:"partial"`
}

func(client *HuggingFaceClient) GetDatasetStatistics( repoName, split string ) (*Statistics, error) {
	url := fmt.Sprintf("https://datasets-server.huggingface.co/statistics?dataset=%s&config=cola&split=%s", repoName, split)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.doRequest( req )
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll( res.Body )
	if err != nil {
		return nil, err
	}

	var stat Statistics
	if err := json.Unmarshal(data, &stat); err != nil {
		return nil, err
	}
	return &stat, nil
}
