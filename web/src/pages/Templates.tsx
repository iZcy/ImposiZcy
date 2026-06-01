import { useEffect, useState } from 'react'
import { PageHeader } from '@kzcy/dashboard'
import { templatesApi, type Template as TemplateType } from '../services/api'
import Modal from '../components/Modal'

interface TemplateVariable {
  name: string
  type: 'text' | 'barcode'
  format?: string
  required: boolean
  default_value?: string
}

const initialForm = {
  name: '',
  slug: '',
  description: '',
  html: '',
  css: '',
  data_schema: '{}',
  width: 800,
  height: 600,
  dimension_unit: 'px',
  dpi: 72,
  output_format: 'png',
  quality: 90,
  tags: '',
  is_active: true,
  variables: [] as TemplateVariable[],
}

export default function Templates() {
  const [templates, setTemplates] = useState<TemplateType[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState(initialForm)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    loadTemplates()
  }, [])

  async function loadTemplates() {
    try {
      const res = await templatesApi.list()
      setTemplates(res.data || [])
    } catch (error) {
      console.error('Failed to load templates:', error)
    } finally {
      setLoading(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Are you sure you want to delete this template?')) return
    try {
      await templatesApi.delete(id)
      loadTemplates()
    } catch (error) {
      console.error('Failed to delete template:', error)
    }
  }

  function addVariable() {
    setForm(prev => ({
      ...prev,
      variables: [...prev.variables, { name: '', type: 'text', format: 'code128', required: true, default_value: '' }],
    }))
  }

  function updateVariable(index: number, field: keyof TemplateVariable, value: any) {
    setForm(prev => {
      const vars = [...prev.variables]
      vars[index] = { ...vars[index], [field]: value }
      return { ...prev, variables: vars }
    })
  }

  function removeVariable(index: number) {
    setForm(prev => ({
      ...prev,
      variables: prev.variables.filter((_, i) => i !== index),
    }))
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    try {
      const payload = {
        ...form,
        width: Number(form.width),
        height: Number(form.height),
        dpi: Number(form.dpi),
        quality: Number(form.quality),
        tags: form.tags.split(',').map(t => t.trim()).filter(Boolean),
        is_active: form.is_active,
        variables: form.variables.filter(v => v.name.trim() !== ''),
      }
      await templatesApi.create(payload)
      setShowCreate(false)
      setForm(initialForm)
      loadTemplates()
    } catch (error) {
      console.error('Failed to create template:', error)
      alert('Failed to create template')
    } finally {
      setSaving(false)
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
        <PageHeader title="Templates" description="Manage printing templates" />
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
          + Create Template
        </button>
      </div>

      <div className="grid grid-cols-1 gap-4 mt-6">
        {templates.length === 0 ? (
          <div className="text-center py-12" style={{ color: 'var(--text-muted)' }}>
            No templates found. Click "Create Template" to add one.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '2px solid var(--border-color)' }}>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Name</th>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Slug</th>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Format</th>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Size</th>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Vars</th>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Status</th>
                  <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {templates.map((t: TemplateType) => (
                  <tr key={t._id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                    <td style={{ padding: '12px', color: 'var(--text)' }}>{t.name}</td>
                    <td style={{ padding: '12px', color: 'var(--text-muted)', fontFamily: 'monospace' }}>{t.slug}</td>
                    <td style={{ padding: '12px' }}>
                      <span style={{
                        padding: '2px 8px',
                        borderRadius: '4px',
                        fontSize: '0.75rem',
                        background: 'var(--badge-info-bg)',
                        color: 'var(--badge-info-color)',
                      }}>
                        {t.format}
                      </span>
                    </td>
                    <td style={{ padding: '12px', color: 'var(--text-muted)' }}>{t.width}×{t.height}</td>
                    <td style={{ padding: '12px', color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                      {t.variables?.length || 0}
                    </td>
                    <td style={{ padding: '12px' }}>
                      <span style={{
                        padding: '2px 8px',
                        borderRadius: '4px',
                        fontSize: '0.75rem',
                        background: t.status === 'active' ? 'var(--badge-success-bg)' : 'var(--badge-warning-bg)',
                        color: t.status === 'active' ? 'var(--badge-success-color)' : 'var(--badge-warning-color)',
                      }}>
                        {t.status}
                      </span>
                    </td>
                    <td style={{ padding: '12px', textAlign: 'right' }}>
                      <a
                        href={`/templates/${t._id}/editor`}
                        style={{
                          padding: '4px 12px',
                          borderRadius: '4px',
                          border: '1px solid var(--primary-color, #2563eb)',
                          color: 'var(--primary-color, #2563eb)',
                          background: 'transparent',
                          cursor: 'pointer',
                          fontSize: '0.8rem',
                          marginRight: '8px',
                          textDecoration: 'none',
                          display: 'inline-block',
                        }}
                      >
                        🎨 Layout
                      </a>
                      <button
                        onClick={() => handleDelete(t._id)}
                        style={{
                          padding: '4px 12px',
                          borderRadius: '4px',
                          border: '1px solid var(--danger-color)',
                          color: 'var(--danger-color)',
                          background: 'transparent',
                          cursor: 'pointer',
                          fontSize: '0.8rem',
                        }}
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="Create Template">
        <form onSubmit={handleCreate}>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>Name *</label>
              <input
                type="text"
                value={form.name}
                onChange={e => setForm({ ...form, name: e.target.value })}
                required
                style={inputStyle}
              />
            </div>
            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>Slug *</label>
              <input
                type="text"
                value={form.slug}
                onChange={e => setForm({ ...form, slug: e.target.value })}
                required
                style={inputStyle}
              />
            </div>
            <div className="col-span-2">
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>Description</label>
              <input
                type="text"
                value={form.description}
                onChange={e => setForm({ ...form, description: e.target.value })}
                style={inputStyle}
              />
            </div>
            <div className="col-span-2">
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>HTML *</label>
              <textarea
                value={form.html}
                onChange={e => setForm({ ...form, html: e.target.value })}
                required
                rows={4}
                placeholder='<div class="label">{{barcode}}<div class="text">{{product_name}}</div></div>'
                style={{ ...inputStyle, fontFamily: 'monospace', fontSize: '0.8rem' }}
              />
              <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '4px' }}>
                Use {'{{variable_name}}'} for text and {'{{barcode}}'} for barcode images
              </p>
            </div>
            <div className="col-span-2">
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>CSS</label>
              <textarea
                value={form.css}
                onChange={e => setForm({ ...form, css: e.target.value })}
                rows={3}
                style={{ ...inputStyle, fontFamily: 'monospace', fontSize: '0.8rem' }}
              />
            </div>
            <div className="col-span-2">
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>Data Schema (JSON) *</label>
              <textarea
                value={form.data_schema}
                onChange={e => setForm({ ...form, data_schema: e.target.value })}
                required
                rows={3}
                style={{ ...inputStyle, fontFamily: 'monospace', fontSize: '0.8rem' }}
              />
            </div>
            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>Width *</label>
              <input
                type="number"
                value={form.width}
                onChange={e => setForm({ ...form, width: Number(e.target.value) })}
                required
                style={inputStyle}
              />
            </div>
            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>Height *</label>
              <input
                type="number"
                value={form.height}
                onChange={e => setForm({ ...form, height: Number(e.target.value) })}
                required
                style={inputStyle}
              />
            </div>
            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>Unit</label>
              <select
                value={form.dimension_unit}
                onChange={e => setForm({ ...form, dimension_unit: e.target.value })}
                style={inputStyle}
              >
                <option value="px">px</option>
                <option value="mm">mm</option>
                <option value="cm">cm</option>
                <option value="in">in</option>
              </select>
            </div>
            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>DPI</label>
              <input
                type="number"
                value={form.dpi}
                onChange={e => setForm({ ...form, dpi: Number(e.target.value) })}
                style={inputStyle}
              />
            </div>
            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>Output Format</label>
              <select
                value={form.output_format}
                onChange={e => setForm({ ...form, output_format: e.target.value })}
                style={inputStyle}
              >
                <option value="png">PNG</option>
                <option value="jpeg">JPEG</option>
                <option value="webp">WebP</option>
                <option value="pdf">PDF</option>
              </select>
            </div>
            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>Quality (1-100)</label>
              <input
                type="number"
                min={1}
                max={100}
                value={form.quality}
                onChange={e => setForm({ ...form, quality: Number(e.target.value) })}
                style={inputStyle}
              />
            </div>
            <div className="col-span-2">
              <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '4px' }}>Tags (comma separated)</label>
              <input
                type="text"
                value={form.tags}
                onChange={e => setForm({ ...form, tags: e.target.value })}
                placeholder="tag1, tag2, tag3"
                style={inputStyle}
              />
            </div>
            <div className="col-span-2">
              <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '0.9rem', color: 'var(--text)', cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={form.is_active}
                  onChange={e => setForm({ ...form, is_active: e.target.checked })}
                />
                Active
              </label>
            </div>
          </div>

          {/* Variables Section */}
          <div className="col-span-2" style={{ marginTop: '20px', borderTop: '1px solid var(--border-color)', paddingTop: '16px' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
              <label style={{ fontSize: '0.9rem', fontWeight: 600, color: 'var(--text)' }}>Template Variables</label>
              <button
                type="button"
                onClick={addVariable}
                style={{
                  padding: '4px 12px',
                  borderRadius: '4px',
                  border: '1px solid var(--primary-color)',
                  background: 'transparent',
                  color: 'var(--primary-color)',
                  cursor: 'pointer',
                  fontSize: '0.8rem',
                }}
              >
                + Add Variable
              </button>
            </div>
            {form.variables.length === 0 && (
              <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontStyle: 'italic' }}>
                No variables defined. Add variables to map barcode/text data from Kafka/API inputs.
              </p>
            )}
            {form.variables.map((v, i) => (
              <div key={i} style={{ display: 'grid', gridTemplateColumns: '2fr 1fr 1fr 1fr auto', gap: '8px', alignItems: 'end', marginBottom: '8px', padding: '8px', borderRadius: '6px', background: 'var(--bg-secondary)' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.7rem', color: 'var(--text-muted)', marginBottom: '2px' }}>Name</label>
                  <input
                    type="text"
                    value={v.name}
                    onChange={e => updateVariable(i, 'name', e.target.value)}
                    placeholder="e.g. barcode"
                    style={{ ...inputStyle, fontSize: '0.8rem', padding: '4px 8px' }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.7rem', color: 'var(--text-muted)', marginBottom: '2px' }}>Type</label>
                  <select
                    value={v.type}
                    onChange={e => updateVariable(i, 'type', e.target.value)}
                    style={{ ...inputStyle, fontSize: '0.8rem', padding: '4px 8px' }}
                  >
                    <option value="text">Text</option>
                    <option value="barcode">Barcode</option>
                  </select>
                </div>
                {v.type === 'barcode' && (
                  <div>
                    <label style={{ display: 'block', fontSize: '0.7rem', color: 'var(--text-muted)', marginBottom: '2px' }}>Format</label>
                    <select
                      value={v.format || 'code128'}
                      onChange={e => updateVariable(i, 'format', e.target.value)}
                      style={{ ...inputStyle, fontSize: '0.8rem', padding: '4px 8px' }}
                    >
                      <option value="code128">Code 128</option>
                      <option value="qr">QR Code</option>
                      <option value="ean13">EAN-13</option>
                      <option value="upc">UPC</option>
                      <option value="code39">Code 39</option>
                    </select>
                  </div>
                )}
                {v.type === 'text' && (
                  <div>
                    <label style={{ display: 'block', fontSize: '0.7rem', color: 'var(--text-muted)', marginBottom: '2px' }}>Default</label>
                    <input
                      type="text"
                      value={v.default_value || ''}
                      onChange={e => updateVariable(i, 'default_value', e.target.value)}
                      placeholder="optional"
                      style={{ ...inputStyle, fontSize: '0.8rem', padding: '4px 8px' }}
                    />
                  </div>
                )}
                <div>
                  <label style={{ display: 'block', fontSize: '0.7rem', color: 'var(--text-muted)', marginBottom: '2px' }}>Required</label>
                  <input
                    type="checkbox"
                    checked={v.required}
                    onChange={e => updateVariable(i, 'required', e.target.checked)}
                    style={{ margin: '4px 0' }}
                  />
                </div>
                <button
                  type="button"
                  onClick={() => removeVariable(i)}
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
                  ✕
                </button>
              </div>
            ))}
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', marginTop: '24px' }}>
            <button
              type="button"
              onClick={() => setShowCreate(false)}
              style={{
                padding: '8px 16px',
                borderRadius: '6px',
                border: '1px solid var(--border-color)',
                background: 'transparent',
                color: 'var(--text)',
                cursor: 'pointer',
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              style={{
                padding: '8px 16px',
                borderRadius: '6px',
                border: 'none',
                background: 'var(--primary-color)',
                color: '#fff',
                cursor: saving ? 'not-allowed' : 'pointer',
                opacity: saving ? 0.7 : 1,
              }}
            >
              {saving ? 'Creating...' : 'Create'}
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
