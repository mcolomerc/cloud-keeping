package client

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/martian/log"
)

type HTTPS struct {
	Client   http.Client
	Bearer   string
	Endpoint string
}

type BasicAuth struct {
	Username string
	Password string
}

type HTTPSClient interface {
	Get() (interface{}, error)
	Post(body []byte) (interface{}, error)
}

func NewHTTPS(url string, username string, password string) *HTTPS {
	client := &http.Client{}
	user := username + ":" + password
	bearer := b64.StdEncoding.EncodeToString([]byte(user))

	return &HTTPS{
		Client:   *client,
		Bearer:   bearer,
		Endpoint: url,
	}
}

func (c *HTTPS) Get() (interface{}, error) {
	// Get request to the endpoint
	req, err := http.NewRequest(http.MethodGet, c.Endpoint, nil)
	if err != nil {
		log.Errorf("Error building POST: " + err.Error())
		return nil, err
	}
	return c.build(req)
}

func (c *HTTPS) Post(requestBody []byte) (interface{}, error) {
	req, err := http.NewRequest(http.MethodPost, c.Endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Errorf("Error building POST: " + err.Error())
		return nil, err
	}
	return c.build(req)
}

// Build request - Client Do
func (c *HTTPS) build(req *http.Request) (interface{}, error) {
	req.Header.Set("Authorization", "Basic "+c.Bearer)
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	res, err := c.Client.Do(req)
	if err != nil {
		log.Errorf("Rest client: error making http request: " + err.Error())
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		errorString := fmt.Sprintf("Rest client:: %d - %s : %v \n", res.StatusCode, req.Method, req.URL)
		return nil, errors.New(errorString)
	}
	resBody, err := io.ReadAll(res.Body)
	log.Infof(string(resBody))
	if err != nil {
		log.Errorf("HTTP client:: could not read response body: " + err.Error())
		return nil, err
	}

	var result interface{}
	json.Unmarshal([]byte(resBody), &result)
	return result, nil
}
