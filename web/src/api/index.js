import http from './http'

export const register = (data) => http.post('/api/register', data).then((r) => r.data)
export const login = (data) => http.post('/api/login', data).then((r) => r.data)
export const getMe = () => http.get('/api/me').then((r) => r.data)

export const listWishes = (params) =>
  http.get('/api/wishes', { params }).then((r) => r.data)
export const getWish = (id) => http.get(`/api/wishes/${id}`).then((r) => r.data)
export const createWish = (data) => http.post('/api/wishes', data).then((r) => r.data)
export const addBounty = (id, amount) =>
  http.post(`/api/wishes/${id}/bounty`, { amount }).then((r) => r.data)
export const cancelWish = (id) => http.post(`/api/wishes/${id}/cancel`).then((r) => r.data)
export const acceptWish = (id, responseId) =>
  http.post(`/api/wishes/${id}/accept/${responseId}`).then((r) => r.data)

export const createResponse = (id, data) =>
  http.post(`/api/wishes/${id}/responses`, data).then((r) => r.data)

export const myCoins = (params) => http.get('/api/my/coins', { params }).then((r) => r.data)
export const myWishes = () => http.get('/api/my/wishes').then((r) => r.data)
export const myResponses = () => http.get('/api/my/responses').then((r) => r.data)

export const uploadImage = (file) => {
  const fd = new FormData()
  fd.append('file', file)
  return http
    .post('/api/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
    .then((r) => r.data)
}
