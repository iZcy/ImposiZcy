import { useEffect, useState } from 'react'
import { PageHeader } from '@kzcy/dashboard'
import { renderJobsApi, templatesApi, type RenderJob as RenderJobType, type Template } from '../services/api'
import Modal from '../components/Modal'

export default function RenderJobs() {
  const [jobs, setJobs] = useState<RenderJobType[]>([])
  const [templates, setTemplates] = useState<Template[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [creating, setCreating] = useState(false)
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null)
  const [variableValues, setVariableValues] = useState<Record<string, string>>({})
  const [formData, setFormData] = useState({
    template_id: '',
    output_format: 'png',
    width: 0,
    height: 0,
    quality: 90,
    async: false,
  })

  useEffect(() => {
    loadJobs()
    loadTemplates()
  }, [])

  async function loadJobs() {
    try {
      const res = await renderJobsApi.list()
      setJobs(res?.data || [])
    } catch (error) {
      console.error('Failed to load render jobs:', error)
    } finally {
      setLoading(false)
    }
  }

  async function loadTemplates() {
    try {
      const res = await templatesApi.list()
      setTemplates(res?.data || [])
    } catch (error) {
      console.error('Failed to load templates:', error)
    }
  }

  function handleTemplateChange(templateId: string) {
    const tmpl = templates.find(t => t._id === templateId || t.id === templateId)
    setSelectedTemplate(tmpl || null)
    setFormData(prev => ({
      ...prev,
      template_id: templateId,
      width: tmpl?.width || 0,
      height: tmpl?.height || 0,
    }))
    // Initialize variable values
    const initialVars: Record<string, string> = {}
    if (tmpl?.variables) {
      tmpl.variables.forEach((v: any) => {
        initialVars[v.name] = v.default_value || ''
      })
    }
    setVariableValues(initialVars)
  }

  function getVariableInputType(variable: any): string {
    switch (variable.type) {
      case 'number': return 'number'
      case 'date': return 'date'
      case 'email': return 'email'
      default: return 'text'
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!formData.template_id) {
      alert('Please select a template')
      return
    }
    if (selectedTemplate && selectedTemplate.variables && selectedTemplate.variables.length > 0) {
      const missing = selectedTemplate.variables.filter((v: any) => v.required && !variableValues[v.name])
      if (missing.length > 0) {
        alert(`Missing required variables: ${missing.map((v: any) => v.name).join(', ')}`)
        return
      }
    }

    setCreating(true)
    try {
      const tmpl = templates.find(t => t._id === formData.template_id || t.id === formData.template_id)
      const data: Record<string, any> = {}
      // Convert variable values to proper types
      if (selectedTemplate?.variables) {
        selectedTemplate.variables.forEach((v: any) => {
          const val = variableValues[v.name]
          if (v.type === 'number') {
            data[v.name] = val ? Number(val) : null
          } else {
            data[v.name] = val
          }
        })
      }

      const endpoint = formData.async ? '/api/v1/render/async' : '/api/v1/render'
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          template_slug: tmpl?.slug,
          data,
          output_format: formData.output_format,
          width: formData.width || undefined,
          height: formData.height || undefined,
          quality: formData.quality || undefined,
        }),
      })
      const result = await res.json()
      if (result.success) {
        setShowCreate(false)
        setFormData({ template_id: '', output_format: 'png', width: 0, height: 0, quality: 90, async: false })
        setSelectedTemplate(null)
        setVariableValues({})
        loadJobs()
        alert(formData.async ? `Render job created: ${result.data?.job_id}` : 'Image rendered successfully!')
      } else {
        alert(result.error || 'Failed to render')
      }
    } catch (err: any) {
      alert(err.message || 'Failed to render')
    } finally {
      setCreating(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2" style={{ borderColor: 'var(--primary-color)' }} />
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <PageHeader title="Render Jobs" />
        <button
          onClick={() => setShowCreate(true)}
          style={{
            padding: '8px 20px',
            borderRadius: '6px',
            border: 'none',
            background: 'var(--primary-color)',
            color: '#fff',
            cursor: 'pointer',
            fontSize: '0.9rem',
            fontWeight: 500,
          }}
        >
          + New Render Job
        </button>
      </div>

      {jobs.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
          No render jobs yet.
        </div>
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '2px solid var(--border-color)' }}>
                <th style={{ padding: '0.75rem', textAlign: 'left', color: 'var(--text-secondary)' }}>ID</th>
                <th style={{ padding: '0.75rem', textAlign: 'left', color: 'var(--text-secondary)' }}>Template</th>
                <th style={{ padding: '0.75rem', textAlign: 'left', color: 'var(--text-secondary)' }}>Status</th>
                <th style={{ padding: '0.75rem', textAlign: 'left', color: 'var(--text-secondary)' }}>Created</th>
              </tr>
            </thead>
            <tbody>
              {jobs.map((job) => (
                <tr key={job._id} style={{ borderBottom: '1px solid var(--border-light)' }}>
                  <td style={{ padding: '0.75rem', color: 'var(--text)', fontFamily: 'monospace', fontSize: '0.85rem' }}>
                    {job._id?.slice(0, 8)}...
                  </td>
                  <td style={{ padding: '0.75rem', color: 'var(--text)' }}>{job.template_id?.slice(0, 8)}...</td>
                  <td style={{ padding: '0.75rem' }}>
                    <span style={{
                      padding: '0.25rem 0.5rem',
                      borderRadius: '0.25rem',
                      fontSize: '0.8rem',
                      background: job.status === 'completed' ? 'var(--badge-success-bg)' :
                                  job.status === 'failed' ? 'var(--badge-danger-bg)' :
                                  'var(--badge-info-bg)',
                      color: job.status === 'completed' ? 'var(--success-color)' :
                             job.status === 'failed' ? 'var(--danger-color)' :
                             'var(--primary-color)',
                    }}>
                      {job.status}
                    </span>
                  </td>
                  <td style={{ padding: '0.75rem', color: 'var(--text-muted)', fontSize: '0.85rem' }}>
                    {new Date(job.created_at).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="New Render Job">
        <form onSubmit={handleCreate}>
          <div style={{ marginBottom: '16px' }}>
            <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
              Template *
            </label>
            <select
              value={formData.template_id}
              onChange={e => handleTemplateChange(e.target.value)}
              required
              style={inputStyle}
            >
              <option value="">Select a template</option>
              {templates.map(t => (
                <option key={t._id || t.id} value={t._id || t.id}>{t.name}</option>
              ))}
            </select>
          </div>

          {selectedTemplate && (
            <div style={{ marginBottom: '16px', padding: '10px', borderRadius: '6px', background: 'var(--bg-secondary)', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
              <p><strong>Slug:</strong> {selectedTemplate.slug}</p>
              <p><strong>Format:</strong> {selectedTemplate.format}</p>
              <p><strong>Size:</strong> {selectedTemplate.width}×{selectedTemplate.height}</p>
            </div>
          )}

          {selectedTemplate?.variables && selectedTemplate.variables.length > 0 && (
            <div style={{ marginBottom: '16px', padding: '12px', borderRadius: '6px', border: '1px solid var(--border-color)', background: 'var(--bg-secondary)' }}>
              <h4 style={{ margin: '0 0 12px 0', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text)' }}>
                Template Variables
              </h4>
              {selectedTemplate.variables.map((variable: any, index: number) => (
                <div key={index} style={{ marginBottom: '10px' }}>
                  <label style={{ display: 'block', marginBottom: '4px', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                    {variable.name}
                    {variable.required && <span style={{ color: 'var(--danger-color)' }}> *</span>}
                    <span style={{ marginLeft: '6px', fontSize: '0.7rem', padding: '1px 4px', borderRadius: '3px', background: 'var(--border-color)', color: 'var(--text-muted)' }}>
                      {variable.type}
                    </span>
                  </label>
                  {variable.type === 'barcode' ? (
                    <div>
                      <input
                        type="text"
                        value={variableValues[variable.name] || ''}
                        onChange={e => setVariableValues({ ...variableValues, [variable.name]: e.target.value })}
                        placeholder={`Enter ${variable.barcode_format || 'code128'} value...`}
                        required={variable.required}
                        style={inputStyle}
                      />
                      <p style={{ fontSize: '0.7rem', color: 'var(--text-muted)', marginTop: '2px' }}>
                        Format: {variable.barcode_format || 'code128'}
                      </p>
                    </div>
                  ) : (
                    <input
                      type={getVariableInputType(variable)}
                      value={variableValues[variable.name] || ''}
                      onChange={e => setVariableValues({ ...variableValues, [variable.name]: e.target.value })}
                      placeholder={variable.description || `Enter ${variable.name}...`}
                      required={variable.required}
                      style={inputStyle}
                    />
                  )}
                </div>
              ))}
            </div>
          )}

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', marginBottom: '16px' }}>
            <div>
              <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
                Output Format
              </label>
              <select
                value={formData.output_format}
                onChange={e => setFormData({ ...formData, output_format: e.target.value })}
                style={inputStyle}
              >
                <option value="png">PNG</option>
                <option value="jpeg">JPEG</option>
                <option value="webp">WebP</option>
              </select>
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
                Quality (1-100)
              </label>
              <input
                type="number"
                min={1}
                max={100}
                value={formData.quality}
                onChange={e => setFormData({ ...formData, quality: parseInt(e.target.value) || 90 })}
                style={inputStyle}
              />
            </div>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', marginBottom: '16px' }}>
            <div>
              <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
                Width (px)
              </label>
              <input
                type="number"
                value={formData.width || ''}
                onChange={e => setFormData({ ...formData, width: parseInt(e.target.value) || 0 })}
                placeholder="Auto"
                style={inputStyle}
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
                Height (px)
              </label>
              <input
                type="number"
                value={formData.height || ''}
                onChange={e => setFormData({ ...formData, height: parseInt(e.target.value) || 0 })}
                placeholder="Auto"
                style={inputStyle}
              />
            </div>
          </div>

          <div style={{ marginBottom: '20px' }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '0.85rem', color: 'var(--text)', cursor: 'pointer' }}>
              <input
                type="checkbox"
                checked={formData.async}
                onChange={e => setFormData({ ...formData, async: e.target.checked })}
              />
              Async mode (create render job instead of immediate response)
            </label>
          </div>

          <div style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end' }}>
            <button
              type="button"
              onClick={() => setShowCreate(false)}
              disabled={creating}
              style={{
                padding: '10px 20px',
                borderRadius: '6px',
                border: '1px solid var(--border-color)',
                background: 'transparent',
                color: 'var(--text)',
                cursor: 'pointer',
                fontSize: '0.9rem',
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={creating}
              style={{
                padding: '10px 20px',
                borderRadius: '6px',
                border: 'none',
                background: 'var(--primary-color)',
                color: '#fff',
                cursor: creating ? 'not-allowed' : 'pointer',
                fontSize: '0.9rem',
                opacity: creating ? 0.7 : 1,
              }}
            >
              {creating ? 'Rendering...' : 'Render'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  )
}

const inputStyle: React.CSSProperties = {
  width: '100%',
  padding: '8px 12px',
  borderRadius: '6px',
  border: '1px solid var(--border-color)',
  background: 'var(--input-bg, var(--card-bg))',
  color: 'var(--text)',
  fontSize: '0.9rem',
  outline: 'none',
}
