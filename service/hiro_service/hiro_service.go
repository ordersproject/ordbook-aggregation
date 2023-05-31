package hiro_service

import (
	"errors"
	"fmt"
	"ordbook-aggregation/config"
	"ordbook-aggregation/tool"
)

// output: txId:index
func GetOutInscription(output string) (*HiroInscription, error){
	var (
		url        string
		result        string
		resp        *HiroResp
		data        []*HiroInscription = make([]*HiroInscription, 0)
		err        error
	)

	url = fmt.Sprintf("%s/ordinals/v1/inscriptions?output=%s", config.HiroDomain, output)
	result, err = tool.GetUrlForSingle(url)
	if err != nil {
		return nil, err
	}

	if err = tool.JsonToObject(result, &resp) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}

	if err = tool.JsonToAny(resp.Results, &data) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}
	if len(data) == 0 {
		return nil, errors.New("No Inscription. ")
	}

	return data[0], nil
}

func GetInscriptionContent(inscriptionId string) (interface{}, error){
	var (
		url        string
		result        string
		err        error
	)

	url = fmt.Sprintf("%s/ordinals/v1/inscriptions/%s/content", config.HiroDomain, inscriptionId)
	result, err = tool.GetUrlForSingle(url)
	if err != nil {
		return nil, err
	}
	fmt.Println(result)
	return result, nil
}