package env

import (
	"github.com/spf13/viper"
)

func Bind() {
	viper.AutomaticEnv()
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
