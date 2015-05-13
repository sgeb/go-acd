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
