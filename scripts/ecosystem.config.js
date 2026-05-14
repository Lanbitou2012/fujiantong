// PM2 守护配置：把 Go 二进制当成普通进程托管。
// 启动命令： pm2 start scripts/ecosystem.config.js
const path = require('path')
const PROJECT_DIR = path.resolve(__dirname, '..')

module.exports = {
  apps: [
    {
      name: 'fujiantong',
      cwd: path.join(PROJECT_DIR, 'server'),
      script: path.join(PROJECT_DIR, 'server', 'fujian-tong'),
      // 关键：以二进制方式启动，不要走 node 解释器
      interpreter: 'none',
      exec_mode: 'fork',
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: '512M',
      // PM2 通过环境变量将 .env 注入；附件通服务自己内部也读 server/.env
      env: {
        // 仅作兜底，真正的密钥放在 server/.env 里
        SERVER_PORT: '8080',
        SERVER_HOST: '127.0.0.1',
        GIN_MODE: 'release'
      },
      out_file: path.join(PROJECT_DIR, 'server', 'logs', 'pm2-out.log'),
      error_file: path.join(PROJECT_DIR, 'server', 'logs', 'pm2-err.log'),
      merge_logs: true,
      time: true
    }
  ]
}
