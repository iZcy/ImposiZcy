import { configureApi as sharedConfigureApi, ApiError } from '@kzcy/dashboard'

// Configure shared API client
sharedConfigureApi('/dashboard/api')

// Re-export shared ApiError
export { ApiError }

// Custom fetch with 15s timeout
async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 15000)

  try {
    const response = await fetch(`/dashboard/api${path}`, {
      ...options,
      signal: controller.signal,
      headers: {
        'Content-Type': 'application/json',
        ...((options.headers as Record<string, string>) || {}),
      },
    })
    clearTimeout(timeoutId)

    let data: any
    try {
      data = await response.json()
    } catch {
      throw new ApiError(response.status, `HTTP ${response.status}: Invalid response from server`)
    }

    if (!response.ok || data.success === false) {
      throw new ApiError(data.error || `HTTP ${response.status}`, data)
    }

    return data
  } catch (error: any) {
    clearTimeout(timeoutId)
    if (error.name === 'AbortError') {
      throw new ApiError(0, 'Request timeout')
    }
    throw error
  }
}

const post = <T>(path: string, body?: unknown) =>
  request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined })

const put = <T>(path: string, body?: unknown) =>
  request<T>(path, { method: 'PUT', body: body ? JSON.stringify(body) : undefined })

const del = <T>(path: string) =>
  request<T>(path, { method: 'DELETE' })

// Templates
export interface FieldPosition {
  x?: number
  y?: number
  width?: number
  height?: number
  font_size?: number
  alignment?: 'left' | 'center' | 'right'
  color?: string
  bold?: boolean
}

export interface TemplateVariable {
  name: string
  type: 'text' | 'barcode' | 'image'
  barcode_format?: 'code128' | 'qr' | 'ean13' | 'ean8' | 'upca' | 'code39'
  required?: boolean
  default_value?: string
  description?: string
  position?: FieldPosition
  font_size?: number
  font_color?: string
  font_weight?: string
  text_align?: string
  source_field?: string
  font_family?: string
  order_index?: number
}

export interface Template {
  _id: string
  id?: string
  name: string
  slug: string
  description?: string
  html?: string
  css?: string
  data_schema?: string
  variables?: TemplateVariable[]
  field_mapping?: Array<{ source_field: string; target_variable: string; default_value?: string; transform?: string }>
  background_image?: string
  render_engine?: string
  default_font?: string
  width: number
  height: number
  dimension_unit?: string
  dpi?: number
  output_format?: string
  // Back-compat: existing pages read `format` directly. Keep it for now.
  format?: string
  quality?: number
  tags?: any[]
  is_active?: boolean
  status?: string
  created_at?: string
  updated_at?: string
}

export const templatesApi = {
  list: () => request<{ success: boolean; data: Template[] }>('/templates'),
  get: (id: string) => request<{ success: boolean; data: Template }>(`/templates/${id}`),
  create: (data: Partial<Template>) => post<{ success: boolean; data: Template }>('/templates', data),
  update: (id: string, data: Partial<Template>) => put<{ success: boolean; data: Template }>(`/templates/${id}`, data),
  delete: (id: string) => del<{ success: boolean }>(`/templates/${id}`),
}

// Fonts
export interface Font {
  id: string
  family: string
  file_name: string
  path: string
  format: 'ttf' | 'otf'
  scope: 'global' | 'template'
  template_id?: string
  size: number
  created_at: string
}

export const fontsApi = {
  list: (params?: { scope?: string; template_id?: string }) => {
    const q = new URLSearchParams()
    if (params?.scope) q.set('scope', params.scope)
    if (params?.template_id) q.set('template_id', params.template_id)
    const qs = q.toString()
    return request<{ success: boolean; data: Font[] }>(`/fonts${qs ? '?' + qs : ''}`)
  },
  upload: async (file: File, family: string, scope: 'global' | 'template' = 'global', templateId?: string) => {
    const fd = new FormData()
    fd.append('font', file)
    fd.append('family', family)
    fd.append('scope', scope)
    if (templateId) fd.append('template_id', templateId)
    const res = await fetch('/dashboard/api/fonts', { method: 'POST', body: fd })
    const data = await res.json()
    if (!res.ok || data.success === false) throw new Error(data.error || 'Upload failed')
    return data as { success: boolean; data: Font }
  },
  delete: (id: string) => del<{ success: boolean }>(`/fonts/${id}`),
}

// Render Jobs
export interface RenderJob {
  _id: string
  template_id: string
  status: string
  format: string
  width: number
  height: number
  variables?: Record<string, any>
  created_at: string
  updated_at: string
}

export const renderJobsApi = {
  list: () => request<{ success: boolean; data: RenderJob[] }>('/render-jobs'),
  get: (id: string) => request<{ success: boolean; data: RenderJob }>(`/render-jobs/${id}`),
}

// Images
export interface Image {
  _id: string
  render_job_id: string
  filename: string
  format: string
  width: number
  height: number
  file_size: number
  url: string
  created_at: string
}

export const imagesApi = {
  list: () => request<{ success: boolean; data: Image[] }>('/images'),
  delete: (id: string) => del<{ success: boolean }>(`/images/${id}`),
}

// Settings
export const settingsApi = {
  list: () => request<{ success: boolean; data: Record<string, any> }>('/settings'),
  update: (key: string, value: any) => put<{ success: boolean }>(`/settings/${key}`, { value }),
}

// Kafka Connections
export interface KafkaConnection {
  _id: string
  name: string
  brokers: string[]
  topic: string
  group_id: string
  status: string
  created_at: string
}

export const kafkaApi = {
  list: () => request<{ success: boolean; data: KafkaConnection[] }>('/kafka/connections'),
  create: (data: Partial<KafkaConnection>) => post<{ success: boolean; data: KafkaConnection }>('/kafka/connections', data),
  update: (id: string, data: Partial<KafkaConnection>) => put<{ success: boolean }>(`/kafka/connections/${id}`, data),
  delete: (id: string) => del<{ success: boolean }>(`/kafka/connections/${id}`),
  start: (id: string) => post<{ success: boolean }>(`/kafka/connections/${id}/start`),
  stop: (id: string) => post<{ success: boolean }>(`/kafka/connections/${id}/stop`),
}

// Roles
export interface Role {
  _id: string
  name: string
  permissions: string[]
  created_at: string
}

export const rolesApi = {
  listInternal: () => request<{ success: boolean; data: Role[] }>('/roles/internal'),
  createInternal: (data: Partial<Role>) => post<{ success: boolean; data: Role }>('/roles/internal', data),
  updateInternal: (id: string, data: Partial<Role>) => put<{ success: boolean }>(`/roles/internal/${id}`, data),
  deleteInternal: (id: string) => del<{ success: boolean }>(`/roles/internal/${id}`),
  listExternal: () => request<{ success: boolean; data: Role[] }>('/roles/external'),
  createExternal: (data: Partial<Role>) => post<{ success: boolean; data: Role }>('/roles/external', data),
  deleteExternal: (id: string) => del<{ success: boolean }>(`/roles/external/${id}`),
}

// API Keys
export interface APIKey {
  _id: string
  name: string
  key_prefix: string
  permissions: string[]
  created_at: string
}

export const apiKeysApi = {
  list: () => request<{ success: boolean; data: APIKey[] }>('/security/api-keys'),
  create: (data: { name: string; permissions: string[] }) => post<{ success: boolean; data: APIKey & { key: string } }>('/security/api-keys', data),
  delete: (id: string) => del<{ success: boolean }>(`/security/api-keys/${id}`),
}
