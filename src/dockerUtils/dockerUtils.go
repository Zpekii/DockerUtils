package dockerUtils

import (
	"context"
	"io"
	"os/exec"
	"fmt"
	"log"
	"reflect"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"utils/config"
)

// 初始化docker客户端,获取当前环境的docker客户端
func GetClient() *client.Client {
	
	// 获取docker客户端
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal("获取docker客户端失败,请检查是否启动docker,错误详情:" + err.Error())
		return nil
	}
	return cli
}

// 创建postgres容器,不管容器是否存在,都会创建一个新的容器
func CreateNewPostgresContainer() bool {
	 cli := GetClient()
	 if cli == nil {
		log.Fatal("获取docker客户端失败,请检查是否启动docker")
	}

	// 获取postgres配置
	postgresConfig := config.GetPostgresConfig()

	// 停止并删除同名容器
	containerID := FindContainer(*cli, postgresConfig["name"].(string))
	if containerID != "" {
		StopContainer(*cli, containerID)
		log.Println("已停止运行已存在的同名容器(ID:"+containerID+")")

		RemoveContainer(*cli, containerID)
		log.Println("已删除已存在的同名容器(ID:"+containerID+")")
	}

	// 获取数据库配置,解析配置文件中的数据库配置
	env := config.GetPostgresConfig()["env"].(map[string]interface{})


	dbName := env["postgres_db"].(string)
	dbUsr := env["postgres_user"].(string)
	dbPsswd := env["postgres_password"].(string)


	// 创建Postgres容器
	respID := CreatePostgresContainer(*cli, postgresConfig["name"].(string), dbName, dbUsr, dbPsswd)
	if respID == "" {
		return false
	}
	
	return true
}

// 创建Redis容器,不管容器是否存在,都会创建一个新的容器
func CreateNewRedisContainer() bool {
	cli := GetClient()
	if cli == nil {
		log.Fatal("获取docker客户端失败,请检查是否启动docker")
	}

	// 获取Redis配置
	redisConfig := config.GetRedisConfig()

	// 停止并删除同名容器
	containerID := FindContainer(*cli, redisConfig["name"])
	if containerID != "" {
		StopContainer(*cli, containerID)
		log.Println("已停止运行已存在的同名容器(ID:"+containerID+")")

		RemoveContainer(*cli, containerID)
		log.Println("已删除已存在的同名容器(ID:"+containerID+")")
	}

	// 创建Redis容器
	respID := CreateRedisContainer(*cli, redisConfig["name"])
	if respID == "" {
		return false
	}
	
	return true


}


// 列出正在运行的容器
func ListRunningContainers(cli client.Client) {
	// 检查传入的参数是否是docker客户端
	if reflect.TypeOf(cli) != reflect.TypeOf(client.Client{}) {
		log.Fatal("当前传入的参数不是docker客户端")
		return
	}

	// 获取正在运行的容器列表
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		log.Fatal("获取容器列表失败,错误详情:" + err.Error())
		return
	}

	// 打印容器ID
	log.Println("正在运行的容器ID:")
	for _, container := range containers {
		fmt.Println(container.ID)
	}

}

// 查找指定容器是否在运行
func FindRunningContainer(cli client.Client, containerID string) bool {
	
	// 检查参数
	if reflect.TypeOf(cli) != reflect.TypeOf(client.Client{}) {
		log.Fatal("当前传入的参数不是docker客户端")
	}

	if containerID == "" {
		log.Fatal("容器ID不能为空")
	}

	// 获取正在运行的容器列表
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		log.Fatal("获取容器列表失败,错误详情:" + err.Error())
		return false
	}

	// 查找指定容器是否在运行的容器列表中
	for _, container := range containers {
		if container.ID == containerID {
			return true
		}
	}
	return false
}


