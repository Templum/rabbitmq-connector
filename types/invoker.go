package types

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Invoker struct {
	PrintResponse bool
	Client        *http.Client
	GatewayURL    string
}

func (i *Invoker) Invoke(topicMap *TopicMap, topic string, message *[]byte) {
	if len(*message) > 0 {

		matchedFunctions := topicMap.Match(topic)
		for _, matchedFunction := range matchedFunctions {

			log.Printf("Invoke function: %s", matchedFunction)

			gwURL := fmt.Sprintf("%s/function/%s", i.GatewayURL, matchedFunction)
			reader := bytes.NewReader(*message)

			body, statusCode, doErr := invokefunction(i.Client, gwURL, reader)

			if doErr != nil {
				log.Printf("Unable to invoke from %s, error: %s\n", matchedFunction, doErr)
				return
			}

			printBody := false
			stringOutput := ""

			if body != nil && i.PrintResponse {
				stringOutput = string(*body)
				printBody = true
			}

			if printBody {
				log.Printf("Response [%d] from %s %s", statusCode, matchedFunction, stringOutput)

			} else {
				log.Printf("Response [%d] from %s", statusCode, matchedFunction)
			}
		}
	}
}

func invokefunction(c *http.Client, gwURL string, reader io.Reader) (*[]byte, int, error) {

	httpReq, _ := http.NewRequest(http.MethodPost, gwURL, reader)
	defer httpReq.Body.Close()

	var body *[]byte
	res, doErr := c.Do(httpReq)
	if doErr != nil {
		return nil, http.StatusServiceUnavailable, doErr
	}
	if res.Body != nil {
		defer res.Body.Close()

		bytesOut, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Printf("Error reading body")
			return nil, http.StatusServiceUnavailable, doErr

		}
		body = &bytesOut
	}

	return body, res.StatusCode, doErr
}
