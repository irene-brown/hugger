package apiv2

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	baseURL   = "https://huggingface.co"
	// TODO: generate user agent according to user's configuration
	UserAgent = "hugger/v0.3.0; None; hf_hub/0.24.6; python/3.12.10; torch/2.4.1; tensorflow/2.17.0"
)

// Core structs for API operations
type HuggingFaceClient struct {
	APIKey string
	Token  string
}

type HFRepo struct {
	RepoID   string `json:"repo_id"`
	RepoType string `json:"type"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
}

type HFFile struct {
	Type string `json:"type"`
	Oid  string `json:"oid"`
	Size uint   `json:"size"`
	Path string `json:"path"`
}

type UFile struct {
	Path   string `json:"path"`
	Sample string `json:"sample"`
	Size   int    `json:"size"`
}

type UFiles struct {
	Files []UFile `json:"files"`
}

type UFileResponse struct {
	Files []struct {
		Path string `json:"path"`
		Blob struct {
			Size int    `json:"size"`
			Oid  string `json:"oid"`
		} `json:"blob"`
	} `json:"files"`
}

type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func NewHuggingFaceClient(apiKey string) *HuggingFaceClient {
	return &HuggingFaceClient{APIKey: apiKey}
}

// Core API methods
func (client *HuggingFaceClient) CreateRepo(repoType, datasetName string, private bool) error {
	url := fmt.Sprintf("%s/api/repos/create", baseURL)
	parts := strings.Split(datasetName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("repo name must be in format 'username/repo-name'")
	}

	payload := HFRepo{
		RepoID:   datasetName,
		RepoType: repoType,
		Name:     parts[1],
		Private:  private,
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("could not create repository request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("Content-Type", "application/json")

	_, err = client.doRequest(req)
	if err != nil {
		return err //fmt.Errorf("%v", err)
	}
	fmt.Println("✨ Success! Your new repo is ready for action.")
	return nil
}

func (client *HuggingFaceClient) UploadFile(repoType, datasetName, filePath string, contents []byte) error {
	// Pre-upload request
	url := fmt.Sprintf("%s/api/%s/%s/preupload/main", baseURL, repoType+"s", datasetName)

	contents64 := base64.StdEncoding.EncodeToString(contents)
	ufiles := UFiles{
		Files: []UFile{
			{Path: filePath, Sample: contents64, Size: len(contents)},
		},
	}

	data, _ := json.Marshal(ufiles)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to prepare upload: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.doRequest(req)
	if err != nil {
		return fmt.Errorf("pre-upload request failed: %v", err)
	}
	defer resp.Body.Close()

	// Commit file upload
	url = fmt.Sprintf("%s/api/%s/%s/commit/main", baseURL, repoType+"s", datasetName)
	kv := KeyValue{
		Key: "header",
		Value: map[string]string{
			"summary":     "Uploading " + filePath,
			"description": "",
		},
	}
	data, _ = json.Marshal(kv)
	kv = KeyValue{
		Key: "file",
		Value: map[string]string{
			"content":  contents64,
			"path":     filePath,
			"encoding": "base64",
		},
	}
	tmp, _ := json.Marshal(kv)
	data = append(data, 0x0a)
	data = append(data, tmp...)

	req, err = http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("file upload request failed: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("Content-Type", "application/x-ndjson")

	_, err = client.doRequest(req)
	if err != nil {
		return fmt.Errorf("file upload failed: %v", err)
	}
	fmt.Println("🚀 File uploaded successfully!")
	return nil
}

func (client *HuggingFaceClient) doRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", UserAgent)
	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("received error: %s", body)
	}
	return resp, nil
}

func (client *HuggingFaceClient) DownloadFile(repoType, repoName, filePath string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/%s/resolve/main/%s", baseURL, repoType+"s", repoName, filePath)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+client.Token)

	resp, err := client.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (client *HuggingFaceClient) DeleteRepo(repoName string) error {
	url := fmt.Sprintf("%s/api/repos/delete", baseURL)
	payload := map[string]string{"name": repoName}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+client.Token)
	req.Header.Set("Content-Type", "application/json")

	_, err = client.doRequest(req)
	return err
}

func (client *HuggingFaceClient) DeleteFile(repoType, repoName, filePath string) error {
	url := fmt.Sprintf("%s/api/%s/%s/commit/main", baseURL, repoType + "s", repoName)

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

func (client *HuggingFaceClient) ListFilesInRepo(repoType, repoName, path string, recursive bool) ([]string, error) {
	// please, do not touch this function.
	// yes, I know huggingface API has more beautiful way to list files
	// however when I try to use it I'm getting error 404:
	// 'Sorry, we can't find the page you are looking for.'
	
	// I'll fix this issue as soon as I understand what I'm doing wrong.

	url := fmt.Sprintf("%s/api/%s/%s/tree/main", baseURL, repoType + "s", repoName)
	if len(path) > 0 {
		url += "/" + path
	}

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
			if recursive {
				dirFiles, err := client.ListFilesInRepo( repoType, repoName, f.Path, recursive )
				if err != nil {
					return nil, err
				}
				totalFiles = append( totalFiles, dirFiles... )
			} else {
				totalFiles = append( totalFiles, f.Path + "/" )
			}
		}
	}

	return totalFiles, nil


}
