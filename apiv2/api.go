package apiv2

import (
	"strings"
	"bytes"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	baseURL = "https://huggingface.co"
	UserAgent = "huggingface-cli/None; hf_hub/0.24.6; python/3.12.10; torch/2.4.1; tensorflow/2.17.0"
)


type HFRepo struct {
	RepoID		string	`json:"repo_id"`
	RepoType	string	`json:"type"`
	Name		string	`json:"name"`
	Private		bool	`json:"private"`
}

type HFFile struct {
	Type	string	`json:"type"`
	Oid	string	`json:"oid"`
	Size	uint	`json:"size"`
	Path	string	`json:"path"`
}

type HuggingFaceClient struct {
	APIKey string
}

func NewHuggingFaceClient(apiKey string) *HuggingFaceClient {
	return &HuggingFaceClient{APIKey: apiKey}
}

func (client *HuggingFaceClient) CreateRepo( repoType, datasetName string) error {
	url := fmt.Sprintf("%s/api/repos/create", baseURL)

	parts := strings.Split( datasetName, "/" )
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo name")
	}

	payload := HFRepo {
		datasetName,
		repoType,
		parts[1],
		true,
	}

	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("Content-Type", "application/json")

	_, err = client.doRequest(req)
	return err
}

type UFile struct {
	Path	string	`json:"path"`
	Sample	string	`json:"sample"`
	Size	int	`json:"size"`
}

type UFiles struct {
	Files	[]UFile	`json:"files"`
}

type UFileResp struct {
	Path	string	`json:"path"`
	ShouldIgnore	bool	`json:"shouldIgnore"`
	UploadMode	string	`json:"uploadMode"`
}

type UFileResponse struct {
	CommitOID	string		`json:"commitOid"`
	Files		[]UFileResp	`json:"files"`
}

type KeyValue struct {
	Key	string			`json:"key"`
	Value	map[string]string	`json:"value"`
}

func (client *HuggingFaceClient) UploadFile( repoType, datasetName, filePath string, contents []byte) error {

	/*
	 * pre-upload request
	 */
	url := fmt.Sprintf("%s/api/%s/%s/preupload/main", baseURL, repoType + "s", datasetName)

	/*
	 * upload file contents in base64-encoding (as huggingface-cli does)
	 */
	contents64 := base64.StdEncoding.EncodeToString( contents )
	ufiles := UFiles {
		[]UFile{
			UFile {
				filePath,
				contents64,
				len(contents), // not len(contents64)!
			},
		},
	}
	data, _ := json.Marshal( ufiles )
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer " + client.APIKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll( resp.Body )
	if err != nil {
		return err
	}


	var pre_response UFileResponse
	if err := json.Unmarshal( data, &pre_response ); err != nil {
		return err
	}

	/*
	 * really upload file via /commit API endpoint
	 */
	url = fmt.Sprintf("%s/api/%s/%s/commit/main", baseURL, repoType + "s", datasetName)
	kv := KeyValue{
		"header",
		map[string]string {
			"summary": "Upload " + filePath,
			"description": "",
		},
	}
	data, _ = json.Marshal( kv )
	kv = KeyValue{
		"file",
		map[string]string {
			"content": contents64,
			"path": filePath,
			"encoding": "base64",
		},
	}
	tmp, _ := json.Marshal( kv )
	data = append( data, 0x0a )
	data = append(data, tmp...)

	req, err = http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer " + client.APIKey)
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err = client.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll( resp.Body )
	return err

}


func (client *HuggingFaceClient) DownloadFile( repoType, datasetName, fileName string) ( []byte, error ) {

	url := fmt.Sprintf("%s/%s/%s/resolve/main/%s", baseURL, repoType + "s", datasetName, fileName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	resp, err := client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (client *HuggingFaceClient) DeleteRepo(datasetName string) error {
	url := fmt.Sprintf("%s/api/repos/delete", baseURL)

	parts := strings.Split( datasetName, "/" )
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo name")
	}

	payload := HFRepo {
		datasetName,
		"dataset",
		parts[1],
		true,
	}

	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("Content-Type", "application/json")

	_, err = client.doRequest(req)
	return err
}

func (client *HuggingFaceClient) DeleteFile( repoType, datasetName, filePath string) error {

	url := fmt.Sprintf("%s/api/%s/%s/commit/main", baseURL, repoType + "s", datasetName)

	kv := KeyValue{
		"header",
		map[string]string {
			"summary": "Delete " + filePath,
			"description": "",
		},
	}
	data, _ := json.Marshal( kv )
	kv = KeyValue{
		"deletedFile",
		map[string]string {
			"path": filePath,
		},
	}
	tmp, _ := json.Marshal( kv )
	data = append( data, 0x0a )
	data = append(data, tmp...)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer " + client.APIKey)
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := client.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll( resp.Body )
	return err
}


func (client *HuggingFaceClient) ListFilesInPrivateDataset( repoType, datasetName, path string ) ([]string, error) {

	url := fmt.Sprintf("%s/api/%s/%s/tree/main", baseURL, repoType + "s", datasetName)
	if len(path) > 0 {
		url += "/" + path
	}

	//fmt.Printf("\nVisiting %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer " + client.APIKey)
	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do( req )
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll( resp.Body )
		fmt.Println(string(body))
		return nil, fmt.Errorf("error: received status code %d, response: %s",
				resp.StatusCode, body)
	}

	body, err := ioutil.ReadAll( resp.Body )
	if err != nil {
		fmt.Println(string(body))
		return nil, err
	}

	var files []HFFile
	if err := json.Unmarshal( body, &files ); err != nil {
		return nil, err
	}

	var totalFiles []string
	for _, f := range files {
		if f.Type == "file" {
			totalFiles = append( totalFiles, f.Path )
		} else if f.Type == "directory" {
			dirFiles, err := client.ListFilesInPrivateDataset( repoType, datasetName, f.Path )
			if err != nil {
				return nil, err
			}
			totalFiles = append( totalFiles, dirFiles... )
		}
	}

	return totalFiles, nil
}

func (client *HuggingFaceClient) doRequest(req *http.Request) (*http.Response, error) {
	
	req.Header.Set("user-agent", UserAgent)
	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("error: %s", body)
	}
	return resp, nil
}
