# 附件通 部署脚本说明

> 完全接管已废弃的「公众号附件助手」项目的部署位置  
> **域名 / 路径 / 端口 / PM2 名 / 二进制名 全部沿用**，仅数据库换为新库 `fujiantong`

## 资源对照表

| 资源 | 值 | 说明 |
|---|---|---|
| 域名 | `fujian.5g6g.top` | 沿用，证书已就绪 |
| 服务器路径 | `/www/wwwroot/fujiantong` | 沿用 |
| 后端端口 | `8080` | 沿用 |
| PM2 进程名 | `fujiantong` | 沿用 |
| Go 二进制 | `fujian-tong` | 沿用 |
| MySQL 数据库 | **`fujiantong`**（新库） | 不复用旧 `fujiantong_v2` |

---

## 一次性接管流程（在宝塔终端按顺序执行一次）

```bash
# ────────────────────────────────────────────────
# 第 1 步：停掉并清理旧项目（已废弃，可放心删）
# ────────────────────────────────────────────────
pm2 delete fujiantong 2>/dev/null || true
pm2 save

# 备份旧代码以防万一（如不需要可直接 rm -rf）
mv /www/wwwroot/fujiantong /www/wwwroot/fujiantong.old.$(date +%Y%m%d) 2>/dev/null || true

# ────────────────────────────────────────────────
# 第 2 步：拉取新仓库
# ────────────────────────────────────────────────
cd /www/wwwroot
git clone <你的 GitHub 仓库地址> fujiantong
cd fujiantong

# ────────────────────────────────────────────────
# 第 3 步：建数据库（宝塔可视化建库也行）
# ────────────────────────────────────────────────
# 库名：fujiantong   字符集：utf8mb4

# ────────────────────────────────────────────────
# 第 4 步：配置 .env
# ────────────────────────────────────────────────
cd server
cp .env.production.example .env
vi .env
#   - DB_DSN 改成 fujiantong 库连接串
#   - JWT_SECRET=$(openssl rand -hex 32) 替换
#   - ADMIN_PASSWORD 改强密码
#   - WX_COMPONENT_* 填微信开放平台后台获取的值
cd ..

# ────────────────────────────────────────────────
# 第 5 步：首次部署
# ────────────────────────────────────────────────
chmod +x scripts/deploy.sh
bash scripts/deploy.sh

# ────────────────────────────────────────────────
# 第 6 步：让 PM2 开机自启
# ────────────────────────────────────────────────
pm2 startup
pm2 save

# ────────────────────────────────────────────────
# 第 7 步：宝塔网站配置
# ────────────────────────────────────────────────
# 网站 fujian.5g6g.top 已存在，只需检查 / 调整：
#   1. 网站根目录已经是 /www/wwwroot/fujiantong/server/public（前端打包路径）
#      若不是，进入「站点 → 网站目录」改为该路径
#   2. SSL 证书已申请且强制 HTTPS 已勾选
#   3. 反向代理 / 配置文件中：API 反代到 127.0.0.1:8080
#      （旧项目已配过，直接复用即可，nginx-fujiantong.conf 在本项目无需新建）

# 验证
curl https://fujian.5g6g.top/ping
# 期望：{"code":200,"msg":"pong - 附件通 is alive!"}
```

---

## 日常更新（本地推 → 服务器拉）

### 本地 Windows

```powershell
cd E:\项目\小项目\附件通
git add -A
git commit -m "feat: xxx"
git push origin main
```

### 服务器（宝塔终端）

```bash
# 完整更新（推荐）
cd /www/wwwroot/fujiantong && bash scripts/deploy.sh

# 只更新后端
cd /www/wwwroot/fujiantong && bash scripts/deploy.sh backend

# 只更新前端
cd /www/wwwroot/fujiantong && bash scripts/deploy.sh frontend
```

脚本会做：
1. `git fetch + reset --hard origin/main` 强制对齐主干
2. 编译 Go 二进制（`.new` 原子替换，零停机）
3. `npm run build --prefix admin` 打包到 `server/public/`
4. `pm2 reload fujiantong` 平滑重启

---

## 常用 PM2 命令

```bash
pm2 list                       # 查看进程
pm2 logs fujiantong            # 实时日志
pm2 logs fujiantong --lines 200
pm2 restart fujiantong         # 直接重启
pm2 stop fujiantong            # 停服
pm2 monit                      # 资源监控面板
tail -f /www/wwwroot/fujiantong/server/logs/pm2-out.log
```

---

## 微信开放平台联调

| 配置项 | URL |
|---|---|
| 服务器配置 URL（ticket 推送） | `https://fujian.5g6g.top/api/v1/wx/component/ticket` |
| 授权事件接收 URL | `https://fujian.5g6g.top/api/v1/wx/component/notify/$APPID$` |
| 授权回调域名 | `fujian.5g6g.top` |

Token / EncodingAESKey 三方同 `.env` 保持一致。启用第三方平台后每 10 分钟一次 ticket 推送。

---

## 常见问题

**Q1: 脚本报 `go: command not found`**  
A: 检查 `/usr/local/go/bin/go version`。旧项目已装过的话直接能用。

**Q2: 端口 8080 被占用**  
A: `lsof -i:8080` 看占用进程；多半是旧 PM2 没清干净，`pm2 delete fujiantong` 后再 `bash scripts/deploy.sh`。

**Q3: 接管后旧 `fujiantong_v2` 数据库要不要删？**  
A: 不影响新项目，留着做历史快照即可。要删就 `DROP DATABASE fujiantong_v2;`。

**Q4: pm2 reload 后旧进程卡死**  
A: `pm2 delete fujiantong && bash scripts/deploy.sh`

**Q5: admin 页面 404**  
A: 检查 `server/public/index.html`。不存在说明前端没编译，运行 `bash scripts/deploy.sh frontend`。

**Q6: 数据库连不上**  
A: `.env` 的 DB_DSN 用 `127.0.0.1:3306`；MySQL 默认只允许 localhost 连接。
