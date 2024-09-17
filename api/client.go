package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Post(url string, requestBody interface{}) ([]byte, error) {
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed with status code: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
