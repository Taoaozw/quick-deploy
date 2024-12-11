# Quick Deploy - 设计文档

## 概述
Quick Deploy 是一个用 Go 语言编写的命令行工具，用于将应用程序部署到多个服务器。它通过读取 `deploy.yaml` 配置文件来获取服务器详情和部署命令，然后按顺序执行这些命令，同时提供友好的输出信息。

## 核心功能
1. YAML 配置解析
   - 解析服务器配置（主机、端口、凭证）
   - 解析部署内容定义（命令序列）
   - 解析部署计划（哪些服务器执行哪些部署）
   - 支持多服务器和多部署配置

2. 命令执行
   - 支持指定工作目录的本地命令执行
   - 通过 SSH 执行远程命令
   - 实时显示命令输出
   - 命令执行状态检查

3. 用户界面
   - 清晰、格式化的控制台输出
   - 每个步骤的进度提示
   - 错误处理和报告

## 技术架构

### 组件

1. **Config 包**
   - `Config`: 主配置结构
   - `Server`: 服务器配置结构
   - `Command`: 命令配置结构
   - YAML 解析功能

2. **SSH 包**
   - SSH 客户端实现
   - 远程命令执行
   - 安全凭证处理

3. **Executor 包**
   - 本地命令执行
   - 通过 SSH 执行远程命令
   - 命令输出处理
   - 执行状态检查

### 数据结构

```go
// 命令类型
type CommandType string

const (
    CommandTypeLocal  CommandType = "local"
    CommandTypeRemote CommandType = "remote"
)

// 命令配置
type Command struct {
    Type       CommandType `yaml:"type"`
    Command    string      `yaml:"command"`
    WorkingDir string      `yaml:"working_dir,omitempty"`
}

// 部署流程
type Pipeline struct {
    Commands []Command `yaml:"commands"`
}

// 服务器配置
type Server struct {
    Name     string `yaml:"name"`
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

// 部署内容定义
type DeploymentDefinition struct {
    Name     string   `yaml:"name"`
    Pipeline Pipeline `yaml:"pipeline"`
}

// 部署计划
type DeploymentPlan struct {
    ServerName      string `yaml:"server"`
    DeploymentName  string `yaml:"deployment"`
}

// 完整配置
type Config struct {
    Servers     []Server         `yaml:"servers"`
    Deployments DeploymentConfig `yaml:"deployments"`
}
```

## 实现计划

1. **阶段1: 配置管理**
   - 实现 YAML 配置解析
   - 添加配置验证
   - 创建配置加载功能

2. **阶段2: 命令执行**
   - 实现本地命令执行器
   - 实现 SSH 客户端和远程执行
   - 添加工作目录支持
   - 实现命令执行状态检查

3. **阶段3: 用户界面**
   - 实现格式化输出
   - 添加进度提示
   - 实现错误处理和报告

4. **阶段4: 测试和文档**
   - 所有组件的单元测试
   - 集成测试
   - 文档和使用示例

## 错误处理
- 配置验证错误
- SSH 连接错误
- 命令执行错误
- 状态检查失败
- 特定命令错误忽略（如 kill 命令）

## 安全考虑
- 安全的密码处理
- SSH 密钥支持（未来增强）
- 命令注入防护
- 敏感信息日志处理

## 未来增强
1. SSH 密钥认证支持
2. 并行部署执行
3. 回滚功能
4. 部署历史和日志
5. 部署确认的交互模式
6. 环境变量支持
7. 命令超时配置
