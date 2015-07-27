package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Client struct {
	url    *url.URL
	client *http.Client
	keyId  string
	key    string
}

func NewClient(uri string, keyId string, key string) (*Client, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	client := &Client{
		url:    parsed,
		client: &http.Client{},
		keyId:  keyId,
		key:    key,
	}

	return client, nil
}

func (c *Client) buildRequest(method, path string) (*http.Request, error) {
	urlString := fmt.Sprintf("%s%s", c.url, path)
	req, err := http.NewRequest(method, urlString, nil)
	if err != nil {
		return nil, err
	}

	values := req.URL.Query()
	values.Add("AWS_SECRET_ACCESS_KEY", c.key)
	values.Add("AWS_ACCESS_KEY_ID", c.keyId)

	req.URL.RawQuery = values.Encode()
	log.Print(req.URL.String())
	return req, nil
}

type listInstancesResponse struct {
	instances []string `json:"instances"`
}

func (c *Client) ListInstances(vpc string) ([]string, error) {
	req, err := c.buildRequest("GET", fmt.Sprintf("/vpc/%s/instances", vpc))
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	instances := new(listInstancesResponse)
	err = json.Unmarshal(body, instances)
	if err != nil {
		return nil, err
	}

	return instances.instances, nil
}

func (c *Client) Terminate(instance string) error {
	req, err := c.buildRequest("DELETE", fmt.Sprintf("/instances/%s", instance))
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

type cloneResponse struct {
	Id string `json:"id"`
}

func (c *Client) Clone(instance string) (string, error) {
	req, err := c.buildRequest("PUT", fmt.Sprintf("/instances/%s/clone", instance))
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	newInstance := new(cloneResponse)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(body, newInstance)
	if err != nil {
		return "", err
	}

	return newInstance.Id, nil
}
