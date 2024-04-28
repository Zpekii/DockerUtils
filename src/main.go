package main

import (
	"fmt"
	"log"
	"utils/dockerUtils"
	"utils/config"
)

func main() {
	// 初始化配置
	config.InitConfig()


	logger := log.New(log.Writer(), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errLogger := log.New(log.Writer(), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// 创建postgres容器
	if dockerUtils.CreateNewPostgresContainer() {
		logger.Println("postgres容器创建成功")
	} else {
		errLogger.Println("postgres容器创建失败")
	}

	// 创建redis容器
	if dockerUtils.CreateNewRedisContainer() {
		logger.Println("redis容器创建成功")
	} else {
		errLogger.Println("redis容器创建失败")
	}


	fmt.Println("执行成功,按回车键退出...")
	fmt.Scanln()
}