package order_brc20_service

import (
	"errors"
	"ordbook-aggregation/config"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
)

func FetchAddressNotification(req *request.OrderNotificationFetchReq) (*respond.NotificationResponse, error) {
	var (
		total      int64 = 0
		entityList []*model.OrderNotificationModel
		list       []*respond.NotificationItem = make([]*respond.NotificationItem, 0)
	)

	if req.Address == "" {
		return nil, errors.New("address is empty")
	}

	total, _ = mongo_service.CountOrderNotificationModelList(req.Address)
	entityList, _ = mongo_service.FindOrderNotificationModelList(req.Address)
	for _, v := range entityList {
		if v.NotificationCount == 0 {
			total--
			continue
		}
		title := ""
		desc := ""
		switch v.NotificationType {
		case model.NotificationTypePoolUsed:
			title = config.NotificationTitlePoolUsed
			desc = config.NotificationDescPoolUsed
			break
		case model.NotificationTypeBidInvalid:
			title = config.NotificationTitleBidInvalid
			desc = config.NotificationDescBidInvalid
			break
		case model.NotificationTypeOrderFinish:
			title = config.NotificationTitleOrderFinish
			desc = config.NotificationDescOrderFinish
			break
		}
		list = append(list, &respond.NotificationItem{
			NotificationType:  v.NotificationType,
			NotificationCount: v.NotificationCount,
			NotificationTitle: title,
			NotificationDesc:  desc,
		})
	}
	return &respond.NotificationResponse{
		Total:   total,
		Results: list,
	}, nil
}

func ClearAllNotification(req *request.OrderNotificationFetchReq) (string, error) {
	clearNotification(req.Address, model.NotificationType(req.NotificationType))
	return "success", nil
}
