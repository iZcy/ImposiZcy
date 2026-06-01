import { useState } from 'react'
import Modal from '../components/Modal'
import { rolesApi } from '../services/api'

interface CreateRoleModalProps {
  isOpen: boolean
  onClose: () => void
  onCreated: () => void
  type: 'internal' | 'external'
}

export default function CreateRoleModal({ isOpen, onClose, onCreated, type }: CreateRoleModalProps) {
  const [name, setName] = useState('')
  const [permissions, setPermissions] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const perms = permissions.split(',').map(p => p.trim()).filter(Boolean)
      if (type === 'internal') {
        await rolesApi.createInternal({ name, permissions: perms })
      } else {
        await rolesApi.createExternal({ name, permissions: perms })
      }
      setName('')
      setPermissions('')
      onCreated()
      onClose()
    } catch (err: any) {
      setError(err.message || 'Failed to create role')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={`Create ${type === 'internal' ? 'Internal' : 'External'} Role`}>
      <form onSubmit={handleSubmit}>
        {error && (
          <div style={{
            padding: '12px',
            background: 'var(--danger-bg, rgba(239,68,68,0.1))',
            color: 'var(--danger-color, #ef4444)',
            borderRadius: '6px',
            marginBottom: '16px',
            fontSize: '0.875rem',
          }}>
            {error}
          </div>
        )}

        <div style={{ marginBottom: '16px' }}>
          <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.875rem', fontWeight: 500, color: 'var(--text)' }}>
            Role Name
          </label>
          <input
            type="text"
            required
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="e.g., admin, editor, viewer"
            style={{
              width: '100%',
              padding: '10px 12px',
              borderRadius: '6px',
              border: '1px solid var(--border-color)',
              background: 'var(--input-bg, var(--card-bg))',
              color: 'var(--text)',
              fontSize: '0.875rem',
              outline: 'none',
            }}
          />
        </div>

        <div style={{ marginBottom: '20px' }}>
          <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.875rem', fontWeight: 500, color: 'var(--text)' }}>
            Permissions
          </label>
          <textarea
            required
            rows={3}
            value={permissions}
            onChange={e => setPermissions(e.target.value)}
            placeholder="e.g., templates:read, templates:write, images:read"
            style={{
              width: '100%',
              padding: '10px 12px',
              borderRadius: '6px',
              border: '1px solid var(--border-color)',
              background: 'var(--input-bg, var(--card-bg))',
              color: 'var(--text)',
              fontSize: '0.875rem',
              outline: 'none',
              resize: 'vertical',
            }}
          />
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '4px' }}>
            Separate permissions with commas
          </p>
        </div>

        <div style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end' }}>
          <button
            type="button"
            onClick={onClose}
            disabled={loading}
            style={{
              padding: '10px 20px',
              borderRadius: '6px',
              border: '1px solid var(--border-color)',
              background: 'transparent',
              color: 'var(--text)',
              cursor: 'pointer',
              fontSize: '0.875rem',
            }}
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            style={{
              padding: '10px 20px',
              borderRadius: '6px',
              border: 'none',
              background: 'var(--primary-color)',
              color: '#fff',
              cursor: 'pointer',
              fontSize: '0.875rem',
            }}
          >
            {loading ? 'Creating...' : 'Create Role'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
