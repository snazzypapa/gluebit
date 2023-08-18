package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	errwrp "github.com/pkg/errors"
	"golang.org/x/net/publicsuffix"
)

const (
	ResponseBodyOK   = "Ok."
	ResponseBodyFAIL = "Fails."
)

var (
	ErrBadResponse      = errors.New("bad response")
	ErrLoginfailed      = errors.New("login failed")
	ErrAddTorrnetfailed = errors.New("add torrnet failed")
)

var defaultTimeout = 1 * time.Second

// Optional parameters when sending HTTP requests
type Optional map[string]any

// StringField returns a map of string representations of all the values in the Optional struct.
func (opt Optional) StringField() map[string]string {
	m := make(map[string]string)
	for k, v := range opt {
		m[k] = fmt.Sprintf("%v", v)
	}
	return m
}

// Client is used to interact with the qBittorrent API.
// It holds the http.Client and the URL of the qBittorrent server
// along with a login cookie after authorizing.
type Client struct {
	*http.Client
	URL string
}

// NewClient creates a new Client for interacting with the qBittorrent API.
func NewClient(url string, username string, password string) (*Client, error) {

	// ensure url ends with "/"
	if url[len(url)-1:] != "/" {
		url += "/"
	}

	// create cookie jar
	cliJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := &Client{
		&http.Client{
			Jar: cliJar,
		},
		url + "api/v2/",
	}

	err := client.Login(username, password)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// postXwwwFormUrlencoded sends a POST request to the specified endpoint
// with the given options encoded as x-www-form-urlencoded.
// Returns the http.Response object and an error if any occurred.
func (c *Client) postXwwwFormUrlencoded(endpoint string, opts Optional) (*http.Response, error) {
	values := url.Values{}
	for k, v := range opts.StringField() {
		values.Set(k, v)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+endpoint, bytes.NewBufferString(values.Encode()))
	if err != nil {
		return nil, errwrp.Wrap(err, "error creating request")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	if err != nil {
		return nil, errwrp.Wrap(err, "failed to perform request")
	}
	return resp, nil
}

// Login logs in the client with the given username and password.
// It returns an error if the login fails.
func (c *Client) Login(username, password string) error {
	opts := Optional{
		"username": username,
		"password": password,
	}
	resp, err := c.postXwwwFormUrlencoded("auth/login", opts)
	err = RespOk(resp, err)
	if err != nil {
		return err
	}
	if err = RespBodyOk(resp.Body, ErrLoginfailed); err != nil {
		return err
	}
	// add the cookie to cookie jar to authenticate later requests
	if cookies := resp.Cookies(); len(cookies) > 0 {
		u, err := url.Parse(c.URL)
		if err != nil {
			return errwrp.Wrap(err, "parse url error")
		}
		u.Path = ""
		c.Jar.SetCookies(u, cookies)
	}
	return nil
}

// GetPreferences retrieves the preferences of the qBittorrent app.
// It returns a Preferences struct and an error if any occurred.
func (c *Client) GetPreferences() (Preferences, error) {
	var prefs Preferences
	resp, err := c.postXwwwFormUrlencoded("app/preferences", nil)
	err = RespOk(resp, err)
	if err != nil {
		return prefs, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return prefs, err
	}
	err = json.Unmarshal(b, &prefs)
	log.Println(prefs)
	return prefs, err
}

// SetPreferences sets the preferences in the qBittorrent app.
// It takes a Preferences struct as input and returns an error if any.
func (c *Client) SetPreferences(pref Preferences) error {
	b, err := json.Marshal(pref)
	if err != nil {
		return err
	}
	opt := Optional{
		"json": string(b),
	}
	resp, err := c.postXwwwFormUrlencoded("app/setPreferences", opt)
	err = RespOk(resp, err)
	if err != nil {
		return err
	}
	ignrBody(resp.Body)
	return nil
}

// RespOk checks if the HTTP response is successful
// (status code 200 OK) and returns an error if not.
func RespOk(resp *http.Response, err error) error {
	switch {
	case err != nil:
		return err
	case resp.Status != "200 OK": // check for correct status code
		return errwrp.Errorf("%v: %s", ErrBadResponse, resp.Status)
	default:
		return nil
	}
}

// RespBodyOk checks if the response body is equal to ResponseBodyOK constant.
// If the response body is not equal to ResponseBodyOK, it returns the bizErr.
// If there is an error while reading the response body, it returns the error.
// Otherwise, it returns nil.
func RespBodyOk(body io.ReadCloser, bizErr error) error {
	defer body.Close()
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if string(b) != ResponseBodyOK {
		return bizErr
	}
	return nil
}

// ignrBody reads the response body and discards it.
// This is useful when the response body is not needed, but the response headers are.
func ignrBody(body io.ReadCloser) error {
	_, err := io.Copy(io.Discard, body)
	return err
}
