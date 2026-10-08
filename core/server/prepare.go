package server

import (
	"encoding/json"
	"errors"
	"github.com/hunterhug/fafacms/core/config"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/util/rdb"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func InitConfig(configFilePath string) error {
	c := new(config.Config)
	if configFilePath == "" {
		return errors.New("config file empty")
	}

	raw, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, c)
	if err != nil {
		return err
	}

	c.DbConfig.Prefix = "fafacms_"
	config.FaFaConfig = c
	return nil
}

func InitYamlConfig(configFilePath string) error {
	c := new(config.Config)
	if configFilePath == "" {
		return errors.New("config file empty")
	}

	raw, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(raw, c)
	if err != nil {
		return err
	}

	c.DbConfig.Prefix = "fafacms_"
	c.DbConfig.DriverName = rdb.MYSQL
	config.FaFaConfig = c
	return nil
}

func InitRdb(dbConfig rdb.MyDbConfig) error {
	db, err := rdb.NewDb(dbConfig)
	if err != nil {
		return err
	}

	model.FaFaRdb = db
	return nil
}
