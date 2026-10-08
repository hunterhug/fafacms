package model

import (
	"github.com/hunterhug/fafacms/core/util/rdb"
	log "github.com/hunterhug/golog"
)

var FaFaRdb *rdb.MyDb

func CreateTable(tables []interface{}) {
	for _, table := range tables {
		ok, err := FaFaRdb.IsTableExist(table)
		if err != nil {
			log.Warnf("CreateTable err: %s", err.Error())
			continue
		}
		if !ok {
			//  change the Charset
			sess := FaFaRdb.Client.NewSession()
			sess.Charset("utf8mb4")
			err = sess.CreateTable(table)
			if err != nil {
				sess.Close()
				log.Warnf("CreateTable err: %s", err.Error())
				continue
			}

			sess.Close()
		} else {
			sess := FaFaRdb.Client.NewSession()
			err = sess.Sync2(table)
			if err != nil {
				sess.Close()
				log.Warnf("CreateTable err: %s", err.Error())
			}
			sess.Close()
		}

		err = FaFaRdb.Client.CreateIndexes(table)
		if err != nil {
			log.Warnf("CreateTable err: %s", err.Error())
			//continue
		}
		err = FaFaRdb.Client.CreateUniques(table)
		if err != nil {
			log.Warnf("CreateTable err: %s", err.Error())
			continue
		}
	}

	u := new(User)
	u.Name = "admin"
	u.Email = "admin@admin"
	u.NickName = "admin"
	u.Password = "admin"
	u.Status = 1
	u.Vip = 1
	u.ShortDescribe = "super man"
	err := u.InsertOne()
	if err != nil {
		log.Warnf("CreateTable err: %s", err.Error())
	}
}
