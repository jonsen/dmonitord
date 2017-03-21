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

type queue struct {
	expiry time.Time
	last   time.Time
	days   int64
}

var (
	cache    = make(map[string]queue)
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
	Log.Debug("start scan domain")
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

		extDomain, ok := cache[domain]
		if ok {
			switch {
			case extDomain.days > 120 && !extDomain.last.Before(now.Add(-24*time.Hour*30)):
				fallthrough
			case extDomain.days >= 60 && !extDomain.last.Before(now.Add(-24*time.Hour*7)):
				fallthrough
			case extDomain.days >= 30 && !extDomain.last.Before(now.Add(-24*time.Hour*3)):
				Log.Debugf("domain %s not need scan. expiry day %d, last scan %s.",
					domain, extDomain.days, extDomain.last)
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
			Log.Errorf("domain %s error %s", domain, err)
			continue
		}

		d := parse(ss)
		count++

		// Check expiry time
		exDay := (d.Expiry.Unix() - nowUnix) / 86400
		if ok {
			extDomain.last = now
			extDomain.days = exDay
		} else {
			cache[domain] = queue{expiry: d.Expiry, last: now, days: exDay}
		}

		mailList[domain] = exDay

		time.Sleep(time.Second * 3)
	}

	// Send email
	total := len(mailList)
	if len(mailList) > 0 {
		subject := fmt.Sprintf("%d domains status report at %s", total,
			now.Format("2006-01-02 15:04:05"))

		content := "Hi,\n\n"

		list := NewMapSorter(mailList)
		for sn, data := range list {
			date := cache[data.Key].expiry.Format("2006-01-02")
			content += fmt.Sprintf("   %d. %s will expired on %d (%s) days.\n",
				sn+1, data.Key, data.Val, date)
		}

		content += "\n\n    report by dmonitord\n\n"

		content += "\r\n\r\n"

		//  MailSender(subject, content, to, cc, bcc string)
		if err := smtp.MailSender(subject, content, mailAddr, "", ""); err != nil {
			Log.Error(err)
		}
	}

	Log.Debugf("end scan domain. with %d domain, using %s.", count, time.Now().Sub(now))

}

func main() {
	EbaseInit()

	/*
		model, err := NewDefaultModels()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer model.Close()
	*/

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

	smtp = NewSmtp()
	mailAddr, _ = Config.String("common.adminer", "")
	retry, _ = Config.Int("common.retry", 3)

	Log.Info("dmonitord running now...")

	go func() {

		for {

			time.Sleep(time.Second)
			now := time.Now()

			hour := now.Hour()
			minute := now.Minute()
			second := now.Second()

			if hour > -1 && second == 0 && minute%5 == 0 {
				scan(dFile)
			}
		}

	}()

	SignalHandle(map[string]interface{}{})
}
