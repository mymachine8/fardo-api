package common

import (
	"github.com/NaySoftware/go-fcm"
	"fmt"
)

const (
	serverKey = "AIzaSyDcXHtO1YB-tCRwFAvaVqJDud8gR00VJs4"
)

func sendGCMNotification() {
	ids := []string{
		"dWVDHS_JCig:APA91bGN_YnMg-nyr8CT2O1t12F3t-3A0zZo2E3jMcPPQY81hkjxUkRbBlAlppizSbVPTbHMpctebcPwC8jeL2pZ9ejzA7Zi-JWftx-4f4qEkeCCLDSSa3XhT3qX7cye1W27_ndJAdM5",
	}

	data := map[string]string{
		"msg": "Hello World1",
		"sum": "Happy Day",
	}

	c := fcm.NewFcmClient(serverKey)
	c.NewFcmRegIdsMsg(ids, data)
	var notification fcm.NotificationPayload;
	notification.Title = "Golang";
	notification.Body = "Hello body";
	c.SetNotificationPayload(&notification);

	status, err := c.Send()

	if err == nil {
		status.PrintResults()
	} else {
		fmt.Println("FCM Error: " +err.Error());
	}
}
