import axios from 'axios'
import { auth } from './store'

const http = axios.create({ baseURL: '/' })

// 请求自动带上登录令牌
http.interceptors.request.use((config) => {
  if (auth.token) config.headers.Authorization = `Bearer ${auth.token}`
  return config
})

// 401 时清除登录态并跳转登录页（登录页自身除外）
http.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response && err.response.status === 401) {
      auth.logout()
      if (!location.hash.startsWith('#/login')) {
        location.hash = '#/login'
      }
    }
    return Promise.reject(err)
  }
)

export default http

// 统一提取后端错误信息
export function errMsg(e, fallback = '操作失败，请稍后再试') {
  return e?.response?.data?.error || fallback
}
