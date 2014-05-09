package twitterstream

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"github.com/mreiferson/go-httpclient"
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
	onLine          func([]byte)
	stopChan        chan bool
}

func NewClient(ConsumerKey, ConsumerSecret, Token, TokenSecret string) *Client {
	c := &Client{
		ConsumerKey:     ConsumerKey,
		ConsumerSecret:  ConsumerSecret,
		Token:           Token,
		TokenSecret:     TokenSecret,
		GzipCompression: false,
		stopChan:        make(chan bool),
	}

	return c
}

func (c *Client) Filter(method string, userParams map[string]string) error {
	return c.Do(method, "https://stream.twitter.com/1.1/statuses/filter.json", userParams)
}

func (c *Client) Sample(method string, userParams map[string]string) error {
	return c.Do(method, "https://stream.twitter.com/1.1/statuses/sample.json", userParams)
}

func (c *Client) Firehose(method string, userParams map[string]string) error {
	return c.Do(method, "https://stream.twitter.com/1.1/statuses/firehose.json", userParams)
}

func (c *Client) Userstream(method string, userParams map[string]string) error {
	return c.Do(method, "https://userstream.twitter.com/1.1/user.json", userParams)
}

func (c *Client) Sitestream(method string, userParams map[string]string) error {
	return c.Do(method, "https://sitestream.twitter.com/1.1/site.json", userParams)
}

func (c *Client) OnLine(handler func(data []byte)) {
	c.onLine = handler
}

func (c *Client) Do(method string, url string, userParams map[string]string) error {

	consumer := oauth.NewConsumer(
		c.ConsumerKey,
		c.ConsumerSecret,
		oauth.ServiceProvider{},
	)

	transport := &httpclient.Transport{}
	defer transport.Close()

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
		return errors.New("support method is GET and POST")
	}

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	var reader io.Reader = resp.Body

	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(reader)
		if err != nil {
			return err
		}
	}
	bufreader := bufio.NewReader(reader)

	go func() {
		<-c.stopChan
		transport.CancelRequest(resp.Request)
	}()

	for {
		line, err := bufreader.ReadBytes('\r')
		if err != nil {
			return err
		}
		line = bytes.TrimSpace(line)
		// empty line
		if len(line) < 1 {
			continue
		}
		if c.onLine != nil {
			c.onLine(line)
		}
	}
}

func (c *Client) Stop() {
	c.stopChan <- true
}
