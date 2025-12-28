# runapi - Go代码注释API文档生成器

一个基于 Go 代码注释自动生成 API 文档的工具，支持解析结构体、嵌套字段、包名引用和 `omitempty` 标签。

## 特性

- 🚀 **零配置启动** - 默认配置即可使用
- 📝 **注释驱动** - 通过代码注释定义API文档
- 🏗️ **结构体解析** - 自动解析Go结构体，支持嵌套和嵌入
- 📦 **包名引用** - 支持跨包的结构体引用
- 🏷️ **智能标签** - 自动识别 `omitempty` 标签，标记必传/非必传字段
- 🔧 **灵活配置** - 支持自定义扫描路径和输出配置
- 📚 **多格式输出** - 生成JSON格式文档，支持ShowDoc推送

## 安装

### 通过 go install 安装

```bash
go install github.com/cheivin/go-runapi@latest
```

### 从源码安装

```bash
git clone https://github.com/cheivin/go-runapi.git
cd go-runapi
go build -o runapi ./cmd/runapi
```

## 快速开始

### 1. 初始化配置文件

```bash
runapi -init
```

这会在当前目录创建一个默认的 `runapi.json` 配置文件：

```json
{
  "scan": {
    "dir": "./example",
    "scan": "",
    "extra_dirs": [],
    "include_vendor": false
  },
  "output": {
    "file": "api-docs.json"
  },
  "showdoc": {
    "url": "https://www.showdoc.cc/server/api/open",
    "api_key": "",
    "api_token": "",
    "enabled": false
  }
}
```

### 2. 编写API注释

在你的 Go 代码中添加如下格式的注释：

```go
// Login
// runapi
// @catalog 用户相关/登录
// @title 用户登录
// @description 用户登录的接口
// @method post
// @url /api/login
// @body user.LoginRequest
// @response_body user.LoginResponse
func Login(w http.ResponseWriter, r *http.Request) {
    // 业务逻辑
}
```

### 3. 生成文档

```bash
runapi
```

这会扫描代码并生成 `api-docs.json` 文档文件。

## 注释定义规范

### 基本格式

API 文档注释以 `// runapi` 开始，包含以下标签：

| 标签 | 说明 | 示例 |
|------|------|------|
| `@catalog` | 文档分类 | `@catalog 用户相关/登录` |
| `@title` | 接口标题 | `@title 用户登录` |
| `@description` | 接口描述 | `@description 用户登录的接口` |
| `@method` | HTTP方法 | `@method post` |
| `@router` | 路由路径 | `@router /api/login` |
| `@url` | URL路径（与router二选一） | `@url /api/login` |
| `@remark` | 备注信息 | `@remark 登录接口` |

### 请求参数

#### Header 参数

```go
// @param token header string true 授权token
```

#### Query 参数

```go
// @param page query int false 页码
// @param size query int false 每页数量
```

#### Form Data 参数

```go
// @param avatar formData file true 头像文件
```

#### 请求体（JSON）

```go
// @body user.LoginRequest
```

### 响应参数

#### 响应头

```go
// @response token header string string 认证token
```

#### 响应体

```go
// @response_body user.LoginResponse
```

#### 嵌套响应格式

```go
// @response_body response.Response{data=user.UserInfo}
```

## 结构体定义

### 基本结构体

```go
type LoginRequest struct {
    Username string `json:"username"`           // 必传字段
    Password string `json:"password"`           // 必传字段
    TOTP     string `json:"totp,omitempty"`     // 非必传字段
    Remember bool   `json:"remember,omitempty"` // 非必传字段
}

type LoginResponse struct {
    User  User   `json:"user"`  // 用户信息
    Token string `json:"token"` // 登录凭证
}
```

### 嵌套结构体

```go
type User struct {
    ID       int    `json:"id"`       // 用户ID
    Username string `json:"username"` // 用户名
    Email    string `json:"email,omitempty"` // 邮箱（可选）
}

type UserInfo struct {
    User      User     `json:"user"`       // 嵌入用户信息
    Avatar    string   `json:"avatar"`     // 头像
    CreatedAt string   `json:"created_at"` // 创建时间
}
```

### 跨包引用

```go
import "example/internal/pkg/response"

// @response_body response.Response{data=UserInfo}
```

## omitempty 标签支持

工具会自动识别 `omitempty` 标签：

- **有 `omitempty` 标签**：字段标记为非必传（`"require": "false"`）
- **没有 `omitempty` 标签**：字段标记为必传（`"require": "true"`）

```go
type CreateUserRequest struct {
    Username string `json:"username"`           // 必传
    Email    string `json:"email,omitempty"`    // 非必传
    Phone    string `json:"phone,omitempty"`    // 非必传
    Age      int    `json:"age"`               // 必传
}
```

生成的文档：
```json
{
  "name": "username", "type": "string", "require": "true",  "remark": ""
},
{
  "name": "email",    "type": "string", "require": "false", "remark": ""
},
{
  "name": "phone",    "type": "string", "require": "false", "remark": ""
},
{
  "name": "age",      "type": "int",    "require": "true",  "remark": ""
}
```

