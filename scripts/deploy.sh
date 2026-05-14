#!/usr/bin/env bash
# ===========================================================================
# 附件通 一键部署脚本（在服务器上执行）
# 完全接管旧「公众号附件助手」的部署位置（域名/路径/端口/PM2 全沿用）
# 数据库换为新库 fujiantong
#
# 用法：
#   cd /www/wwwroot/fujiantong
#   bash scripts/deploy.sh           # 完整：pull + 编译后端 + 编译前端 + 重启
#   bash scripts/deploy.sh backend   # 只更新后端
#   bash scripts/deploy.sh frontend  # 只更新前端
# ===========================================================================
set -e

PROJECT_DIR="/www/wwwroot/fujiantong"
GO_BIN="/usr/local/go/bin/go"
APP_NAME="fujiantong"
MODE="${1:-all}"

cd "$PROJECT_DIR"

echo "==> [1/5] git pull"
git fetch origin
git reset --hard origin/main

if [[ "$MODE" == "all" || "$MODE" == "backend" ]]; then
  echo "==> [2/5] 编译 Go 后端"
  cd "$PROJECT_DIR/server"
  $GO_BIN env -w GOPROXY=https://goproxy.cn,direct
  $GO_BIN mod tidy
  # 先编译到 .new，原子替换，避免 pm2 仍在运行旧进程时锁文件
  $GO_BIN build -o fujian-tong.new ./cmd/api
  mv -f fujian-tong.new fujian-tong
  chmod +x fujian-tong
  cd "$PROJECT_DIR"
fi

if [[ "$MODE" == "all" || "$MODE" == "frontend" ]]; then
  echo "==> [3/5] 安装 admin 依赖"
  npm install --prefix admin --silent --no-audit --no-fund

  echo "==> [4/5] 编译 admin（输出到 server/public/）"
  npm run build --prefix admin
fi

echo "==> [5/5] 通过 PM2 重启后端"
if pm2 describe "$APP_NAME" > /dev/null 2>&1; then
  pm2 reload "$APP_NAME" --update-env
else
  pm2 start "$PROJECT_DIR/scripts/ecosystem.config.js"
fi
pm2 save

echo
echo "✅ 部署完成"
pm2 list | grep -E "(App name|$APP_NAME)" || true
