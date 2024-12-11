# Quick Deploy

## 系统说明
1. 使用 golang 开发，充分利用 golang 的特性
2. 这是一个命令行工具，用于帮助将程序部署到多个服务器
3. 主要功能：
   - 解析当前文件夹的 `deploy.yaml` 配置文件
   - 支持多服务器配置和多部署定义
   - 支持本地和远程命令的交错执行
   - 支持 SCP 文件传输功能
   - 提供友好的命令执行输出格式和进度提示
   - 支持部署计划的定义和执行
   - 支持读取系统 SSH 配置（~/.ssh/config）

## 配置文件格式
```yaml
# 服务器配置
servers:
  - name: server1                # 服务器名称，可以使用 ~/.ssh/config 中定义的 Host
    host: 192.168.1.100         # 主机地址，如果在 SSH 配置中定义了 HostName，这里可以使用 Host 别名
    port: 22                    # SSH 端口
    username: admin             # 用户名
    password: password123       # 密码（可选，如果使用 SSH 密钥认证可以省略）

  - name: server2
    host: my-server            # 可以使用 SSH 配置中定义的别名
    port: 22
    username: root             # 如果 SSH 配置中定义了 User，这里可以省略

# 部署内容定义
deployments:
  - name: service1-deploy
    pipeline:
      commands:
        - type: remote
          command: "systemctl stop service1"    # 先停止远程服务
        - type: local
          command: "go build"                   # 本地编译
          working_dir: "./service1"
        - type: scp                            # 新增 scp 类型
          local_path: "./service1/service1"     # 本地文件路径
          remote_path: "/opt/services/"         # 远程目标路径
        - type: remote
          command: "systemctl start service1"   # 启动服务
  
  - name: service2-deploy
    pipeline:
      commands:
        - type: remote
          command: "kill -9 $(ps aux | grep image | awk '{print $2}' | head -n 1)"
        - type: local
          command: "go build"
          working_dir: "./service2"
        - type: remote
          command: "scp ./service2 server:/tmp"
        - type: remote
          command: "cd /tmp && ./service2"

# 部署计划
deploy_plans:
  - server: server1
    deployment: service1-deploy
  - server: server2
    deployment: service2-deploy
```

## 特性
1. 支持本地和远程命令交错执行
2. 支持指定本地命令的工作目录
3. 自动忽略特定命令的��误（如 kill、rm -f、systemctl stop 命令）
4. 实时显示命令执行输出和执行时间
5. 友好的错误处理和提示
6. 分离部署内容定义和实际部署计划
7. 支持 SCP 文件传输，可自动创建远程目录
8. 提供详细的部署进度和执行状态信息
9. 支持系统 SSH 配置集成：
   - 自动读取 ~/.ssh/config 配置
   - 支持 SSH 配置中的 Host、HostName、User、IdentityFile 等配置项
   - 支持使用 SSH 配置中定义的别名和密钥文件

## 使用方法
1. 准备配置文件 `deploy.yaml`
2. 运行命令：
   ```bash
   quick-deploy -config deploy.yaml
   ```

## 注意事项
1. 确保本地环境能够访问目标服务器
2. 确保配置文件中的服务器信息正确
3. SSH 认证优先级：
   - 首先使用 ~/.ssh/config 中指定的配置和密钥
   - 如果未找到配置，尝试使用默认密钥（~/.ssh/id_rsa 或 ~/.ssh/id_ed25519）
   - 如果配置了密码，将作为最后的认证方式
4. 注意保护配置文件中的敏感信息
5. 部署前请确保本地构建的文件路径正确
6. 远程路径会自动创建，无需手动创建目录
7. 可以在 ~/.ssh/config 中配置服务器别名和认证信息，简化 deploy.yaml 的配置

## SSH 配置示例
```
# ~/.ssh/config 示例
Host my-server
    HostName 192.168.1.100
    User admin
    IdentityFile ~/.ssh/my-server.pem

Host prod-*
    User root
    IdentityFile ~/.ssh/prod-key.pem
```

对应的 deploy.yaml 配置：
```yaml
servers:
  - name: server1
    host: my-server    # 使用 SSH 配置中的别名
    port: 22
  
  - name: server2
    host: prod-app1    # 匹配 prod-* 配置
    port: 22
```