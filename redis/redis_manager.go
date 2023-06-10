package redis

import (
	"errors"
	"fmt"
	"ordbook-aggregation/tool"
	"reflect"
	"strconv"
	"strings"
)

var (
	cacheTime int = 60
)

func SetRedisUtxoInfo(utxoType_, utxoId string, sortIndex int) (string, error) {
	key := fmt.Sprintf("%s%s%s", CacheGetUtxo_, utxoType_, utxoId)
	//GetRedisKeyItemMap().Set(key, true)
	v, err := GetRedisManager().Set(redisDbUtxo, key, sortIndex, cacheTime)
	if tool.AnyToStr(v) != "OK" {
		return "", errors.New("Has been locked. ")
	}
	return tool.AnyToStr(v), err
}

func GetRedisUtxoInfo(utxoType_, utxoId string) (string, error)  {
	key := fmt.Sprintf("%s%s%s", CacheGetUtxo_, utxoType_, utxoId)
	v, err := GetRedisManager().Get(redisDbUtxo, key)
	if err == nil {
		if value, ok := v.([]byte); ok {
			info := strings.Trim(string(value), "\"")
			if info != "" {
				return info, nil
			}
		}
	}else {
		return "", err
	}
	return "", nil
}

func UnSetUtxoInfo(utxoType_, utxoId string) error {
	key := fmt.Sprintf("%s%s%s", CacheGetUtxo_, utxoType_, utxoId)
	_, err := GetRedisManager().Get(redisDbUtxo, key)
	if err == nil {
		_, err = GetRedisManager().Delete(redisDbUtxo, key)
	}else {
		//GetRedisKeyItemMap().Deleted(key)
	}

	//GetRedisKeyItemMap().Deleted(key)
	//_, err := GetRedisManager().Delete(key)
	return err
}


func GetUtxoInfoKeyList(keyPrefix string) ([]string, error) {
	result, err := GetRedisManager().GetList(redisDbUtxo, fmt.Sprintf("%s*", keyPrefix))
	//result, err := GetRedisManager().GetList(fmt.Sprintf("*"))
	if err != nil {
		return nil, err
	}
	//fmt.Println(result)
	keyList := make([]string, 0)
	if reflect.TypeOf(result).Kind() == reflect.Slice {
		valList := result.([]interface{})
		if len(valList) == 0 {
			return keyList, nil
		}
		for _, v := range valList {
			keyList = append(keyList, string(v.([]byte)))
		}
	}
	return keyList, nil
}

func GetUtxoInfoKeyValueList(keyPrefix string) ([]string, []int, error) {
	result, err := GetRedisManager().GetList(redisDbUtxo, fmt.Sprintf("%s*", keyPrefix))
	//result, err := GetRedisManager().GetList(fmt.Sprintf("*"))
	if err != nil {
		return nil, nil, err
	}
	//fmt.Println(result)
	keyList := make([]string, 0)
	valueList := make([]int, 0)
	if reflect.TypeOf(result).Kind() == reflect.Slice {
		valList := result.([]interface{})
		if len(valList) == 0 {
			return keyList, valueList, nil
		}
		for _, va := range valList {
			vStr := string(va.([]byte))
			v, _ := GetRedisManager().Get(redisDbUtxo, vStr)
			if v != nil {
				if value, ok := v.([]byte); ok {
					info := strings.Trim(string(value), "\"")
					if info != "" {
						infoInt, _ := strconv.ParseInt(info, 10, 64)
						valueList = append(valueList, int(infoInt))
					}
				}
			}
			keyList = append(keyList, strings.ReplaceAll(vStr, keyPrefix, ""))
		}
	}
	return keyList, valueList, nil
}
