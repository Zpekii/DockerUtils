### 简介:

- 本程序使用了Go实现通过docker客户端进行创建postgres(使用镜像:postgres:alpine)容器和redis(使用镜像:redis:latest)容器
- 程序执行过程:   
  - 停止与配置文件中同名容器(如果存在) --> 删除与配置文件中同名容器(如果存在) --> 创建并运行容器
  - 如果当前docker客户端不存在指定镜像，则会自动拉取
- 程序仍有诸多不足，望谅解，同时非常欢迎提出改进建议

### 要求:

- 当前Docker客户端处于运行状态

- 与当前"utils.exe"一个目录下有"config.json"配置文件且配置正确

  - 容器的配置可根据需要自定义，以下是默认配置
  
  - ```json
    {
        "postgres":{
            "name":"test-postgres",
            "host":"localhost",
            "port":"5433",
            "containerPort":"5432",
            "image":"postgres:alpine",
            "env":{
                "POSTGRES_USER":"test",
                "POSTGRES_PASSWORD":"test",
                "POSTGRES_DB":"test"
            }
        },
        "redis":{
            "name":"test-redis",
            "host":"localhost",
            "port":"6379",
            "containerPort":"6379",
            "image":"redis:latest"
        }
    
    }
    ```
  
    
  

### 使用步骤:

- 运行当前目录下的 utils.exe

  - 注意：当前目录必须有配置文件"config.json"

- 执行成功后根据提示输入回车结束

  - 正常执行界面：

    - 当前Docker客户端没有镜像:

      ![image-20240428184429813](C:\Users\Zpekii\AppData\Roaming\Typora\typora-user-images\image-20240428184429813.png)

      ![image-20240428184435859](C:\Users\Zpekii\AppData\Roaming\Typora\typora-user-images\image-20240428184435859.png)

      ![image-20240428184440840](C:\Users\Zpekii\AppData\Roaming\Typora\typora-user-images\image-20240428184440840.png)

    - 当前客户端已有镜像:

      ![image-20240428184509550](C:\Users\Zpekii\AppData\Roaming\Typora\typora-user-images\image-20240428184509550.png)

​		