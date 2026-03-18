# CLIAPI Release

这是一个**干净可发布**的发行目录，可单独上传到 GitHub 仓库或你的静态下载站。

## 内置功能

- 用户注册 / 登录
- JWT 鉴权
- API Key 管理
- OpenAI 兼容 `/v1/chat/completions`
- Token 额度扣费
- 兑换码充值
- 超级管理员生成兑换码
- Docker Compose 部署
- curl 一键安装脚本

## 目录

```text
.
├── app/
├── sql/
├── installer/
│   └── install.sh
├── scripts/
│   ├── init-super-admin.sh
│   └── release-package.sh
├── .env.example
├── .gitignore
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## 一、发布前要做的事

### 1. 上传这个目录到一个独立仓库
建议仓库名：

- `cliapi`

### 2. 生成发行包

```bash
cd cliapi-release
bash scripts/release-package.sh v0.1.0
```

会生成：

- `release/cliapi-v0.1.0.tar.gz`
- `release/cliapi-latest.tar.gz`

### 3. 把这些文件发布到你的下载地址
至少要能访问：

- `https://raw.githubusercontent.com/xzy593232500/cliapi/main/install.sh`
- `https://raw.githubusercontent.com/xzy593232500/cliapi/main/.env.example`
- `https://github.com/xzy593232500/cliapi/releases/download/v0.1.0/cliapi-latest.tar.gz`

你可以放在：

- GitHub Releases
- Nginx 静态目录
- 对象存储 + CDN

---

## 二、用户安装方式

```bash
bash <(curl -sL https://raw.githubusercontent.com/xzy593232500/cliapi/main/install.sh) app CLIAPI
```

或者：

```bash
bash <(curl -sL https://raw.githubusercontent.com/xzy593232500/cliapi/main/install.sh) install CLIAPI
```

如果你临时不用域名，也可以这样指定：

```bash
CLIPROXYAPI_PUBLIC_BASE_URL="https://your-download-base-url" \
  bash <(curl -sL https://your-download-base-url/install.sh) install CLIAPI
```

---

## 三、部署后初始化管理员

```bash
cd /opt/cliapi
bash scripts/init-super-admin.sh admin@example.com 'StrongPass123!' 100000
```

这会创建或更新一个 `super_admin` 账户，并设置初始余额。

---

## 四、关键配置

编辑 `/opt/cliapi/.env`：

```env
POSTGRES_DB=cliapi
POSTGRES_USER=cliapi
POSTGRES_PASSWORD=your-postgres-password
JWT_SECRET=your-jwt-secret
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_API_KEY=sk-xxxx
API_KEY_PREFIX=ocp
TOKEN_RATE_TOKENS=500000
TOKEN_RATE_CENTS=100
ALLOW_REGISTRATION=true
PORT=8787
```

默认价格：

- `500000 tokens = $1`
- 即 `1M tokens = $2`

---

## 五、推荐发布方式

### 方案 A：GitHub Releases
最简单。

发布：
- `install.sh`
- `.env.example`
- `cliapi-latest.tar.gz`

然后把 `installer/install.sh` 里的 `PUBLIC_BASE_URL` 改成你的真实 URL。

### 方案 B：Nginx 静态目录
比如：

```text
/var/www/html/cliapi/
  install.sh
  .env.example
  release/cliapi-latest.tar.gz
```

这样用户就能直接 curl 安装。

---

## 六、现阶段说明

这是一个最小可用发行版，已经适合：

- 云服务器部署
- 自己试用
- 小规模分发

下一版建议加：

- `stream=true` 流式转发
- 后台管理页面
- Redis 限流
- 多模型单独计价
- 支付入口
- 更安全的兑换码哈希存储
