package twitterstream

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"github.com/mrjones/oauth"
	"io"
	"net/http"
)

type Client struct {
	ConsumerKey     string
	ConsumerSecret  string
	Token           string
	TokenSecret     string
	GzipCompression bool
}

type Connection struct {
	response *http.Response
	reader   *bufio.Reader
	consumer *oauth.Consumer
}

func (c *Client) Filter(method string, userParams map[string]string) (*Connection, error) {
	return c.Do(method, "https://stream.twitter.com/1.1/statuses/filter.json", userParams)
}

func (c *Client) Sample(method string, userParams map[string]string) (*Connection, error) {
	return c.Do(method, "https://stream.twitter.com/1.1/statuses/sample.json", userParams)
}

func (c *Client) Firehose(method string, userParams map[string]string) (*Connection, error) {
	return c.Do(method, "https://stream.twitter.com/1.1/statuses/firehose.json", userParams)
}

func (c *Client) Userstream(method string, userParams map[string]string) (*Connection, error) {
	return c.Do(method, "https://userstream.twitter.com/1.1/user.json", userParams)
}

func (c *Client) Sitestream(method string, userParams map[string]string) (*Connection, error) {
	return c.Do(method, "https://sitestream.twitter.com/1.1/site.json", userParams)
}

func (c *Client) Do(method string, url string, userParams map[string]string) (*Connection, error) {

	consumer := oauth.NewConsumer(
		c.ConsumerKey,
		c.ConsumerSecret,
		oauth.ServiceProvider{},
	)

	transport := &http.Transport{}

	client := &http.Client{Transport: transport}
	consumer.HttpClient = client

	accesstoken := &oauth.AccessToken{
		Token:  c.Token,
		Secret: c.TokenSecret,
	}

	additionalHeaders := make(map[string][]string)

	if c.GzipCompression {
		additionalHeaders["Accept-Encoding"] = []string{"deflate, gzip"}
	}

	consumer.AdditionalHeaders = additionalHeaders

	var resp *http.Response
	var err error
	switch method {
	case "GET":
		resp, err = consumer.Get(url, userParams, accesstoken)
	case "POST":
		resp, err = consumer.Post(url, userParams, accesstoken)
	default:
		return nil, errors.New("support method is GET and POST")
	}

	if err != nil {
		return nil, err
	}

	return newConnection(consumer, resp)
}

func newConnection(consumer *oauth.Consumer, resp *http.Response) (*Connection, error) {
	var reader io.Reader = resp.Body
	var err error

	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
	}

	conn := &Connection{
		response: resp,
		reader:   bufio.NewReader(reader),
		consumer: consumer,
	}

	return conn, nil
}

func (c *Connection) Next() ([]byte, error) {
	for {
		line, err := c.reader.ReadBytes('\r')
		if err != nil {
			return nil, err
		}
		line = bytes.TrimSpace(line)
		// empty line
		if len(line) < 1 {
			continue
		}
		return line, nil
	}
}

func (c *Connection) Close() error {
	return c.response.Body.Close()
}

func (c *Connection) Stop() {
	c.consumer.HttpClient.(*http.Client).Transport.(*http.Transport).CancelRequest(c.response.Request)
}
