package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
)

type Client struct {
	httpC http.Client
}

func Connect() (*Client, error) {
	client := &Client{httpC: http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return Dial()
			},
		},
	}}
	if err := client.Ping(); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) Ping() error {
	if c.sendRequest(PingPath) != nil {
		return errors.New("ping failed")
	}
	return nil
}

func (c *Client) sendRequest(path string) error {
	resp, err := c.httpC.Get("http://xx/" + path)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var r Response
		json.NewDecoder(resp.Body).Decode(&r)
		return errors.New(r.Error)
	}
	return nil
}

func (c *Client) Show() error {
	return c.sendRequest(ShowPath)
}

func (c *Client) Quit() error {
	return c.sendRequest(QuitPath)
}
