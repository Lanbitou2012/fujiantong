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

# 自动定位项目根目录（脚本所在目录的上一级）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
GO_BIN="${GO_BIN:-$(command -v go || echo /usr/local/go/bin/go)}"
APP_NAME="fujiantong"
MODE="${1:-all}"

# 修复 git "dubious ownership"
git config --global --add safe.directory "$PROJECT_DIR" 2>/dev/null || true

cd "$PROJECT_DIR"
echo "==> 项目目录: $PROJECT_DIR"

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

  # 修复某些 npm 镜像/Linux 文件系统下 .bin 缺失可执行位的问题
  if [ -d admin/node_modules/.bin ]; then
    chmod -R +x admin/node_modules/.bin 2>/dev/null || true
  fi

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
