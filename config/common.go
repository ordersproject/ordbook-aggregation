package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

//
func ReadConfig(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	result, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return result, nil
}

//
func ReadJsonConfig(path string, result interface{}) error  {
	if data, err := ReadConfig(path); err != nil {
		return err
	}else {
		err := json.Unmarshal(data, result)
		if err != nil {
			log.Println("Parse config err: " + err.Error())
			return err
		}
		return nil
	}
}
