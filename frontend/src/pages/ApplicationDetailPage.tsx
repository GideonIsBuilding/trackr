import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, ExternalLink, MapPin, Calendar, Clock } from 'lucide-react'
import {
  getApplication, getStatusHistory, updateStatus,
  snoozeReminder, configureReminder,
} from '@/api/client'
import { StatusBadge } from '@/components/StatusBadge'
import DeleteApplicationButton from '@/components/DeleteApplicationButton'
import ApplicationChecklistCard from '@/components/ApplicationChecklistCard'
import type { ApplicationStatus } from '@/types'
import { ALL_STATUSES, STATUS_LABELS, SOURCE_LABELS, TERMINAL_STATUSES } from '@/types'
import { format, formatDistanceToNow } from 'date-fns'

export default function ApplicationDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [newStatus, setNewStatus] = useState<ApplicationStatus | ''>('')
  const [statusNote, setStatusNote] = useState('')
  const [reminderDays, setReminderDays] = useState(14)
  const [snoozeDays] = useState(7)

  const { data: app, isLoading } = useQuery({
    queryKey: ['application', id],
    queryFn: () => getApplication(id!),
  })

  const { data: history = [] } = useQuery({
    queryKey: ['history', id],
    queryFn: () => getStatusHistory(id!),
    enabled: !!id,
  })

  const statusMutation = useMutation({
    mutationFn: () => updateStatus(id!, newStatus as ApplicationStatus, statusNote || undefined),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['application', id] })
      qc.invalidateQueries({ queryKey: ['history', id] })
      qc.invalidateQueries({ queryKey: ['applications'] })
      setNewStatus('')
      setStatusNote('')
    },
  })

  const snoozeMutation = useMutation({
    mutationFn: () => snoozeReminder(id!, snoozeDays),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['application', id] }),
  })

  const reminderMutation = useMutation({
    mutationFn: () => configureReminder(id!, reminderDays),
  })

  if (isLoading || !app) {
    return (
      <div style={{
        minHeight: '100vh', background: '#F6F6F6',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
      }}>
        <div style={{ color: '#9CA3AF', fontSize: 15 }}>Loading…</div>
      </div>
    )
  }

  const isTerminal = TERMINAL_STATUSES.includes(app.status)
  const daysSinceActivity = Math.floor(
    (Date.now() - new Date(app.last_activity_at).getTime()) / 86_400_000
  )

  return (
    <div style={{ minHeight: '100vh', background: '#F6F6F6' }}>
      {/* Nav */}
      <nav style={{
        background: '#fff', borderBottom: '1px solid #EBEBEB',
        padding: '0 32px', height: 64,
        display: 'flex', alignItems: 'center', gap: 16,
        position: 'sticky', top: 0, zIndex: 100,
      }}>
        <button onClick={() => navigate('/')} style={{
          display: 'flex', alignItems: 'center', gap: 6,
          background: '#F6F6F6', borderRadius: 8, padding: '8px 12px',
          fontSize: 13, fontWeight: 700, color: '#494949',
        }}>
          <ArrowLeft size={15} /> Back
        </button>
        <span style={{ color: '#E8E8E8' }}>|</span>
        <span style={{ fontWeight: 800, fontSize: 16, color: '#191919' }}>
          {app.company} — {app.role}
        </span>
      </nav>

      <div style={{
        maxWidth: 900, margin: '0 auto', padding: '40px 32px',
        display: 'grid', gridTemplateColumns: '1fr 340px', gap: 24,
      }}>
        {/* ── Left column ── */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>

          {/* Hero card */}
          <div style={{
            background: '#fff', borderRadius: 16, padding: '28px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
          }}>
            <div style={{
              display: 'flex', alignItems: 'flex-start',
              justifyContent: 'space-between', marginBottom: 20,
            }}>
              <div style={{ display: 'flex', gap: 16, alignItems: 'center' }}>
                <div style={{
                  width: 56, height: 56, borderRadius: 14,
                  background: '#FF3008', color: '#fff',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: 24, fontWeight: 800,
                }}>
                  {app.company[0].toUpperCase()}
                </div>
                <div>
                  <h1 style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.5px', color: '#191919' }}>
                    {app.company}
                  </h1>
                  <p style={{ fontSize: 15, color: '#767676', marginTop: 2 }}>{app.role}</p>
                </div>
              </div>
              <StatusBadge status={app.status} />
            </div>

            <div style={{ display: 'flex', gap: 20, flexWrap: 'wrap' }}>
              {app.location && (
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, color: '#767676' }}>
                  <MapPin size={14} /> {app.location}
                </div>
              )}
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, color: '#767676' }}>
                <Calendar size={14} /> Applied {format(new Date(app.applied_at), 'MMM d, yyyy')}
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, color: '#767676' }}>
                <Clock size={14} /> Last activity {formatDistanceToNow(new Date(app.last_activity_at), { addSuffix: true })}
              </div>
              <div style={{ fontSize: 13, color: '#767676' }}>
                Source: {SOURCE_LABELS[app.source]}
              </div>
            </div>

            {app.job_url && (
              <a href={app.job_url} target="_blank" rel="noreferrer" style={{
                display: 'inline-flex', alignItems: 'center', gap: 6,
                marginTop: 16, fontSize: 13, fontWeight: 700, color: '#FF3008',
              }}>
                View job posting <ExternalLink size={13} />
              </a>
            )}

            {app.notes && (
              <div style={{
                marginTop: 20, padding: '14px 16px', background: '#F6F6F6',
                borderRadius: 10, fontSize: 14, color: '#494949', lineHeight: 1.6,
                borderLeft: '3px solid #FF3008',
              }}>
                {app.notes}
              </div>
            )}
          </div>

          {/* Status update */}
          {!isTerminal && (
            <div style={{
              background: '#fff', borderRadius: 16, padding: '24px 28px',
              boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
            }}>
              <h2 style={{ fontSize: 16, fontWeight: 800, marginBottom: 16, color: '#191919' }}>
                Update status
              </h2>
              <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 16 }}>
                {ALL_STATUSES.filter(s => s !== app.status).map(s => (
                  <button key={s} onClick={() => setNewStatus(s)} style={{
                    padding: '8px 14px', borderRadius: 8,
                    fontSize: 13, fontWeight: 700,
                    background: newStatus === s ? '#FF3008' : '#F6F6F6',
                    color: newStatus === s ? '#fff' : '#494949',
                    transition: 'all 0.1s',
                  }}>
                    {STATUS_LABELS[s]}
                  </button>
                ))}
              </div>
              {newStatus && (
                <>
                  <textarea
                    value={statusNote}
                    onChange={e => setStatusNote(e.target.value)}
                    placeholder="Add a note (optional) — what happened?"
                    rows={2}
                    style={{
                      width: '100%', padding: '12px 14px', borderRadius: 10,
                      border: '1.5px solid #E8E8E8', fontSize: 14, outline: 'none',
                      marginBottom: 12, resize: 'vertical', lineHeight: 1.5,
                    }}
                  />
                  <button
                    onClick={() => statusMutation.mutate()}
                    disabled={statusMutation.isPending}
                    style={{
                      padding: '11px 22px', borderRadius: 10,
                      background: '#FF3008', color: '#fff',
                      fontWeight: 800, fontSize: 14,
                      boxShadow: '0 4px 12px rgba(255,48,8,0.3)',
                    }}
                  >
                    {statusMutation.isPending ? 'Saving…' : `Move to ${STATUS_LABELS[newStatus]}`}
                  </button>
                </>
              )}
            </div>
          )}

          {/* Activity timeline */}
          <div style={{
            background: '#fff', borderRadius: 16, padding: '24px 28px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
          }}>
            <h2 style={{ fontSize: 16, fontWeight: 800, marginBottom: 20, color: '#191919' }}>
              Activity timeline
            </h2>
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {history.map((h, i) => (
                <div key={h.id} style={{ display: 'flex', gap: 16, position: 'relative' }}>
                  {i < history.length - 1 && (
                    <div style={{
                      position: 'absolute', left: 7, top: 24, bottom: -4,
                      width: 2, background: '#F0F0F0',
                    }} />
                  )}
                  <div style={{
                    width: 16, height: 16, borderRadius: '50%', flexShrink: 0, marginTop: 2,
                    background: i === history.length - 1 ? '#FF3008' : '#E8E8E8',
                    boxShadow: i === history.length - 1 ? '0 0 0 3px #FFDDD8' : 'none',
                  }} />
                  <div style={{ paddingBottom: 20 }}>
                    <div style={{ fontSize: 14, fontWeight: 700, color: '#191919' }}>
                      {h.from_status ? `${STATUS_LABELS[h.from_status]} → ` : ''}
                      {STATUS_LABELS[h.to_status]}
                    </div>
                    {h.note && (
                      <div style={{ fontSize: 13, color: '#767676', marginTop: 3 }}>{h.note}</div>
                    )}
                    <div style={{ fontSize: 12, color: '#9CA3AF', marginTop: 4 }}>
                      {format(new Date(h.changed_at), 'MMM d, yyyy · h:mm a')}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* ── Right sidebar ── */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>

          {/* Follow-up alert */}
          {daysSinceActivity >= 14 && !isTerminal && (
            <div style={{
              background: '#FFFBEB', border: '1.5px solid #FDE68A',
              borderRadius: 14, padding: '18px 20px',
            }}>
              <div style={{ fontSize: 14, fontWeight: 800, color: '#92400E', marginBottom: 4 }}>
                ⚠ Follow-up needed
              </div>
              <div style={{ fontSize: 13, color: '#B45309', lineHeight: 1.5 }}>
                No activity for {daysSinceActivity} days. Consider sending a follow-up email.
              </div>
              <button
                onClick={() => snoozeMutation.mutate()}
                disabled={snoozeMutation.isPending}
                style={{
                  marginTop: 12, width: '100%', padding: '10px 0',
                  borderRadius: 8, background: '#F59E0B', color: '#fff',
                  fontWeight: 700, fontSize: 13,
                }}
              >
                {snoozeMutation.isPending ? 'Snoozing…' : `Snooze ${snoozeDays} days`}
              </button>
            </div>
          )}

          {/* Checklist card — NEW */}
          <ApplicationChecklistCard app={app} />

          {/* Reminder settings */}
          <div style={{
            background: '#fff', borderRadius: 14, padding: '20px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
          }}>
            <h3 style={{ fontSize: 14, fontWeight: 800, marginBottom: 14, color: '#191919' }}>
              Reminder settings
            </h3>
            <label style={{ fontSize: 12, fontWeight: 700, color: '#767676', display: 'block', marginBottom: 8 }}>
              Alert after no activity for
            </label>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <input
                type="number" min={1} max={90}
                value={reminderDays}
                onChange={e => setReminderDays(Number(e.target.value))}
                style={{
                  width: 72, padding: '10px 12px', borderRadius: 8,
                  border: '1.5px solid #E8E8E8', fontSize: 14, fontWeight: 700,
                  textAlign: 'center', outline: 'none',
                }}
              />
              <span style={{ fontSize: 13, color: '#767676' }}>days</span>
              <button
                onClick={() => reminderMutation.mutate()}
                disabled={reminderMutation.isPending}
                style={{
                  flex: 1, padding: '10px 0', borderRadius: 8,
                  background: '#191919', color: '#fff', fontWeight: 700, fontSize: 13,
                }}
              >
                {reminderMutation.isPending ? 'Saving…' : 'Save'}
              </button>
            </div>
          </div>

          {/* Quick stats */}
          <div style={{
            background: '#fff', borderRadius: 14, padding: '20px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
          }}>
            <h3 style={{ fontSize: 14, fontWeight: 800, marginBottom: 14, color: '#191919' }}>
              Quick stats
            </h3>
            {[
              ['Status',           STATUS_LABELS[app.status]],
              ['Transitions',      `${history.length} total`],
              ['Days since applied', `${Math.floor((Date.now() - new Date(app.applied_at).getTime()) / 86_400_000)}d`],
              ['Last activity',    formatDistanceToNow(new Date(app.last_activity_at), { addSuffix: true })],
            ].map(([k, v]) => (
              <div key={k} style={{
                display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                padding: '10px 0', borderBottom: '1px solid #F6F6F6',
              }}>
                <span style={{ fontSize: 13, color: '#767676' }}>{k}</span>
                <span style={{ fontSize: 13, fontWeight: 700, color: '#191919' }}>{v}</span>
              </div>
            ))}
          </div>

          {/* Delete */}
          <DeleteApplicationButton applicationId={app.id} company={app.company} />
        </div>
      </div>
    </div>
  )
}
