package apiv2

import (
	"fmt"
	"net/http"
	"io/ioutil"
)

func(client *HuggingFaceClient) Covert2( format, what string) error {

	url := fmt.Sprintf("%s/api/%s/%s", baseURL, what, format)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	res, err := client.doRequest( req )
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll( res.Body )
	if err != nil {
		return err
	}
	
	return nil
}
