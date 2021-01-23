package main

import (
	"fmt"

	"cross_share_server/database"
	"cross_share_server/router"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.Infof("main()")
	defer func(){
		database.Release()
	}()

	addr := fmt.Sprintf(":%d", viper.GetInt("server.port"))
	tls := viper.GetBool("tls.enable")
	certFile := viper.GetString("tls.certFile")
	keyFile := viper.GetString("tls.keyFile")
	router.StartServer(addr, tls, certFile, keyFile)
}
