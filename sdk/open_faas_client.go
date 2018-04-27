package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openfaas/faas/gateway/requests"
)

type OpenFaaSClient struct {
	url        string
	httpClient *http.Client
}

func buildUrl(baseUrl string, function string) string {
	if len(function) == 0 {
		return fmt.Sprintf("%s/system/functions", baseUrl)
	} else {
		return fmt.Sprintf("%s/function/%s", baseUrl, function)
	}
}

func (client *OpenFaaSClient) FetchFunctions() ([]requests.Function, error) {
	req, _ := http.NewRequest(http.MethodGet, buildUrl(client.url, ""), nil)
	req.Close = true

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var functions []requests.Function
	if err := json.Unmarshal(bytesOut, &functions); err != nil {
		return nil, err
	} else {
		return functions, nil
	}
}

func (client *OpenFaaSClient) InvokeFunction(function string, message []byte) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodPost, buildUrl(client.url, function), bytes.NewReader(message))
	req.Close = true
	defer req.Body.Close()

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	} else {
		return bytesOut, nil
	}

}