## 配置说明

### 扫描配置

```json
{
  "scan": {
    "dir": "./example",                    // 根扫描路径（用于结构体解析）
    "scan": "./example/controller",        // 文档注释扫描路径（可选，默认同dir）
    "extra_dirs": [],                      // 额外的扫描目录
    "include_vendor": false                // 是否包含vendor目录
  }
}
```

**配置规则：**
- 如果 `scan` 没指定，则默认同 `dir` 路径
- 如果 `dir` 路径没指定，则默认同当前运行路径

### 输出配置

```json
{
  "output": {
    "file": "api-docs.json"  // 输出文件路径
  }
}
```

### ShowDoc 配置

```json
{
  "showdoc": {
    "url": "https://www.showdoc.cc/server/api/open",
    "api_key": "your_api_key",
    "api_token": "your_api_token",
    "enabled": true
  }
}
```

## 使用示例

### 完整的API注释示例

```go
// GetUserInfo
// runapi
// @catalog 用户相关/用户信息
// @title 获取用户信息
// @description 根据用户ID获取用户详细信息
// @method get
// @router /api/user/{id}
// @param id path int true 用户ID
// @param token header string true 授权token
// @param fields query string false 返回字段，逗号分隔
// @response_body response.Response{data=user.UserInfo}
// @remark 需要登录才能访问
func GetUserInfo(w http.ResponseWriter, r *http.Request) {
    // 业务逻辑
}
```

### 生成的文档示例

```json
{
  "title": "获取用户信息",
  "catalog": "用户相关/用户信息",
  "description": "根据用户ID获取用户详细信息",
  "method": "get",
  "router": "/api/user/{id}",
  "header": [
    {
      "name": "token",
      "type": "string",
      "require": "true",
      "remark": "授权token"
    }
  ],
  "query": [
    {
      "name": "fields",
      "type": "string",
      "require": "false",
      "remark": "返回字段，逗号分隔"
    }
  ],
  "response_body": [
    {
      "name": "code",
      "type": "int",
      "required": true,
      "remark": "状态码"
    },
    {
      "name": "msg",
      "type": "string",
      "required": true,
      "remark": "提示信息"
    },
    {
      "name": "data",
      "type": "object",
      "required": false,
      "remark": "数据"
    },
    {
      "name": "data.user.id",
      "type": "int",
      "required": true,
      "remark": "用户ID"
    },
    {
      "name": "data.user.username",
      "type": "string",
      "required": true,
      "remark": "用户名"
    },
    {
      "name": "data.avatar",
      "type": "string",
      "required": true,
      "remark": "头像"
    }
  ],
  "remark": "需要登录才能访问"
}
```

## 命令行选项

```bash
# 显示帮助信息
runapi -help

# 初始化配置文件
runapi -init

# 使用指定配置文件
runapi -config ./custom.json

# 仅生成文档
runapi -mode generate

# 仅推送到ShowDoc
runapi -mode push

# 生成并推送变更文档
runapi -mode genpush
```

## 运行模式

| 模式 | 说明 |
|------|------|
| `generate` | 仅生成文档文件（默认模式） |
| `push` | 仅推送现有文档到ShowDoc |
| `genpush` | 生成文档并推送变更到ShowDoc |

## 最佳实践

### 1. 项目结构建议

```
project/
├── api/           # API处理器
├── internal/      # 内部包
│   ├── model/     # 数据模型
│   └── pkg/       # 内部包
├── runapi.json    # 配置文件
└── README.md      # 项目文档
```

### 2. 配置建议

```json
{
  "scan": {
    "dir": "./internal",
    "scan": "./api",
    "extra_dirs": ["./pkg"]
  }
}
```

### 3. 注释规范

- 每个API接口都应该有清晰的 `@title` 和 `@description`
- 使用 `@catalog` 进行合理的分类
- 为复杂的数据结构创建专门的结构体
- 使用 `omitempty` 标签明确标识可选字段
- 添加必要的 `@remark` 说明特殊要求

### 4. 版本控制

建议将 `runapi.json` 加入版本控制，但将生成的文档文件（如 `api-docs.json` 加入 `.gitignore`）。

## 故障排除

### 常见问题

1. **找不到结构体**
   - 检查 `scan.dir` 和 `scan.extra_dirs` 配置
   - 确认结构体在扫描路径内

2. **跨包引用失败**
   - 确认包导入路径正确
   - 检查结构体是否在 `extra_dirs` 中

3. **字段必传性不正确**
   - 检查 JSON 标签中的 `omitempty` 设置
   - 确认结构体字段有正确的 JSON 标签

### 调试技巧

1. 使用详细的配置来查看扫描路径：
   ```json
   {
     "scan": {
       "dir": "./example",
       "scan": "./example",
       "extra_dirs": []
     }
   }
   ```

2. 检查生成的文档文件，确认字段和类型是否正确

3. 对于复杂的嵌套结构，建议逐步测试

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License