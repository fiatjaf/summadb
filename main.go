package main

import (
	"github.com/inconshreveable/log15"
	"github.com/spf13/viper"
	"github.com/summadb/summadb/database"
	"github.com/summadb/summadb/server"
)

var log = log15.New()

func main() {
	viper.SetDefault("path", "/tmp/summadb-server")
	viper.SetDefault("addr", "https://0.0.0.0:6423")
	viper.SetDefault("crt", "default.crt")
	viper.SetDefault("key", "default.key")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AddConfigPath("/etc/summadb/")
	viper.AddConfigPath("$HOME/.summadb/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Error("reading config file", "err", err)
		return
	}

	db := database.Open(viper.GetString("path"))
	defer db.Close()

	server.Start(db, viper.GetString("addr"))
}
