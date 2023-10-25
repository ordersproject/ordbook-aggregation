package respond

import "ordbook-aggregation/model"

type NotificationResponse struct {
	Total   int64               `json:"total,omitempty"`
	Results []*NotificationItem `json:"results,omitempty"`
}

type NotificationItem struct {
	NotificationType  model.NotificationType `json:"notificationType,omitempty"`  // 1-pool used, 2-bid invalid, 3-order finish
	NotificationCount int64                  `json:"notificationCount,omitempty"` //available count
	NotificationDesc  string                 `json:"notificationDesc,omitempty"`  //description
	NotificationTitle string                 `json:"notificationTitle,omitempty"` //title
}
