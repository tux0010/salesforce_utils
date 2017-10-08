package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("settings")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Unable to read config file: %s", err)
	}

	port := viper.GetInt("port")

	s, err := NewSalesforce(
		viper.GetString("consumer_key"),
		viper.GetString("consumer_secret"),
		viper.GetString("redirect_url"),
		viper.GetString("login_base_url"),
		viper.GetString("refresh_token"),
	)

	if err != nil {
		log.Fatal(err)
	}

	router := httprouter.New()
	router.GET("/token/new", s.loginHandler)
	router.GET("/token/receive", s.receiveTokenHandler)
	router.POST("/token/parse", s.parseTokenHandler)
	router.GET("/token/refresh", s.refreshTokenHandler)

	log.WithFields(log.Fields{"port": port}).Println("Starting server")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
