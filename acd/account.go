// Copyright (c) 2015 Serge Gebhardt. All rights reserved.
//
// Use of this source code is governed by the ISC
// license that can be found in the LICENSE file.

package acd

import (
	"net/http"
	"time"
)

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

// Provides information about the current user account like the status and the
// accepted “Terms Of Use”.
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

// AccountQuota represents information about the account quotas.
type AccountQuota struct {
	Quota          *uint64    `json:"quota"`
	LastCalculated *time.Time `json:"lastCalculated"`
	Available      *uint64    `json:"available"`
}

// Gets account quota and storage availability information.
func (s *AccountService) GetQuota() (*AccountQuota, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "account/quota", nil)
	if err != nil {
		return nil, nil, err
	}

	accountQuota := new(AccountQuota)
	resp, err := s.client.Do(req, accountQuota)
	if err != nil {
		return nil, resp, err
	}

	return accountQuota, resp, err
}
