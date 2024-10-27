package apiv2

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

type CroissantMetadata struct {

	Context		map[string]any		`json:"@context"`
	Type		string			`json:"@type"`
	Distribution	[]map[string]any	`json:"distribution"`
	RecordSet	[]map[string]any	`json:"recordSet"`
	ConformsTo	string			`json:"conformsTo"`
	Name		string			`json:"name"`
	Description	string			`json:"description"`
	AlternateName	[]string		`json:"alternateName"`
	Creator		map[string]string	`json:"creator"`
	Keywords	[]string		`json:"keywords"`
	License		string			`json:"license"`
	URL		string			`json:"url"`
}

func(client *HuggingFaceClient) GetMetadata( repoType string, repoID string ) (*CroissantMetadata, error) {

	url := fmt.Sprintf("%s/api/%s/%s/croissant", baseURL, repoType + "s", repoID)
	req, err := http.NewRequest( "GET", url, nil )
	if err != nil {
		return nil, err
	}
	req.Header.Set( "Authorization", "Bearer " + client.APIKey )
	res, err := client.doRequest( req )
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll( res.Body )
	if err != nil {
		return nil, err
	}

	var cr CroissantMetadata
	if err := json.Unmarshal( data, &cr ); err != nil {
		return nil, err
	}
	return &cr, nil
}
