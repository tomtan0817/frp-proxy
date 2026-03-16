# frp-proxy Web 管理面板设计

## 概述

构建一个 Web 管理面板，结合 Nginx 反向代理与 frps，让用户可以通过面板自助注册二级域名并管理 frpc 连接。管理员可以管理用户、分配域名配额、管理邀请码。

## 技术栈

| 项目 | 选择 |
|------|------|
| 部署方式 | Linux 裸机 |
| 架构 | 单体 Go 应用 + frps + Nginx |
| 后端 | Go + Gin + PostgreSQL |
| 前端 | React + Ant Design，嵌入 Go 二进制 |
| 面板端口 | 5040（内部） |
| 用户认证 | 用户名密码 + JWT |
| frpc 认证 | 每域名独立 token + frps server plugin 验证 |

## 整体架构

```
                    ┌─────────────────────────────────┐
                    │           Nginx                  │
                    │  *.example.com:80/443            │
                    ├─────────────────────────────────┤
                    │  panel.example.com → Go面板:5040 │
                    │  *.example.com    → frps:8081    │
                    └──────┬──────────────┬───────────┘
                           │              │
                    ┌──────▼──────┐ ┌─────▼─────┐
                    │  Go 面板     │ │   frps    │
                    │  :5040      │ │   :7000   │ (frpc连接端口)
                    │  Web UI     │ │   :8081   │ (HTTP代理端口)
                    │  REST API   │ │           │
                    │  Plugin API │◄┤ plugin回调 │
                    │             │ │           │
                    └──────┬──────┘ └───────────┘
                           │
                    ┌──────▼──────┐
                    │ PostgreSQL  │
                    └─────────────┘
```

**端口分配：**
- Nginx: 80/443（对外）
- Go 面板: 5040（内部）
- frps: 7000（frpc 连接，直接对外）、8081（HTTP 代理，Nginx 转发）

## 数据模型

```sql
-- 用户表
users (
  id            SERIAL PRIMARY KEY,
  username      VARCHAR UNIQUE NOT NULL,
  password_hash VARCHAR NOT NULL,
  role          VARCHAR NOT NULL DEFAULT 'user',    -- 'admin' | 'user'
  status        VARCHAR NOT NULL DEFAULT 'pending', -- 'pending' | 'active' | 'disabled'
  max_domains   INT NOT NULL DEFAULT 1,
  created_at    TIMESTAMP,
  updated_at    TIMESTAMP
)

-- 域名表
domains (
  id            SERIAL PRIMARY KEY,
  user_id       INT REFERENCES users(id),
  subdomain     VARCHAR UNIQUE NOT NULL,
  token         VARCHAR UNIQUE NOT NULL,
  status        VARCHAR NOT NULL DEFAULT 'active',  -- 'active' | 'disabled'
  created_at    TIMESTAMP,
  updated_at    TIMESTAMP
)

-- 邀请码表
invite_codes (
  id          SERIAL PRIMARY KEY,
  code        VARCHAR UNIQUE NOT NULL,
  max_uses    INT NOT NULL DEFAULT 1,
  used_count  INT NOT NULL DEFAULT 0,
  created_by  INT REFERENCES users(id),
  expires_at  TIMESTAMP,                            -- NULL 为永不过期
  created_at  TIMESTAMP
)
```

## API 设计

### 认证
- `POST /api/auth/register` — 注册（可选 invite_code 字段，有邀请码直接激活）
- `POST /api/auth/login` — 登录，返回 JWT
- `POST /api/auth/logout` — 登出

### 用户端
- `GET /api/domains` — 查看自己的域名列表
- `POST /api/domains` — 注册新域名（检查配额 + 唯一性）
- `DELETE /api/domains/:id` — 删除自己的域名

