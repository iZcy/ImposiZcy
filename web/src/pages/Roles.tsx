import { useState } from 'react'
import { PageHeader } from '@kzcy/dashboard'
import Modal from '../components/Modal'
import { rolesApi, type Role } from '../services/api'

export default function Roles() {
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: '', permissions: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [roles, setRoles] = useState<Role[]>([])
  const [activeTab, setActiveTab] = useState<'internal' | 'external'>('internal')

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const permissions = form.permissions.split(',').map(p => p.trim()).filter(Boolean)
      if (activeTab === 'internal') {
        await rolesApi.createInternal({ name: form.name, permissions })
        const res = await rolesApi.listInternal()
        setRoles(res.data || [])
      } else {
        await rolesApi.createExternal({ name: form.name, permissions })
        const res = await rolesApi.listExternal()
        setRoles(res.data || [])
      }
      setShowCreate(false)
      setForm({ name: '', permissions: '' })
    } catch (err: any) {
      setError(err.message || 'Failed to create role')
    } finally {
      setLoading(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Are you sure you want to delete this role?')) return
    try {
      if (activeTab === 'internal') {
        await rolesApi.deleteInternal(id)
        const res = await rolesApi.listInternal()
        setRoles(res.data || [])
      } else {
        await rolesApi.deleteExternal(id)
        const res = await rolesApi.listExternal()
        setRoles(res.data || [])
      }
    } catch (err: any) {
      alert(err.message || 'Failed to delete role')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <PageHeader title="Roles & Access Control" description="Manage internal and external roles" />
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
          + Create Role
        </button>
      </div>

      <div style={{ display: 'flex', gap: '8px', marginTop: '16px', marginBottom: '16px' }}>
        <button
          onClick={() => setActiveTab('internal')}
          style={{
            padding: '8px 16px',
            borderRadius: '6px',
            border: 'none',
            background: activeTab === 'internal' ? 'var(--primary-color)' : 'var(--card-bg)',
            color: activeTab === 'internal' ? '#fff' : 'var(--text)',
            cursor: 'pointer',
            fontSize: '0.85rem',
          }}
        >
          Internal Roles
        </button>
        <button
          onClick={() => setActiveTab('external')}
          style={{
            padding: '8px 16px',
            borderRadius: '6px',
            border: 'none',
            background: activeTab === 'external' ? 'var(--primary-color)' : 'var(--card-bg)',
            color: activeTab === 'external' ? '#fff' : 'var(--text)',
            cursor: 'pointer',
            fontSize: '0.85rem',
          }}
        >
          External Roles
        </button>
      </div>

      {roles.length === 0 ? (
        <div className="text-center py-12" style={{ color: 'var(--text-muted)' }}>
          No {activeTab} roles found. Create one to get started.
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '2px solid var(--border-color)' }}>
                <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Name</th>
                <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Permissions</th>
                <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Created</th>
                <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {roles.map((r) => (
                <tr key={r._id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                  <td style={{ padding: '12px', color: 'var(--text)' }}>{r.name}</td>
                  <td style={{ padding: '12px', color: 'var(--text-muted)' }}>
                    {r.permissions?.join(', ') || '-'}
                  </td>
                  <td style={{ padding: '12px', color: 'var(--text-muted)', fontSize: '0.85rem' }}>
                    {new Date(r.created_at).toLocaleDateString()}
                  </td>
                  <td style={{ padding: '12px', textAlign: 'right' }}>
                    <button
                      onClick={() => handleDelete(r._id)}
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

      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title={`Create ${activeTab} Role`}>
        <form onSubmit={handleCreate}>
          {error && (
            <div style={{ padding: '10px 14px', borderRadius: '6px', background: 'var(--danger-bg)', color: 'var(--danger-color)', marginBottom: '16px', fontSize: '0.85rem' }}>
              {error}
            </div>
          )}

          <div style={{ marginBottom: '16px' }}>
            <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
              Role Name *
            </label>
            <input
              type="text"
              value={form.name}
              onChange={e => setForm({ ...form, name: e.target.value })}
              required
              style={{
                width: '100%',
                padding: '10px 14px',
                borderRadius: '6px',
                border: '1px solid var(--border-color)',
                background: 'var(--input-bg)',
                color: 'var(--text)',
                fontSize: '0.9rem',
                outline: 'none',
              }}
              placeholder="e.g., admin, editor, viewer"
            />
          </div>

          <div style={{ marginBottom: '20px' }}>
            <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500, color: 'var(--text)' }}>
              Permissions (comma-separated)
            </label>
            <input
              type="text"
              value={form.permissions}
              onChange={e => setForm({ ...form, permissions: e.target.value })}
              style={{
                width: '100%',
                padding: '10px 14px',
                borderRadius: '6px',
                border: '1px solid var(--border-color)',
                background: 'var(--input-bg)',
                color: 'var(--text)',
                fontSize: '0.9rem',
                outline: 'none',
              }}
              placeholder="e.g., templates:read, templates:write, images:read"
            />
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
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
                fontSize: '0.9rem',
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              style={{
                padding: '8px 20px',
                borderRadius: '6px',
                border: 'none',
                background: 'var(--primary-color)',
                color: '#fff',
                cursor: loading ? 'not-allowed' : 'pointer',
                fontSize: '0.9rem',
                fontWeight: 500,
                opacity: loading ? 0.7 : 1,
              }}
            >
              {loading ? 'Creating...' : 'Create Role'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
