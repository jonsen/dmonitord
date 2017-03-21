// Copyright 2017, Jonsen Yang.  All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"regexp"
	"time"
)

type domainInfo struct {
	Domain      string
	Status      []string
	Registrant  string
	Email       string
	Create      time.Time
	Expiry      time.Time
	NameServers []string
}

var (
	domain     = regexp.MustCompile("Domain ?(N|n)ame: (.*)")
	status     = regexp.MustCompile("(Status|Domain Status): (.*)")
	expiry     = regexp.MustCompile(`(Expiration Time|Expiration Date|Registry Expiry Date): (.*)`)
	dns        = regexp.MustCompile("Name Server: (.*)")
	registrant = regexp.MustCompilePOSIX("Registrant Name: (.*)")
	create     = regexp.MustCompile("(Registration Time|Creation Date): (.*)")
	email      = regexp.MustCompile("(Registrant Email|Email): (.*)")

	layout = []string{"02-Jan-2006", "2006-01-02", "2006-01-02T15:04:05Z", "2006-01-02 15:04:05"}
)

func parseTime(t string) (tm time.Time, err error) {

	// .me 2013-01-26T04:07:01Z
	// .com 2008-01-24T19:00:21Z
	// .net 2018-01-10T00:00:00Z
	// .cn 2017-04-28 15:10:46
	for _, l := range layout {
		tm, err = time.Parse(l, t)
		if err == nil {
			break
		}
	}

	return
}

func parse(respone []byte) (info domainInfo) {

	dm := domain.Find(respone)
	if len(dm) > 0 {
		ss := bytes.Split(dm, []byte(":"))
		if len(ss) == 2 {
			info.Domain = string(bytes.TrimSpace(ss[1]))
		}
	}

	sall := status.FindAll(respone, -1)
	for _, v := range sall {
		ss := bytes.SplitN(v, []byte(":"), 2)
		if len(ss) == 2 {
			info.Status = append(info.Status, string(bytes.TrimSpace(ss[1])))
		}
	}

	expir := expiry.Find(respone)
	if len(expir) > 0 {
		ss := bytes.SplitN(expir, []byte(":"), 2)
		if len(ss) == 2 {
			t, err := parseTime(string(bytes.TrimSpace(ss[1])))
			if err == nil {
				info.Expiry = t
			}
		}
	}

	sdns := dns.FindAll(respone, -1)
	for _, v := range sdns {
		ss := bytes.Split(v, []byte(":"))
		if len(ss) == 2 {
			info.NameServers = append(info.NameServers, string(bytes.TrimSpace(ss[1])))
		}

	}
	registr := registrant.Find(respone)
	if len(registr) > 0 {
		ss := bytes.Split(registr, []byte(":"))
		if len(ss) == 2 {
			info.Registrant = string(bytes.TrimSpace(ss[1]))
		}
	}

	mail := email.Find(respone)
	if len(mail) > 0 {
		ss := bytes.Split(mail, []byte(":"))
		if len(ss) == 2 {
			info.Email = string(bytes.TrimSpace(ss[1]))
		}
	}

	registrat := create.Find(respone)
	if len(registrat) > 0 {
		ss := bytes.SplitN(registrat, []byte(":"), 2)
		t, err := parseTime(string(bytes.TrimSpace(ss[1])))
		if err == nil {
			info.Create = t
		}
	}

	return
}
