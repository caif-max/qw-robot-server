package handler

import (
	"encoding/json"
	"qw-robot-server/common/websocket"

	"github.com/gin-gonic/gin"
)

type SendMessageToUserRequest struct {
	Action   string `json:"action"`
	Account  string `json:"account"`
	UserId   string `json:"userId"`
	FriendId string `json:"friendId"`
	Message  string `json:"message"`
}

type SendMessageToAllRequest struct {
	Account string   `json:"account"`
	UserIds []string `json:"userIds"`
	Message string   `json:"message"`
}

func SendMessageToUser(c *gin.Context) {
	request := &SendMessageToUserRequest{}
	c.ShouldBindJSON(request)

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		c.JSON(200, gin.H{
			"result":  "Failed",
			"message": "JSON marshal failed: " + err.Error(),
		})
		return
	}

	err = websocket.SendMessageToUser(request.Account, request.UserId, string(jsonBytes))
	if err != nil {
		c.JSON(200, gin.H{
			"result":  "Failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"result": "OK",
	})
}

func SendMessageToAll(c *gin.Context) {
	request := &SendMessageToAllRequest{}
	c.ShouldBindJSON(request)

	err := websocket.SendMessageToAll(request.Account, request.UserIds, request.Message)
	if err != nil {
		c.JSON(200, gin.H{
			"result":  "Failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"result": "OK",
	})
}
