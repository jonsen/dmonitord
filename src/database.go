package main

import (
	. "github.com/forease/ebase"
	"time"
)

type domainCache struct {
	Id     int64     `xorm:"int(10) pk autoincr"`
	Name   string    `xorm:"Varchar(255) index"`
	Create time.Time `xorm:"DateTime"`
	Expiry time.Time `xorm:"DateTime"`
	Last   time.Time `xorm:"DateTime"`
	Days   int64     `xorm:"int"`
}

var (
	cache = make(map[string]*domainCache)
)

func (dc *domainCache) Update() (err error) {
	session := Dbh.Orm.NewSession()
	defer session.Close()

	_, err = session.AllCols().Id(dc.Id).Update(dc)

	return
}

func (dc *domainCache) Fetch() (err error) {
	_, err = Dbh.Orm.Id(dc.Id).Get(dc)

	return
}

func (dc *domainCache) FetchByName() (err error) {
	_, err = Dbh.Orm.Where("name = ?", dc.Name).Get(dc)

	return
}

func (dc *domainCache) Insert() (err error) {
	session := Dbh.Orm.NewSession()
	defer session.Close()
	_, err = session.InsertOne(dc)

	return
}

func (dc *domainCache) Delete() (err error) {
	session := Dbh.Orm.NewSession()
	defer session.Close()
	_, err = session.Delete(dc)

	return
}

func (dc *domainCache) FetchAll() (all []*domainCache, err error) {
	err = Dbh.Orm.Find(&all)

	return
}

func loadCache() {
	dc := new(domainCache)
	all, err := dc.FetchAll()

	if err != nil {
		Log.Error("load domain cache error", err)
		return
	}

	for _, v := range all {

		cache[v.Name] = v
	}

	Log.Info("load cache done.")
}
