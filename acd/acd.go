// Copyright (c) 2015 Serge Gebhardt. All rights reserved.
//
// Use of this source code is governed by the ISC
// license that can be found in the LICENSE file.

package acd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

const (
	libraryVersion = "0.1"
	defaultBaseURL = "https://drive.amazonaws.com/drive/v1/"
	userAgent      = "go-acd/" + libraryVersion
)

// A Client manages communication with the Amazon Cloud Drive API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// Base URL for API requests. Defaults to the public Amazon Cloud Drive API.
	// BaseURL should always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the API.
	UserAgent string

	// Services used for talking to different parts of the API.
	Account *AccountService
}

// NewClient returns a new Amazon Cloud Drive API client. If a nil httpClient is
// provided, http.DefaultClient will be used. To use API methods which require
// authentication, provide an http.Client that will perform the authentication
// for you (such as that provided by the golang.org/x/oauth2 library).
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent}
	c.Account = &AccountService{client: c}
	return c
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	//	req.Header.Add("Accept", mediaTypeV3)
	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}
	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	//	err = CheckResponse(resp)
	//	if err != nil {
	//		// even though there was an error, we still return the response
	//		// in case the caller wants to inspect it further
	//		return resp, err
	//	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}
	return resp, err
}
