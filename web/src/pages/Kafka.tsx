import { useEffect, useState } from 'react'
import { PageHeader } from '@kzcy/dashboard'
import { kafkaApi, type KafkaConnection } from '../services/api'
import Modal from '../components/Modal'

export default function Kafka() {
  const [connections, setConnections] = useState<KafkaConnection[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({
    name: '',
    brokers: '',
    topic: '',
    group_id: '',
  })
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    loadConnections()
  }, [])

  async function loadConnections() {
    try {
      const res = await kafkaApi.list()
      setConnections(res.data || [])
    } catch (error) {
      console.error('Failed to load kafka connections:', error)
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    try {
      await kafkaApi.create({
        name: form.name,
        brokers: form.brokers.split(',').map(b => b.trim()).filter(Boolean),
        topic: form.topic,
        group_id: form.group_id,
      })
      setShowCreate(false)
      setForm({ name: '', brokers: '', topic: '', group_id: '' })
      loadConnections()
    } catch (error: any) {
      alert('Failed to create connection: ' + (error.message || 'Unknown error'))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Are you sure you want to delete this connection?')) return
    try {
      await kafkaApi.delete(id)
      setConnections(prev => prev.filter(c => c._id !== id))
    } catch (error) {
      console.error('Failed to delete connection:', error)
    }
  }

  async function handleToggle(id: string, currentStatus: string) {
    try {
      if (currentStatus === 'running') {
        await kafkaApi.stop(id)
      } else {
        await kafkaApi.start(id)
      }
      loadConnections()
    } catch (error) {
      console.error('Failed to toggle connection:', error)
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
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <PageHeader title="Kafka Connections" description="Manage Kafka consumer connections" />
        <button
          onClick={() => setShowCreate(true)}
          style={{
            padding: '8px 20px',
            borderRadius: '6px',
            background: 'var(--primary-color)',
            color: '#fff',
            border: 'none',
            cursor: 'pointer',
            fontSize: '0.9rem',
            fontWeight: 500,
          }}
        >
          + Create Connection
        </button>
      </div>

      <div className="grid grid-cols-1 gap-4 mt-6">
        {connections.length === 0 ? (
          <div className="text-center py-12" style={{ color: 'var(--text-muted)' }}>
            No Kafka connections found.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '2px solid var(--border-color)' }}>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Name</th>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Brokers</th>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Topics</th>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Group ID</th>
                  <th style={{ padding: '12px', textAlign: 'left', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Status</th>
                  <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {connections.map((conn) => (
                  <tr key={conn._id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                    <td style={{ padding: '12px', color: 'var(--text)' }}>{conn.name}</td>
                    <td style={{ padding: '12px', color: 'var(--text-muted)', fontFamily: 'monospace', fontSize: '0.8rem' }}>
                      {conn.brokers?.join(', ')}
                    </td>
                    <td style={{ padding: '12px', color: 'var(--text-muted)' }}>
                      {conn.topic}
                    </td>
                    <td style={{ padding: '12px', color: 'var(--text-muted)', fontFamily: 'monospace', fontSize: '0.8rem' }}>
                      {conn.group_id}
                    </td>
                    <td style={{ padding: '12px' }}>
                      <span style={{
                        padding: '2px 8px',
                        borderRadius: '4px',
                        fontSize: '0.75rem',
                        background: conn.status === 'running' ? 'var(--badge-success-bg)' : 'var(--badge-muted-bg)',
                        color: conn.status === 'running' ? 'var(--badge-success-color)' : 'var(--badge-muted-color)',
                      }}>
                        {conn.status}
                      </span>
                    </td>
                    <td style={{ padding: '12px', textAlign: 'right' }}>
                      <button
                        onClick={() => handleToggle(conn._id, conn.status)}
                        style={{
                          padding: '4px 12px',
                          borderRadius: '4px',
                          border: '1px solid var(--primary-color)',
                          color: 'var(--primary-color)',
                          background: 'transparent',
                          cursor: 'pointer',
                          fontSize: '0.8rem',
                          marginRight: '8px',
                        }}
                      >
                        {conn.status === 'running' ? 'Stop' : 'Start'}
                      </button>
                      <button
                        onClick={() => handleDelete(conn._id)}
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

      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="Create Kafka Connection">
        <form onSubmit={handleCreate}>
          <div style={{ marginBottom: '16px' }}>
            <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500 }}>Name</label>
            <input
              type="text"
              value={form.name}
              onChange={e => setForm({ ...form, name: e.target.value })}
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
            />
          </div>
          <div style={{ marginBottom: '16px' }}>
            <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500 }}>
              Brokers <span style={{ color: 'var(--text-muted)', fontWeight: 400 }}>(comma-separated)</span>
            </label>
            <input
              type="text"
              value={form.brokers}
              onChange={e => setForm({ ...form, brokers: e.target.value })}
              placeholder="localhost:9092, localhost:9093"
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
            />
          </div>
          <div style={{ marginBottom: '16px' }}>
            <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500 }}>
              Topics <span style={{ color: 'var(--text-muted)', fontWeight: 400 }}>(comma-separated)</span>
            </label>
            <input
              type="text"
              value={form.topic}
              onChange={e => setForm({ ...form, topic: e.target.value })}
              placeholder="render-requests, print-jobs"
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
            />
          </div>
          <div style={{ marginBottom: '24px' }}>
            <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.85rem', fontWeight: 500 }}>Group ID</label>
            <input
              type="text"
              value={form.group_id}
              onChange={e => setForm({ ...form, group_id: e.target.value })}
              placeholder="imposizcy-consumer"
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
            />
          </div>
          <div style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end' }}>
            <button
              type="button"
              onClick={() => setShowCreate(false)}
              style={{
                padding: '10px 20px',
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
              disabled={submitting}
              style={{
                padding: '10px 20px',
                borderRadius: '6px',
                border: 'none',
                background: 'var(--primary-color)',
                color: '#fff',
                cursor: submitting ? 'not-allowed' : 'pointer',
                opacity: submitting ? 0.7 : 1,
              }}
            >
              {submitting ? 'Creating...' : 'Create Connection'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
