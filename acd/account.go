// Copyright (c) 2015 Serge Gebhardt. All rights reserved.
//
// Use of this source code is governed by the ISC
// license that can be found in the LICENSE file.

package acd

import "net/http"

// AccountService provides access to the account related functions
// in the Amazon Cloud Drive API.
//
// See: https://developer.amazon.com/public/apis/experience/cloud-drive/content/account
type AccountService struct {
	client *Client
}

// AccountInfo represents information about an Amazon Cloud Drive account.
type AccountInfo struct {
	TermsOfUse *string `json:"termsOfUse"`
	Status     *string `json:"status"`
}

func (s *AccountService) GetInfo() (*AccountInfo, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "account/info", nil)
	if err != nil {
		return nil, nil, err
	}

	accountInfo := new(AccountInfo)
	resp, err := s.client.Do(req, accountInfo)
	if err != nil {
		return nil, resp, err
	}

	return accountInfo, resp, err
}
