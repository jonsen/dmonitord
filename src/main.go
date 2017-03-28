// Copyright 2017, Jonsen Yang.  All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	. "github.com/forease/ebase"
	"io"
	"os"
	"strings"
	"time"
)

var (
	smtp     *Smtp
	mailAddr string
	retry    int
)

//
// check rules:
// more than 120 days, check every 30 days.
// more than 60 days, check every 7 days.
// more than 30 days, check every 3 days.
// less than 30 days, check every day.
func scan(fileName string) {
	Log.Debug("start scan domains")
	// get domain from file
	f, err := os.Open(fileName)
	if err != nil {
		Log.Error(err)
		return
	}
	defer f.Close()

	now := time.Now()
	nowUnix := now.Unix()

	mailList := make(map[string]int64)
	count := 0

	buf := bufio.NewReader(f)
	for {

		domain, err := buf.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				break
			}
			Log.Error(err)
			break
		}

		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}

		domain = strings.ToLower(domain)

		extDomain, ok := cache[domain]
		if ok {
			Log.Tracef("get domain %s from cache", domain)
			switch {
			case extDomain.Days > 120 && !extDomain.Last.Before(now.Add(-24*time.Hour*30)):
				fallthrough
			case extDomain.Days >= 60 && !extDomain.Last.Before(now.Add(-24*time.Hour*7)):
				fallthrough
			case extDomain.Days >= 30 && !extDomain.Last.Before(now.Add(-24*time.Hour*3)):
				Log.Debugf("domain %s not need scan. expiry day %d, last scan %s.",
					domain, extDomain.Days, extDomain.Last)
				continue
			default:
			}

		}

		var ss []byte

		for i := 0; i < retry; i++ {
			ss, err = Whois(domain)
			if err == nil {
				break
			}
		}

		if err != nil {
			Log.Errorf("whois domain %s error %s", domain, err)
			continue
		}

		d := parse(ss)
		if d.Domain == "" {
			Log.Debugf("not found domain for %s", domain)
			if ok {
				delete(cache, domain)
			}
			continue
		}

		count++

		// Check expiry time
		exDay := (d.Expiry.Unix() - nowUnix) / 86400

		// insert or update
		if ok {
			extDomain.Last = now
			extDomain.Days = exDay
			extDomain.Create = d.Create
			extDomain.Expiry = d.Expiry

			err := extDomain.Update()
			if err != nil {
				Log.Errorf("update domain %s error %s", domain, err)
			}
		} else {
			dc := &domainCache{
				Name:   domain,
				Create: d.Create,
				Expiry: d.Expiry,
				Last:   now,
				Days:   exDay,
			}

			cache[domain] = dc

			err := dc.Insert()
			if err != nil {
				Log.Errorf("insert domain %s error %s", domain, err)
			}
		}

		mailList[domain] = exDay

		time.Sleep(time.Second * 3)
	}

	Log.Debugf("end scan domains. with %d domains, using %s.", count, time.Now().Sub(now))

	// Send email
	total := len(mailList)
	if len(mailList) > 0 {
		subject := fmt.Sprintf("%d domains status report at %s", total,
			now.Format("2006-01-02 15:04:05"))

		content := "Hi,\n\n"

		list := NewMapSorter(mailList)
		for sn, data := range list {
			date := cache[data.Key].Expiry.Format("2006-01-02")
			content += fmt.Sprintf("   %d. %s will expire on %d (%s) days.\n",
				sn+1, data.Key, data.Val, date)
		}

		content += "\n\n    report by dmonitord\n\n"

		content += "\r\n\r\n"

		//  MailSender(subject, content, to, cc, bcc string)
		if err := smtp.MailSender(subject, content, mailAddr, "", ""); err != nil {
			Log.Error(err)
		}
	}

	Log.Debugf("sended mail to %s, using %s.", mailAddr, time.Now().Sub(now))

}

func main() {
	EbaseInit()

	model, err := NewDefaultModels()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer model.Close()

	model.Orm.Sync2(new(domainCache))

	dFile, err := Config.String("common.dfile", "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = os.Stat(dFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// initial smtp for send mail
	smtp = NewSmtp()

	mailAddr, _ = Config.String("common.adminer", "")
	retry, _ = Config.Int("common.retry", 3)
	cHour, _ := Config.Int("check.hour", 5)
	cMinute, _ := Config.Int("check.minute", 0)
	cSecond, _ := Config.Int("check.second", 0)

	Log.Info("dmonitord running now...")

	// load cache from database
	loadCache()

	go func() {

		for {

			time.Sleep(time.Second)
			now := time.Now()

			hour := now.Hour()
			minute := now.Minute()
			second := now.Second()

			if hour == cHour && minute == cMinute && second == cSecond {
				scan(dFile)
			}
		}

	}()

	SignalHandle(map[string]interface{}{})
}
