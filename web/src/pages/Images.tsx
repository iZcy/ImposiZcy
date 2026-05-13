import { useState, useEffect } from 'react'
import { PageHeader } from '@kzcy/dashboard'
import { imagesApi, templatesApi, type Template } from '../services/api'
import Modal from '../components/Modal'

interface TemplateVariable {
  name: string
  type: string
  format?: string
  required?: boolean
  default_value?: string
}

export default function Images() {
  const [images, setImages] = useState<any[]>([])
  const [templates, setTemplates] = useState<Template[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [creating, setCreating] = useState(false)
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null)
  const [variableValues, setVariableValues] = useState<Record<string, string>>({})
  const [outputFormat, setOutputFormat] = useState('png')
  const [width, setWidth] = useState(0)
  const [height, setHeight] = useState(0)

  useEffect(() => {
    loadImages()
    loadTemplates()
  }, [])

  async function loadImages() {
    try {
      const res = await imagesApi.list()
      setImages(res.data || [])
    } catch (err) {
      console.error('Failed to load images:', err)
    } finally {
      setLoading(false)
    }
  }

  async function loadTemplates() {
    try {
      const res = await templatesApi.list()
      setTemplates(res.data || [])
    } catch (err) {
      console.error('Failed to load templates:', err)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Are you sure you want to delete this image?')) return
    try {
      await imagesApi.delete(id)
      setImages(images.filter(img => img.id !== id))
    } catch (err) {
      console.error('Failed to delete image:', err)
    }
  }

  function handleTemplateChange(templateId: string) {
    const tmpl = templates.find(t => t._id === templateId || t.id === templateId)
    setSelectedTemplate(tmpl || null)
    setWidth(tmpl?.width || 0)
    setHeight(tmpl?.height || 0)

    // Initialize variable values with defaults
    const vars: TemplateVariable[] = tmpl?.variables || []
    const defaults: Record<string, string> = {}
    vars.forEach((v: TemplateVariable) => {
      defaults[v.name] = v.default_value || ''
    })
    setVariableValues(defaults)
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!selectedTemplate) {
      alert('Please select a template')
      return
    }

    // Build data object from variable values
    const data: Record<string, any> = {}
    const vars: TemplateVariable[] = selectedTemplate.variables || []
    for (const v of vars) {
      const val = variableValues[v.name]
      if (v.required && (!val || val.trim() === '')) {
        alert(`Variable "${v.name}" is required`)
        return
      }
      data[v.name] = val || ''
    }

    setCreating(true)
    try {
      const res = await fetch('/api/v1/render', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          template_slug: selectedTemplate.slug,
          data,
          output_format: outputFormat,
          width: width || undefined,
          height: height || undefined,
        }),
      })
      const result = await res.json()
      if (result.success) {
        setShowCreate(false)
        setSelectedTemplate(null)
        setVariableValues({})
        loadImages()
        alert('Image generated successfully!')
      } else {
        alert(result.error || 'Failed to generate image')
      }
    } catch (err: any) {
      alert(err.message || 'Failed to generate image')
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
        <PageHeader title="Images" description="Generated images from templates" />
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
          + Generate Image
        </button>
      </div>

      {images.length === 0 ? (
        <div className="text-center py-12" style={{ color: 'var(--text-muted)' }}>
          <div className="text-4xl mb-4">🖼️</div>
          <p>No images generated yet.</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {images.map((img: any) => (
            <div
              key={img._id || img.id}
              style={{
                borderRadius: '8px',
                border: '1px solid var(--border-color)',
                overflow: 'hidden',
                background: 'var(--card-bg)',
              }}
            >
              <div style={{ aspectRatio: '1', background: 'var(--bg-secondary)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                {img.base64 ? (
                  <img src={`data:image/${img.format || 'png'};base64,${img.base64}`} alt={img.filename} style={{ maxWidth: '100%', maxHeight: '100%' }} />
                ) : (
                  <span style={{ fontSize: '2rem' }}>🖼️</span>
                )}
              </div>
              <div style={{ padding: '12px' }}>
                <p style={{ fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)', marginBottom: '4px' }}>
                  {img.filename}
                </p>
                <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                  {img.format?.toUpperCase()} · {img.width}×{img.height}
                </p>
                <div style={{ display: 'flex', gap: '8px', marginTop: '8px' }}>
                  <a
                    href={`/api/v1/images/${img._id || img.id}/download`}
                    target="_blank"
                    rel="noopener noreferrer"
                    style={{
                      flex: 1,
                      padding: '4px 8px',
                      borderRadius: '4px',
                      border: '1px solid var(--primary-color)',
                      color: 'var(--primary-color)',
                      background: 'transparent',
                      cursor: 'pointer',
                      fontSize: '0.75rem',
                      textAlign: 'center',
                      textDecoration: 'none',
                    }}
                  >
                    Download
                  </a>
                  <button
                    onClick={() => handleDelete(img._id || img.id)}
                    style={{
                      padding: '4px 8px',
                      borderRadius: '4px',
                      border: '1px solid var(--danger-color)',
                      color: 'var(--danger-color)',
                      background: 'transparent',
                      cursor: 'pointer',
                      fontSize: '0.75rem',
                    }}
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="Generate Image">
        <form onSubmit={handleCreate}>
          <div style={{ marginBottom: '16px' }}>
            <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
              Template *
            </label>
            <select
              value={selectedTemplate?._id || selectedTemplate?.id || ''}
              onChange={e => handleTemplateChange(e.target.value)}
              required
              style={{
                width: '100%',
                padding: '10px 12px',
                borderRadius: '6px',
                border: '1px solid var(--border-color)',
                background: 'var(--input-bg)',
                color: 'var(--text)',
                fontSize: '0.9rem',
              }}
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

          {/* Dynamic Variable Inputs */}
          {selectedTemplate && (selectedTemplate.variables || []).length > 0 && (
            <div style={{ marginBottom: '16px' }}>
              <label style={{ display: 'block', marginBottom: '8px', fontSize: '0.9rem', fontWeight: 600, color: 'var(--text)' }}>
                Template Variables
              </label>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {(selectedTemplate.variables || []).map((v: TemplateVariable) => (
                  <div key={v.name}>
                    <label style={{ display: 'block', marginBottom: '4px', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                      {v.name} {v.required && <span style={{ color: 'var(--danger-color)' }}>*</span>}
                      <span style={{ marginLeft: '8px', fontSize: '0.7rem', padding: '1px 6px', borderRadius: '4px', background: 'var(--badge-info-bg)', color: 'var(--badge-info-color)' }}>
                        {v.type}{v.format ? ` (${v.format})` : ''}
                      </span>
                    </label>
                    <input
                      type="text"
                      value={variableValues[v.name] || ''}
                      onChange={e => setVariableValues(prev => ({ ...prev, [v.name]: e.target.value }))}
                      placeholder={v.default_value || `Enter ${v.name}`}
                      style={{
                        width: '100%',
                        padding: '8px 12px',
                        borderRadius: '6px',
                        border: '1px solid var(--border-color)',
                        background: 'var(--input-bg)',
                        color: 'var(--text)',
                        fontSize: '0.9rem',
                      }}
                    />
                  </div>
                ))}
              </div>
            </div>
          )}

          {selectedTemplate && (selectedTemplate.variables || []).length === 0 && (
            <div style={{ marginBottom: '16px', padding: '10px', borderRadius: '6px', background: 'var(--badge-warning-bg)', fontSize: '0.8rem', color: 'var(--badge-warning-color)' }}>
              This template has no variables defined. The HTML will be rendered as-is.
            </div>
          )}

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', marginBottom: '16px' }}>
            <div>
              <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
                Output Format
              </label>
              <select
                value={outputFormat}
                onChange={e => setOutputFormat(e.target.value)}
                style={{
                  width: '100%',
                  padding: '10px 12px',
                  borderRadius: '6px',
                  border: '1px solid var(--border-color)',
                  background: 'var(--input-bg)',
                  color: 'var(--text)',
                  fontSize: '0.9rem',
                }}
              >
                <option value="png">PNG</option>
                <option value="jpg">JPG</option>
                <option value="jpeg">JPEG</option>
                <option value="webp">WebP</option>
                <option value="pdf">PDF</option>
              </select>
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
                Width (px)
              </label>
              <input
                type="number"
                value={width || ''}
                onChange={e => setWidth(parseInt(e.target.value) || 0)}
                placeholder="Auto"
                style={{
                  width: '100%',
                  padding: '10px 12px',
                  borderRadius: '6px',
                  border: '1px solid var(--border-color)',
                  background: 'var(--input-bg)',
                  color: 'var(--text)',
                  fontSize: '0.9rem',
                }}
              />
            </div>
          </div>

          <div style={{ marginBottom: '20px' }}>
            <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
              Height (px)
            </label>
            <input
              type="number"
              value={height || ''}
              onChange={e => setHeight(parseInt(e.target.value) || 0)}
              placeholder="Auto"
              style={{
                width: '100%',
                padding: '10px 12px',
                borderRadius: '6px',
                border: '1px solid var(--border-color)',
                background: 'var(--input-bg)',
                color: 'var(--text)',
                fontSize: '0.9rem',
              }}
            />
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
              {creating ? 'Generating...' : 'Generate'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
