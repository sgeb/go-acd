// Copyright (c) 2015 Serge Gebhardt. All rights reserved.
//
// Use of this source code is governed by the ISC
// license that can be found in the LICENSE file.

package acd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccount_getInfo(t *testing.T) {
	r := *NewMockResponseOkString(`{ "termsOfUse": "1.0.0", "status": "ACTIVE" }`)
	c := NewMockClient(r)

	info, _, err := c.Account.GetInfo()

	assert.NoError(t, err)
	assert.Equal(t, "ACTIVE", *info.Status)
	assert.Equal(t, "1.0.0", *info.TermsOfUse)
}

func TestAccount_getQuota(t *testing.T) {
	r := *NewMockResponseOkString(`
{
"quota": 5368709120,
"lastCalculated": "2014-08-13T23:01:47.479Z",
"available": 4069088896
}
	`)
	c := NewMockClient(r)

	quota, _, err := c.Account.GetQuota()

	assert.NoError(t, err)
	assert.Equal(t, "2014-08-13 23:01:47.479 +0000 UTC", quota.LastCalculated.String())
	assert.Equal(t, uint64(5368709120), *quota.Quota)
	assert.Equal(t, uint64(4069088896), *quota.Available)
}
