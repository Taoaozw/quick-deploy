# Quick Deploy

## 系统说明
1. 使用 golang 开发，充分利用 golang 的特性
2. 这是一个命令行工具，用于帮助将程序部署到多个服务器
3. 主要功能：
   - 解析当前文件夹的 `deploy.yaml` 配置文件
   - 支持多服务器配置
   - 支持本地和远程命令的交错执行
   - 提供友好的命令执行输出格式

## 配置文件格式
```yaml
# 服务器配置
servers:
  - name: server1
    host: 192.168.1.100
    port: 22
    username: admin
    password: password123
  - name: server2
    host: 192.168.1.101
    port: 22
    username: admin
    password: password123

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
        - type: remote
          command: "scp ./service1 server:/tmp" # 上传新版本
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
3. 自动忽略特定命令的错误（如 kill 命令）
4. 实时显示命令执行输出
5. 友好的错误处理和提示
6. 分离部署内容定义和实际部署计划

## 使用方法
1. 准备配置文件 `deploy.yaml`
2. 运行命令：
   ```bash
   quick-deploy -config deploy.yaml
   ```

## 注意事项
1. 确保本地环境能够访问目标服务器
2. 确保配置文件中的服务器信息正确
3. 建议使用 SSH 密钥认证（未来支持）
4. 注意保护配置文件中的敏感信息