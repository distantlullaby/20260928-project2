import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 开发环境把 /api 与 /uploads 代理到 Go 服务（:8081）
export default defineConfig({
  plugins: [vue()],
  server: {
    host: '127.0.0.1',
    port: 5180,
    strictPort: true,
    proxy: {
      '/api': 'http://127.0.0.1:8081',
      '/uploads': 'http://127.0.0.1:8081'
    }
  }
})
