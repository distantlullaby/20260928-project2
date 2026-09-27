import axios from 'axios'
import { useAuth } from '../store/auth'

const http = axios.create({ baseURL: '/' })

// 请求自动带上 Token
http.interceptors.request.use((config) => {
  const auth = useAuth()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

// 统一错误提示
http.interceptors.response.use(
  (res) => res,
  (err) => {
    const msg = err.response?.data?.error || '网络开小差了，请稍后再试'
    err.message = msg
    if (err.response?.status === 401) {
      const auth = useAuth()
      auth.logout()
    }
    return Promise.reject(err)
  }
)

export default http
