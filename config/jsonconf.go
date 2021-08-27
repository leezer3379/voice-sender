package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var (
	configPath = "./etc/conf.json"
	wl JsonConf
)


type JsonConf struct {
	WLs []WL   `json:"whitelist"`
	Tk  string `json:"token"`
	Ups []Up   `json:"update"`
}

type Up struct {
	InstanceId string `json:"instanceid"`
	Count      int64  `json:"count"`
}

type WL struct {
	InstanceId string `json:"instanceid"`
	STime      int64  `json:"stime"`
}

func LoadJsonConfig()(config *JsonConf){
	data,err:=ioutil.ReadFile(configPath)
	if err!=nil{
		log.Fatal(err)
	}
	config=&JsonConf{}
	err=json.Unmarshal(data,&config)
	if err!=nil{
		log.Fatal(err)
	}
	return config
}
func SaveJsonConfig(config *JsonConf){

	data,err:=json.Marshal(config)
	if err!=nil{
		log.Fatal(err)
	}
	err=ioutil.WriteFile(configPath,data,0660)
	if err!=nil{
		log.Fatal(err)
	}

}