### 管理员端
- `GET /api/admin/users` — 用户列表（支持 ?status=pending 筛选）
- `POST /api/admin/users` — 创建用户
- `PUT /api/admin/users/:id` — 编辑用户（含修改 max_domains）
- `PUT /api/admin/users/:id/activate` — 激活用户
- `DELETE /api/admin/users/:id` — 删除用户（级联删除域名）
- `GET /api/admin/domains` — 所有域名列表
- `POST /api/admin/domains` — 为任意用户创建域名
- `PUT /api/admin/domains/:id` — 编辑域名（启用/禁用）
- `DELETE /api/admin/domains/:id` — 删除域名
- `GET /api/admin/invite-codes` — 邀请码列表
- `POST /api/admin/invite-codes` — 生成邀请码（设定次数、过期时间）
- `DELETE /api/admin/invite-codes/:id` — 删除邀请码

### frps Plugin 端点
- `POST /api/plugin/login` — frpc 连接时验证 token
- `POST /api/plugin/new-proxy` — 创建代理时验证 token + subdomain 匹配

## frps Plugin 验证流程

frps 配置启用 server plugin：

```toml
# frps.toml
bindPort = 7000
vhostHTTPPort = 8081

[httpPlugins]
  [httpPlugins.login]
    addr = "http://127.0.0.1:5040/api/plugin/login"
    path = "/api/plugin/login"
    ops = ["Login"]

  [httpPlugins.new-proxy]
    addr = "http://127.0.0.1:5040/api/plugin/new-proxy"
    path = "/api/plugin/new-proxy"
    ops = ["NewProxy"]
```

**Login 阶段：** frpc 连接时，frps 将 `metadatas.token` 发给面板验证，token 存在且 status=active 则放行。

**NewProxy 阶段：** frpc 创建代理时，面板验证 token 对应的域名是否与请求的 subdomain 匹配。

**重复 token 处理：** 允许多个 frpc 用同一 token 连接，frps 自身会拒绝重复 subdomain 注册。

## frpc 配置示例

```toml
serverAddr = "example.com"
serverPort = 7000
metadatas.token = "用户从面板获取的token"

[[proxies]]
name = "web"
type = "http"
localPort = 3000
subdomain = "myapp"
```

## Nginx 配置

```nginx
# panel.example.com → Go 面板
server {
    listen 80;
    server_name panel.example.com;

    location / {
        proxy_pass http://127.0.0.1:5040;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# *.example.com → frps HTTP 代理
server {
    listen 80;
    server_name *.example.com;

    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 注册机制

1. 用户注册 → 账号状态为 `pending`
2. 激活方式二选一：
   - **管理员审核** — 管理员在后台手动激活
   - **邀请码** — 注册时填入有效邀请码，直接激活
3. 未激活用户（pending）不能注册域名

## 前端页面

### 用户端
- **登录页 / 注册页**
- **我的域名** — 域名列表、注册新域名、删除域名、复制 token、frpc 配置示例

### 管理员端
- **用户管理** — 用户列表、创建/编辑/删除/激活用户、设置 max_domains
- **域名管理** — 所有域名列表、CRUD、按用户筛选
- **邀请码管理** — 生成/删除邀请码、设定次数和过期时间

UI 使用 React + Ant Design。

## 项目目录结构

```
frp-proxy/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── admin.go
│   ├── handler/
│   │   ├── auth.go
│   │   ├── domain.go
│   │   ├── admin_user.go
│   │   ├── admin_domain.go
│   │   ├── admin_invite.go
│   │   └── plugin.go
│   ├── model/
│   │   ├── user.go
│   │   ├── domain.go
│   │   └── invite_code.go
│   ├── service/
│   │   ├── auth.go
│   │   ├── domain.go
│   │   ├── user.go
│   │   └── invite.go
│   └── database/
│       └── db.go
├── web/
│   ├── src/
│   ├── package.json
│   └── ...
├── configs/
│   ├── app.toml
│   ├── frps.toml
│   └── frpc.toml
├── go.mod
└── go.sum
```
