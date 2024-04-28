package config

import (
	"log"

	"github.com/spf13/viper"
)



func InitConfig() {
	// 设置配置文件名
	viper.SetConfigName("config")
	
	// 设置配置文件类型
	viper.SetConfigType("json")
	
	// 设置配置文件路径
	viper.AddConfigPath(".")
	
	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("读取配置文件失败,错误详情:" + err.Error())
	}

	// 设置默认值
	viper.SetDefault("postgres.name", "postgres")
	viper.SetDefault("postgres.port", "5432")
	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.containerPort", "5432")
	viper.SetDefault("postgres.image", "postgres:alpine")
	viper.SetDefault("postgres.env",map[string]string{
		"POSTGRES_USER":"postgres",
		"POSTGRES_PASSWORD":"postgres",
		"POSTGRES_DB":"postgres",
	})


	viper.SetDefault("redis.name", "redis")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.containerPort", "6379")
	viper.SetDefault("redis.image", "redis:latest")

}

func GetPostgresConfig() map[string]interface{}{
	return viper.GetStringMap("postgres")
}

func GetRedisConfig() map[string]string{
	return viper.GetStringMapString("redis")
}
