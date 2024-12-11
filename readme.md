## 系统
1. 使用golang开发，充分使用golang的特性
2. 是一个命令行工具,帮助部署程序到多个服务器
    - 解析当前文件夹的 deploy.yaml文件，包括多个服务器地址，用户名，密码，需要本地执行的命令和远程执行的命令 还有检测执行结果的命令，比如pgrep -f "image*" 看对应的进程是否存在
    - 执行命令，同时能够看到执行命令的过程和输出，包括本地执行的命令，远程执行的命令，检测执行结果的命令，打印的格式要求阅读友好
3. 先给出一个设计文档，使用md格式，再实现
4. 实现的时候需要完整的注释
5. 以下是配置文件格式
    ```yaml
    servers:
    - name: server1
    host: 192.168.1.100
    port: 22
    username: admin
    password: password123
    servers:
    - name: server2
    host: 192.168.1.100
    port: 22
    username: admin
    password: password123 
    
    deployments:
        servers:
            - name: server1
            pipe:
                commands:
                    - type: remote
                        command: "systemctl stop service1"    # 先停止远程服务
                    - type: local
                        command: "go build"                   # 本地编译
                        working_dir: "./service1"
                    - type: remote
                        command: "scp ./service1 server:/tmp" # 上传新版本
                    - type: remote
                        command: "systemctl start service1"   # 启动服务
            - name: server2
            pipe:
                commands:
                    - type: remote
                        command: "systemctl stop service1"    # 先停止远程服务
                    - type: local
                        command: "go build"                   # 本地编译
                        working_dir: "./service1"
                    - type: remote
                        command: "scp ./service1 server:/tmp" # 上传新版本
                    - type: remote
                        command: "systemctl start service1"   # 启动服务              

    ```