// 创建PostgreSQL数据库容器
func CreatePostgresContainer(cli client.Client, containerName string, dbName string, dbUsr string,  dbPsswd string) string {
	
	// 检查参数
	if reflect.TypeOf(cli) != reflect.TypeOf(client.Client{}) {
		log.Fatal("当前传入的参数不是docker客户端")
		return ""
	}
	
	if dbName == "" || dbPsswd == ""{
		log.Fatal("数据库名称或密码不能为空")
		return ""
	}

	
	// 获取postgres配置
	postgresConfig := config.GetPostgresConfig()

	// 检查当前环境是否存在PostgreSQL镜像,不存在则拉取
	_, _, err := cli.ImageInspectWithRaw(context.Background(), postgresConfig["image"].(string))
	if err != nil {
		log.Println("当前环境不存在"+postgresConfig["image"].(string)+"镜像,正在拉取...")

		// 拉取PostgreSQL镜像
		cmd := exec.Command("docker", "pull", postgresConfig["image"].(string))
		if err := cmd.Run(); err != nil {
			log.Fatal("拉取PostgreSQL镜像失败,错误详情:" + err.Error())
			return ""
		}

		log.Println("拉取"+postgresConfig["image"].(string)+"镜像成功")
	}

	// 设置容器配置
	containerConfig := &container.Config{
		Image: postgresConfig["image"].(string), // 使用最新的PostgreSQL官方镜像
		Env: []string{
			"POSTGRES_USER="+dbUsr,          // 设置PostgreSQL数据库的用户名
			"POSTGRES_PASSWORD="+dbPsswd, // 设置PostgreSQL数据库的密码
			"POSTGRES_DB="+dbName,       // 设置PostgreSQL数据库的名称
		},
	}



	// 设置主机配置，映射PostgreSQL端口
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			// 映射容器的5432端口到主机的5433端口
			nat.Port(postgresConfig["containerport"].(string) + "/tcp"): []nat.PortBinding{
				{
					HostIP:   postgresConfig["host"].(string), // 映射到主机的本地地址
					HostPort: postgresConfig["port"].(string),     // 映射到主机的5433端口
				},
			},
		},
	}

	ctx := context.Background()

	// 创建容器
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		log.Fatal("创建容器失败,错误详情:" + err.Error())
	}

	// 启动容器
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Fatal("启动容器"+postgresConfig["name"].(string)+"失败,错误详情:" + err.Error())
	}


	log.Println(postgresConfig["name"].(string)+"容器创建并运行成功,容器ID:" + resp.ID)
	return resp.ID
}

// 创建Radis容器
func CreateRedisContainer(cli client.Client, containerName string) string {
	// 检查参数
	if reflect.TypeOf(cli) != reflect.TypeOf(client.Client{}) {
		log.Fatal("当前传入的参数不是docker客户端")
		return ""
	}

	// 获取Redis配置
	redisConfig := config.GetRedisConfig()

	// 检查当前环境是否存在Redis镜像,不存在则拉取
	_, _, err := cli.ImageInspectWithRaw(context.Background(), redisConfig["image"])
	if err != nil {
		log.Println("当前环境不存在"+redisConfig["image"]+"镜像,正在拉取...")
		
		// 拉取Redis镜像
		cmd := exec.Command("docker", "pull", redisConfig["image"])
		if err := cmd.Run(); err != nil {
			log.Fatal("拉取"+redisConfig["image"]+"镜像失败,错误详情:" + err.Error())
			return ""
		}

		log.Println("拉取"+redisConfig["image"]+"镜像成功")
		
	}


	// 设置容器配置
	containerConfig := &container.Config{
		Image:  redisConfig["image"], // 使用最新的Redis官方镜像
	}

	// 设置主机配置，映射Redis端口
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			// 映射容器的6379端口到主机的6379端口
			nat.Port(redisConfig["containerport"] + "/tcp"): []nat.PortBinding{
				{
					HostIP:    redisConfig["host"], // 映射到主机的本地地址
					HostPort:  redisConfig["port"],     // 映射到主机的6379端口
				},
			},
		},

	}

	// 创建容器
	ctx := context.Background()

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		log.Fatal("创建"+redisConfig["name"]+"容器失败,错误详情:" + err.Error())
		return ""
	}

	// 启动容器
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Fatal(redisConfig["name"]+"启动容器失败,错误详情:" + err.Error())
		return ""
	}

	log.Println(redisConfig["name"]+"容器创建成功,容器ID:" + resp.ID)

	return resp.ID
}



