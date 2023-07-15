package order_brc20_service

import (
	"errors"
	"fmt"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
)

func getWhitelistCount(address, ip string, whitelistType model.WhitelistType) (int, error) {
	var (
		entity *model.WhitelistModel
	)

	entity, _ = mongo_service.FindWhitelistModelByIpAndType(ip, whitelistType)
	if entity != nil {
		return 0, errors.New("the address of this ip had claimed")
	}

	entity, _ = mongo_service.FindWhitelistModelByAddressAndType(address, whitelistType)
	if entity == nil || entity.Id == 0 {
		return 0, errors.New("not in whitelist")
	}
	if entity.WhiteUseState == model.WhiteUseStateYes {
		return 0, nil
	}
	return 1, nil
}

func setWhitelist(address string, whitelistType model.WhitelistType) {
	var (
		entity *model.WhitelistModel
		err    error
	)

	entity, _ = mongo_service.FindWhitelistModelByAddressAndType(address, whitelistType)
	if entity != nil {
		return
	}
	entity = &model.WhitelistModel{
		AddressId:     fmt.Sprintf("%s_%d", address, whitelistType),
		Address:       address,
		IP:            "",
		WhitelistType: whitelistType,
		WhiteUseState: model.WhiteUseStateNo,
		Timestamp:     tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetWhitelistModel(entity)
	if err != nil {
		return
	}
	return
}
