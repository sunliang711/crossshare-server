package handlers

import (
	"cross_share_server/database"
	"cross_share_server/types"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Pull(c *gin.Context) {
	key := c.Param("key")
	delete := c.Request.Header.Get("Delete-After-Pull")
	logrus.Infof("Header DeleteAfterPull: %v", delete)
	keys := []string{"name", "content", "hash"}
	fields, err := database.Rdb.HMGet(database.Ctx, key, keys...).Result()
	if err != nil {
		msg := "internal redis error"
		logrus.Errorf(msg)
		c.Header("Crossshare-Type", "error")
		c.JSON(200, types.Share{
			Code: 1,
			Msg:  msg,
			Type: types.InvalidType,
		})
		return
	}

	if fields[1] == nil {
		msg := "Not found"
		logrus.Infof(msg)
		c.Header("Crossshare-Type", "NotFound")
		c.Data(200, "application/binary", nil)
		return
	}

	keyDeleted := "false"
	if delete == "true" {
		_, err = database.Rdb.Del(database.Ctx, key).Result()
		if err != nil {
			msg := fmt.Sprintf("Delete key error: %v", err)
			logrus.Error(msg)
		} else {
			logrus.Infof("Key deleted: %v", key)
			keyDeleted = "true"
		}
	}

	c.Header("Key-Deleted", keyDeleted)
	filename := fields[0].(string)
	content := fields[1].(string)
	// hash :=fields[1].(string)
	if filename == "" {
		logrus.Infof("text type, text: %v", content)
		c.Header("Crossshare-Type", "Text")
		c.Data(200, "application/binary", []byte(content))
	} else {
		logrus.Infof("file type")
		logrus.Debugf("file name: %v content len: %v", filename, len(content))
		c.Header("Crossshare-Type", "File")
		c.Header("Crossshare-Filename", filename)
		c.Data(200, "application/binary", []byte(content))
	}
}
