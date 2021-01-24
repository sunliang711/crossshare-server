package handlers

import (
	"cross_share_server/database"
	"cross_share_server/types"
	"cross_share_server/utils"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Push(c *gin.Context) {
	filename := c.Request.Header.Get("Filename")
	if filename != "" {
		filename = path.Base(filename)
	}
	logrus.Debugf("filename: %v", filename)
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		msg := fmt.Sprintf("Read request error: %v", err)
		logrus.Error(msg)
		c.JSON(200, types.PushResp{Code: 1, Msg: msg})
		return
	}
	if len(data) > viper.GetInt("business.push_limit") {
		msg := fmt.Sprintf("Request body size: %v exceeds config size: %v", len(data), viper.GetInt("business.push_limit"))
		logrus.Errorf(msg)
		c.JSON(200, types.PushResp{Code: 1, Msg: msg})
		return
	}
	logrus.Debugf("file content len: %v", len(data))
	newKey, hash, err := data2key(data)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		logrus.Errorf(msg)
		c.JSON(200, types.PushResp{Code: 1, Msg: msg})
		return
	}

	if err := database.Rdb.HSet(database.Ctx, newKey, "name", filename, "content", data, "hash", hash).Err(); err != nil {
		msg := "internal redis error"
		c.JSON(200, types.PushResp{Code: 100, Msg: msg})
		logrus.Errorf(msg)
		return
	}

	if err := database.Rdb.Expire(database.Ctx, newKey, time.Second*time.Duration(viper.GetInt64("business.ttl"))).Err(); err != nil {
		msg := "internal redis error"
		c.JSON(200, types.PushResp{Code: 101, Msg: msg})
		logrus.Errorf(msg)
		return
	}

	// c.JSON(200, types.Resp{Code: 0, Msg: "OK", Data: gin.H{"key": md5Str, "ttl": viper.GetInt64("business.ttl")}})
	c.JSON(200, types.PushResp{Code: 0, Msg: "OK", Key: newKey, TTL: viper.GetInt64("business.ttl")})

}

func data2key(data []byte) (string, string, error) {
	hashValue := utils.Md5(data)
	hashStr := hex.EncodeToString(hashValue)
	logrus.Debugf("Request hash: %v", hashStr)

	//hash prefix collision
	n := viper.GetInt("business.hash_min_len")
	newKey := ""
	for n < len(hashStr) {
		fields, err := database.Rdb.HMGet(database.Ctx, hashStr[:n], "type", "hash").Result()
		if err != nil {
			return "", "", fmt.Errorf("Internale redis error")
		}
		// Not exist, not collision OR already exist complete hash
		if fields[0] == nil || fields[1].(string) == hashStr {
			newKey = hashStr[:n]
			break
		}

		logrus.Debugf("hash collision: %v", hashStr[:n])
		n++
	}

	return newKey, hashStr, nil
}
