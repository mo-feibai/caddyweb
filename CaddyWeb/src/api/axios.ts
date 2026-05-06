import axios, { AxiosInstance } from 'axios'

interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

api.interceptors.request.use(
  (config) => {
    return config
  },
  (error) => Promise.reject(error)
)

// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(api.interceptors.response.use as any)(
  (response: any) => {
    const res = response.data as ApiResponse
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(res)
    }
    return res.data
  },
  (error: any) => {
    const message =
      error.response?.data?.message || error.message || '请求失败'
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

export async function get<T>(url: string, config?: Parameters<typeof api.get>[1]): Promise<T> {
  return api.get<T>(url, config) as Promise<T>
}

export async function post<T>(url: string, data?: unknown): Promise<T> {
  return api.post<T>(url, data) as Promise<T>
}

export async function put<T>(url: string, data?: unknown): Promise<T> {
  return api.put<T>(url, data) as Promise<T>
}

export async function patch<T>(url: string, data?: unknown): Promise<T> {
  return api.patch<T>(url, data) as Promise<T>
}

export async function del<T>(url: string): Promise<T> {
  return api.delete<T>(url) as Promise<T>
}

export { api }
