// Copyright 2017, Jonsen Yang.  All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/domainr/whois"
)

func Whois(query string) (body []byte, err error) {
	req, err := whois.NewRequest(query)
	if err != nil {
		return nil, err
	}

	res, err := whois.DefaultClient.Fetch(req)
	if err != nil {
		return nil, err
	}

	return res.Text()
}
