import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import path from 'path'

export default defineConfig({
  plugins: [
    vue({
      template: {
        compilerOptions: {
          // 微信开放标签：wx-open-launch-weapp / wx-open-launch-app
          isCustomElement: tag => tag.startsWith('wx-open-'),
        },
      },
    }),
    AutoImport({
      resolvers: [ElementPlusResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    // 与公众号附件助手项目部署模式对齐：前端直接打包到 Go 后端的 public 目录
    // 由 Go Gin 通过 r.Static("/", "./public") 提供静态服务，Nginx 仅反代 8081
    outDir: '../server/public',
    emptyOutDir: true,
  },
})
