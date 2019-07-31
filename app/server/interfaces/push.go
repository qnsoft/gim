/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         push.go
@ Create Time:  2019-07-31 17:42
@ Software:     GoLand
*/

package interfaces

type PushParams struct {
	AppKey  string `form:"appkey" json:"appkey" binding:"required"`
	Mode    string `form:"mode" json:"mode" binding:"required"`
	Message string `form:"message" json:"message" binding:"required"`
}

//func Push(ctx *gin.Context) {
//	var params PushParams
//	if err := ctx.Bind(&params); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{
//			"message": "Parameter error",
//		})
//	}
//	switch params.Mode {
//	case "im":
//		if onlineMap, err := ChatRoomInstance.GetOnlineMap(params.AppKey); err != nil {
//			ctx.Status(http.StatusBadRequest)
//		} else {
//			// 消息将推送至在线用户私人频道
//			for _, unique := range onlineMap {
//				privateChannel := strings.Join([]string{ChatRoomInstance.ServiceName, unique}, ":")
//				ChatRoomInstance.Publish(privateChannel, params.Message, false)
//			}
//			ctx.Status(http.StatusOK)
//		}
//	case "push":
//		if onlineMap, err := MessagePushInstance.GetOnlineMap(params.AppKey); err != nil {
//			ctx.Status(http.StatusBadRequest)
//		} else {
//			// 消息将推送至在线用户私人频道
//			for _, unique := range onlineMap {
//				privateChannel := strings.Join([]string{MessagePushInstance.ServiceName, unique}, ":")
//				MessagePushInstance.Publish(privateChannel, params.Message, false)
//			}
//			ctx.Status(http.StatusOK)
//		}
//	default:
//		ctx.JSON(http.StatusBadRequest, gin.H{
//			"message": "No matching pattern for mode: " + params.Mode,
//		})
//	}
//}