// 删除容器
func RemoveContainer(cli client.Client, containerID string) {
	// 检查参数
	if reflect.TypeOf(cli) != reflect.TypeOf(client.Client{}) {
		log.Fatal("当前传入的参数不是docker客户端")
		return
	}

	if containerID == "" {
		log.Fatal("容器ID不能为空")
		return
	}

	// 执行删除容器
	ctx := context.Background()
	if err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{}); err != nil {
		log.Fatal("删除容器失败,错误详情:" + err.Error())
		return
	}
	log.Println("删除容器(ID:"+containerID+")成功")
}

// 启动容器
func StartContainer(cli client.Client, containerID string) {
	// 检查参数
	if reflect.TypeOf(cli) != reflect.TypeOf(client.Client{}) {
		log.Fatal("当前传入的参数不是docker客户端")
		return
	}

	if containerID == "" {
		log.Fatal("容器ID不能为空")
		return
	}

	// 执行启动容器
	ctx := context.Background()
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		log.Fatal("启动容器失败,错误详情:" + err.Error())
		return
	}
	log.Println("启动容器(ID:"+containerID+")成功")
}


// 停止容器
func StopContainer(cli client.Client, containerID string) {
	// 检查参数
	if reflect.TypeOf(cli) != reflect.TypeOf(client.Client{}) {
		log.Fatal("当前传入的参数不是docker客户端")
		return
	}

	if containerID == "" {
		log.Fatal("容器ID不能为空")
		return
	}

	// 执行停止容器
	ctx := context.Background()
	if err := cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		log.Fatal("停止运行容器(ID:"+containerID+")失败,错误详情:" + err.Error())
		return
	}
	log.Println("停止运行容器(ID:"+containerID+")成功")
}

// 获取指定容器的信息
func GetContainerInfo(cli client.Client, containerID string) types.ContainerJSON {
	// 检查参数
	if reflect.TypeOf(cli) != reflect.TypeOf(client.Client{}) {
		log.Fatal("当前传入的参数不是docker客户端")
		return types.ContainerJSON{}
	}

	if containerID == "" {
		log.Fatal("容器ID不能为空")
		return types.ContainerJSON{}
	}

	// 获取容器信息
	ctx := context.Background()
	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Fatal("获取容器(ID:"+containerID+")信息失败,错误详情:" + err.Error())
		return  types.ContainerJSON{}
	}
	
	log.Println("获取容器(ID:"+containerID+")信息成功")
	
	return  containerInfo
}

// 查找指定容器名是否存在
func FindContainer(cli client.Client, containerName string) string {
	
	// 检查参数
	if reflect.TypeOf(cli) != reflect.TypeOf(client.Client{}) {
		log.Fatal("当前传入的参数不是docker客户端")
		return ""
	}

	if containerName == "" {
		log.Fatal("容器名不能为空")
		return ""
	}

	// 获取容器列表
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		log.Fatal("获取容器列表失败,错误详情:" + err.Error())
		return ""
	}

	// 查找指定容器名是否存在于容器列表中
	for _, container := range containers {
		if container.Names[0] == "/"+containerName {
			log.Println("容器名为:"+containerName+"存在,该容器ID为:"+container.ID)
			return container.ID
		}
	}
	return ""
}

// 获取指定容器的日志
func GetContainerLogs(cli client.Client, containerID string) string {
	
	// 检查参数
	if reflect.TypeOf(cli) != reflect.TypeOf(client.Client{}) {
		log.Fatal("当前传入的参数不是docker客户端")
		return ""
	}

	if containerID == "" {
		log.Fatal("容器ID不能为空")
		return ""
	}

	// 获取容器日志
	ctx := context.Background()
	logs, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true,Tail: "all"})
	if err != nil {
		log.Fatal("获取容器(ID:"+containerID+")日志失败,错误详情:" + err.Error())
		return ""
	}

	content, err := io.ReadAll(logs)
	if err != nil {
		log.Fatal("读取容器(ID:"+containerID+")日志失败,错误详情:" + err.Error())
		return ""
	}


	return string(content)

	
}