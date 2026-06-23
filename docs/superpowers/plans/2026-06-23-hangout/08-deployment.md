# Hangout Phase 8 Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成香港 ECS 单机部署、Nginx/Caddy 反代、Supabase 生产配置、小范围 IP 内测和正式域名 HTTPS 切换清单。

**Architecture:** 前端构建产物由 Nginx 或 Caddy 托管，`/api/*` 反代到同机 Gin systemd service，`/e/:slug` 也交给 Gin 返回 OG HTML。数据库/Auth 使用阿里云 Supabase，Cloudflare 默认 DNS only。

**Tech Stack:** Linux systemd, Nginx or Caddy, Go binary, Vite static build, Supabase hosted Postgres/Auth, Cloudflare DNS, Let's Encrypt。

## Global Constraints

- 前后端同机部署在香港阿里云 ECS。
- Cloudflare 仅负责 DNS 解析，默认不开代理。
- 正式域名上线前允许使用 ECS 公网 IP 做小范围内测。
- 前端只通过 Supabase 做 Auth/session，业务数据读写一律走 Gin。
- 正式上线后配置 HTTPS，并更新 Supabase `Site URL` / `Redirect URLs`。

---

## File Structure

- Create: `deploy/hangout-api.service`
- Create: `deploy/nginx.conf`
- Create: `deploy/caddy/Caddyfile`
- Create: `deploy/env.production.example`
- Create: `deploy/scripts/build-api.sh`
- Create: `deploy/scripts/build-web.sh`
- Create: `deploy/scripts/smoke-test.sh`
- Create: `docs/deployment/hangout-ecs-runbook.md`
- Create: `.github/workflows/ci.yml` or `scripts/ci.sh` when GitHub Actions is not used.

## Deployment Interfaces

- Gin listens on `127.0.0.1:8080`.
- Reverse proxy serves `/` from `/opt/hangout/web`.
- Reverse proxy forwards `/api/*` and `/e/*` to `http://127.0.0.1:8080`.
- systemd unit reads `/etc/hangout/api.env`.

### Task 1: CI and build scripts

- [ ] **Step 1: 创建后端构建脚本**

Create `deploy/scripts/build-api.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
go test ./apps/api/...
go build -o dist/hangout-api ./apps/api/cmd/api
```

- [ ] **Step 2: 创建前端构建脚本**

Create `deploy/scripts/build-web.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../apps/web"
npm ci
npm run test -- --run
npm run build
mkdir -p ../../dist/web
rm -rf ../../dist/web/*
cp -R dist/* ../../dist/web/
```

- [ ] **Step 3: 创建本地 CI 脚本**

Create `scripts/ci.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
supabase db reset
./deploy/scripts/build-api.sh
./deploy/scripts/build-web.sh
```

- [ ] **Step 4: 验证**

Run:

```bash
chmod +x deploy/scripts/build-api.sh deploy/scripts/build-web.sh scripts/ci.sh
./scripts/ci.sh
```

Expected: PASS.

### Task 2: systemd 与反向代理配置

- [ ] **Step 1: 创建 systemd unit**

Create `deploy/hangout-api.service`:

```ini
[Unit]
Description=Hangout API
After=network.target

[Service]
Type=simple
User=hangout
Group=hangout
WorkingDirectory=/opt/hangout
EnvironmentFile=/etc/hangout/api.env
ExecStart=/opt/hangout/hangout-api
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

- [ ] **Step 2: 创建 Nginx 配置**

Create `deploy/nginx.conf`:

```nginx
server {
    listen 80;
    server_name _;

    root /opt/hangout/web;
    index index.html;

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /e/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        try_files $uri /index.html;
    }
}
```

- [ ] **Step 3: 创建 Caddy 备选配置**

Create `deploy/caddy/Caddyfile`:

```caddyfile
{
	email admin@example.com
}

:80 {
	root * /opt/hangout/web
	encode gzip

	handle /api/* {
		reverse_proxy 127.0.0.1:8080
	}

	handle /e/* {
		reverse_proxy 127.0.0.1:8080
	}

	handle {
		try_files {path} /index.html
		file_server
	}
}
```

- [ ] **Step 4: 创建生产环境变量样例**

Create `deploy/env.production.example`:

```dotenv
API_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://USER:PASSWORD@HOST:PORT/postgres?sslmode=require
SUPABASE_URL=https://PROJECT.supabase.co
SUPABASE_ANON_KEY=replace-with-production-anon-key
SUPABASE_JWT_SECRET=replace-with-production-jwt-secret
SUPABASE_SERVICE_ROLE_KEY=replace-with-production-service-role-key
```

### Task 3: ECS 部署 Runbook

- [ ] **Step 1: 创建部署文档**

Create `docs/deployment/hangout-ecs-runbook.md` with these sections:

- ECS prerequisites: Ubuntu LTS, inbound 80/443, outbound Supabase allowed
- Create `hangout` user
- Install Go runtime only if building on server; otherwise upload built binary
- Install Nginx or Caddy
- Copy `/opt/hangout/hangout-api`
- Copy `/opt/hangout/web/*`
- Copy `/etc/hangout/api.env`
- Install and start systemd unit
- Configure Supabase Auth Site URL and Redirect URLs
- IP smoke test
- Domain DNS only switch
- HTTPS certificate issue
- rollback steps: restore previous binary and web directory, restart service

- [ ] **Step 2: 创建冒烟脚本**

Create `deploy/scripts/smoke-test.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
BASE_URL="${1:?usage: smoke-test.sh https://example.com}"
curl -fsS "$BASE_URL/api/health" | grep -q '"status":"ok"'
curl -fsS "$BASE_URL/" | grep -q '<div id="root"></div>'
curl -fsS "$BASE_URL/e/non-existent-slug" | grep -Eq '404|not_found|Not Found'
echo "smoke test passed"
```

- [ ] **Step 3: 验证脚本语法**

Run:

```bash
bash -n deploy/scripts/*.sh scripts/ci.sh
```

Expected: PASS.

### Task 4: 上线切换清单

- [ ] **Step 1: Supabase hosted 配置**

Complete before public domain switch:

- run migrations against production Supabase
- enable email/password Auth
- set Site URL to `https://正式域名`
- set Redirect URLs to `https://正式域名/*`
- record anon key, service role key, JWT secret in `/etc/hangout/api.env`

- [ ] **Step 2: Cloudflare DNS**

Complete:

- add A record from domain to ECS public IP
- set proxy status to DNS only
- wait for DNS propagation
- verify `curl -I http://domain`

- [ ] **Step 3: HTTPS**

Use Caddy automatic HTTPS or Nginx with certbot. Verify:

```bash
curl -fsS https://domain/api/health
curl -I https://domain/e/example-slug
```

Expected: health returns 200; event share page returns HTML with OG meta for a real slug.

### Commit

- [ ] **Step 1: 阶段验证**

Run:

```bash
bash -n deploy/scripts/*.sh scripts/ci.sh
```

Expected: PASS。

- [ ] **Step 2: 提交**

Run:

```bash
git add deploy docs/deployment scripts
git commit -m "chore: add deployment runbook"
```
