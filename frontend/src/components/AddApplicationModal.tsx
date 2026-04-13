import { useState } from 'react'
import { X } from 'lucide-react'
import type { ApplicationSource } from '@/types'
import { SOURCE_LABELS } from '@/types'
import { createApplication } from '@/api/client'

interface Props {
  onClose: () => void
  onCreated: () => void
}

export default function AddApplicationModal({ onClose, onCreated }: Props) {
  const [company, setCompany] = useState('')
  const [role, setRole] = useState('')
  const [jobUrl, setJobUrl] = useState('')
  const [location, setLocation] = useState('')
  const [source, setSource] = useState<ApplicationSource>('other')
  const [notes, setNotes] = useState('')
  const [appliedAt, setAppliedAt] = useState(new Date().toISOString().slice(0, 10))
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await createApplication({
        company, role,
        job_url: jobUrl || undefined,
        location: location || undefined,
        source,
        notes: notes || undefined,
        applied_at: appliedAt,
      })
      onCreated()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to create application')
    } finally {
      setLoading(false)
    }
  }

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '12px 14px',
    borderRadius: 10,
    border: '1.5px solid #E8E8E8',
    fontSize: 14,
    outline: 'none',
    background: '#fff',
    color: '#191919',
  }

  const labelStyle: React.CSSProperties = {
    display: 'block',
    fontSize: 12,
    fontWeight: 700,
    marginBottom: 6,
    color: '#494949',
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
  }

  return (
    <div style={{
      position: 'fixed', inset: 0, zIndex: 1000,
      background: 'rgba(25,25,25,0.55)',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      padding: 24,
    }} onClick={e => { if (e.target === e.currentTarget) onClose() }}>
      <div style={{
        background: '#fff',
        borderRadius: 20,
        width: '100%',
        maxWidth: 560,
        maxHeight: '90vh',
        overflowY: 'auto',
        boxShadow: '0 24px 80px rgba(0,0,0,0.2)',
      }}>
        {/* Header */}
        <div style={{
          padding: '24px 28px 20px',
          borderBottom: '1px solid #F0F0F0',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          <div>
            <h2 style={{ fontSize: 20, fontWeight: 800, letterSpacing: '-0.3px' }}>Log application</h2>
            <p style={{ fontSize: 13, color: '#767676', marginTop: 2 }}>Add a new job application to your tracker</p>
          </div>
          <button onClick={onClose} style={{
            background: '#F6F6F6', border: 'none', borderRadius: 8,
            padding: 8, display: 'flex', color: '#494949',
          }}>
            <X size={18} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} style={{ padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 18 }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <div>
              <label style={labelStyle}>Company *</label>
              <input required value={company} onChange={e => setCompany(e.target.value)}
                placeholder="Google" style={inputStyle}
                onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
                onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'} />
            </div>
            <div>
              <label style={labelStyle}>Role *</label>
              <input required value={role} onChange={e => setRole(e.target.value)}
                placeholder="Software Engineer" style={inputStyle}
                onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
                onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'} />
            </div>
          </div>

          <div>
            <label style={labelStyle}>Job URL</label>
            <input type="url" value={jobUrl} onChange={e => setJobUrl(e.target.value)}
              placeholder="https://careers.google.com/..." style={inputStyle}
              onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
              onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'} />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <div>
              <label style={labelStyle}>Location</label>
              <input value={location} onChange={e => setLocation(e.target.value)}
                placeholder="Lagos, NG / Remote" style={inputStyle}
                onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
                onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'} />
            </div>
            <div>
              <label style={labelStyle}>Date applied</label>
              <input type="date" value={appliedAt} onChange={e => setAppliedAt(e.target.value)}
                style={inputStyle}
                onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
                onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'} />
            </div>
          </div>

          <div>
            <label style={labelStyle}>Source</label>
            <select value={source} onChange={e => setSource(e.target.value as ApplicationSource)}
              style={{ ...inputStyle, appearance: 'none' }}>
              {(Object.entries(SOURCE_LABELS) as [ApplicationSource, string][]).map(([v, l]) => (
                <option key={v} value={v}>{l}</option>
              ))}
            </select>
          </div>

          <div>
            <label style={labelStyle}>Notes</label>
            <textarea value={notes} onChange={e => setNotes(e.target.value)}
              placeholder="Referral from Jane, salary range $120k–$150k…"
              rows={3}
              style={{ ...inputStyle, resize: 'vertical', lineHeight: 1.6 }}
              onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
              onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'} />
          </div>

          {error && (
            <div style={{ background: '#FEF2F2', color: '#B91C1C', padding: '12px 16px', borderRadius: 10, fontSize: 14 }}>
              {error}
            </div>
          )}

          <div style={{ display: 'flex', gap: 12, paddingTop: 4 }}>
            <button type="button" onClick={onClose} style={{
              flex: 1, padding: '13px 0', borderRadius: 12,
              background: '#F6F6F6', color: '#494949', fontWeight: 700, fontSize: 15,
            }}>
              Cancel
            </button>
            <button type="submit" disabled={loading} style={{
              flex: 2, padding: '13px 0', borderRadius: 12,
              background: loading ? '#ccc' : '#FF3008',
              color: '#fff', fontWeight: 800, fontSize: 15,
              boxShadow: loading ? 'none' : '0 4px 16px rgba(255,48,8,0.3)',
            }}>
              {loading ? 'Saving…' : 'Log application'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
