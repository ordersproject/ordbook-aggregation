package order_brc20_service

import (
	"fmt"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
)

func AddNotificationForPoolUsed(address string) {
	addNotification(address, model.NotificationTypePoolUsed)
}
func RemoveNotificationForPoolUsed(address string) {
	removeNotification(address, model.NotificationTypePoolUsed)
}

func AddNotificationForBidInvalid(address string) {
	addNotification(address, model.NotificationTypeBidInvalid)
}
func RemoveNotificationForBidInvalid(address string) {
	removeNotification(address, model.NotificationTypeBidInvalid)
}

func AddNotificationForOrderFinish(address string) {
	addNotification(address, model.NotificationTypeOrderFinish)
}
func RemoveNotificationForOrderFinish(address string) {
	removeNotification(address, model.NotificationTypeOrderFinish)
}

// clear all notification
func clearNotification(address string, notificationType model.NotificationType) {
	fmt.Printf("clear notificationType: %d\n", notificationType)
	switch notificationType {
	case model.NotificationTypePoolUsed:
		removeNotification(address, model.NotificationTypePoolUsed)
		break
	case model.NotificationTypeBidInvalid:
		removeNotification(address, model.NotificationTypeBidInvalid)
		break
	case model.NotificationTypeOrderFinish:
		removeNotification(address, model.NotificationTypeOrderFinish)
		break
	default:
		removeNotification(address, model.NotificationTypePoolUsed)
		removeNotification(address, model.NotificationTypeBidInvalid)
		removeNotification(address, model.NotificationTypeOrderFinish)
	}
}

// addNotification
func addNotification(address string, notificationType model.NotificationType) {
	var (
		entity *model.OrderNotificationModel
		err    error
	)
	entity, _ = mongo_service.FindOrderNotificationModelByAddressAndNotificationType(address, notificationType)
	if entity == nil {
		entity = &model.OrderNotificationModel{
			Address:          address,
			NotificationType: notificationType,
		}
	}
	entity.NotificationCount++
	_, err = mongo_service.SetOrderNotificationModel(entity)
	if err != nil {
		major.Println(fmt.Sprintf("addNotification error: %s", err.Error()))
	}
}

// removeNotification
func removeNotification(address string, notificationType model.NotificationType) {
	var (
		entity *model.OrderNotificationModel
		err    error
	)
	entity, _ = mongo_service.FindOrderNotificationModelByAddressAndNotificationType(address, notificationType)
	if entity == nil {
		return
	}
	entity.NotificationCount = 0
	_, err = mongo_service.SetOrderNotificationModel(entity)
	if err != nil {
		major.Println(fmt.Sprintf("removeNotification error: %s", err.Error()))
	}
}
