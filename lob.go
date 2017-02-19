package postcards2diane

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	lobAPIVersion = "2016-06-30"
)

// Confirmation describes a confirmation that a postcard has been sent.
type Confirmation struct {
	ID               string `json:"id"`
	URL              string `json:"url"`
	ExpectedDelivery string `json:"expected_delivery_date"`
}

func NewLobClient(apiKey, configPath string) (c LobClient, err error) {
	f, err := os.Open(configPath)
	if err != nil {
		return c, err
	}
	err = json.NewDecoder(f).Decode(&c.addresses)
	if err != nil {
		return c, err
	}
	c.apiKey = apiKey
	if c.addresses.Addresses == nil {
		c.addresses.Addresses = make(map[string]string)
	}
	return c, nil
}

type LobClient struct {
	apiKey    string
	addresses addressConfig
}

func (c LobClient) Send(p *Postcard, to string) (confirmation Confirmation, err error) {
	toAddress, ok := c.addresses.Addresses[to]
	if !ok {
		return confirmation, fmt.Errorf("unknown `to` address alias %q", to)
	}
	fromAddress, ok := c.addresses.Addresses[c.addresses.From]
	if !ok {
		return confirmation, fmt.Errorf("unknown `from` address alias %q", c.addresses.From)
	}

	// Render the postcard and encode as a png into a buffer.
	pngBytes, err := p.Render()
	if err != nil {
		return confirmation, err
	}

	var multipartBuf bytes.Buffer
	w := multipart.NewWriter(&multipartBuf)
	part, err := w.CreateFormFile("front", "front.png")
	if err != nil {
		return confirmation, err
	}
	_, err = io.Copy(part, bytes.NewReader(pngBytes))
	if err != nil {
		return confirmation, err
	}

	_ = w.WriteField("to", toAddress)
	_ = w.WriteField("from", fromAddress)
	_ = w.WriteField("size", p.size)
	_ = w.WriteField("message", p.message)

	err = w.Close()
	if err != nil {
		return confirmation, err
	}

	err = c.postMultipart(
		"/v1/postcards",
		multipartBuf.Bytes(),
		w.FormDataContentType(),
		&confirmation,
	)
	return confirmation, err
}

func (c LobClient) postMultipart(endpoint string, body []byte, contentType string, respBody interface{}) error {
	url := fmt.Sprintf("https://api.lob.com%s", endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Lob-Version", lobAPIVersion)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "postcards2diane 0.0.1")
	req.SetBasicAuth(c.apiKey, "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// do nothing
	case http.StatusUnauthorized:
		return errors.New("lob API auth failed")
	case http.StatusBadRequest:
		return errors.New("bad request")
	case http.StatusNotFound:
		return errors.New("not found")
	case http.StatusTooManyRequests:
		return errors.New("too many requests")
	default:
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %s\n\n%s", resp.Status, body)
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return err
}

type addressConfig struct {
	// From sets the alias of the default address to send postcards
	// from.
	From string `json:"from,omitempty"`

	// Addresses maps addresses from a human-readable alias to
	// the Lob ID of the address.
	Addresses map[string]string `json:"addresses"`
}
