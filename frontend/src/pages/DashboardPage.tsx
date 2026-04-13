import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate, Link } from 'react-router-dom'
import { Plus, LogOut, Briefcase, Clock, TrendingUp, Search, BarChart2 } from 'lucide-react'
import { getApplications } from '@/api/client'
import { useAuth } from '@/hooks/useAuth'
import { StatusBadge } from '@/components/StatusBadge'
import AddApplicationModal from '@/components/AddApplicationModal'
import type { ApplicationStatus } from '@/types'
import { ALL_STATUSES, STATUS_LABELS } from '@/types'
import { formatDistanceToNow } from 'date-fns'

const FILTER_TABS: { label: string; value: ApplicationStatus | 'all' }[] = [
  { label: 'All', value: 'all' },
  { label: 'Active', value: 'applied' },
  { label: 'Interviewing', value: 'interview' },
  { label: 'Offers', value: 'offer' },
  { label: 'Rejected', value: 'rejected' },
]

export default function DashboardPage() {
  const { user, signOut } = useAuth()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [activeFilter, setActiveFilter] = useState<ApplicationStatus | 'all'>('all')
  const [search, setSearch] = useState('')
  const [showAdd, setShowAdd] = useState(false)

  const { data: apps = [], isLoading } = useQuery({
    queryKey: ['applications'],
    queryFn: () => getApplications(),
  })

  const active = apps.filter(a => !['accepted','rejected','withdrawn','ghosted'].includes(a.status)).length
  const offers  = apps.filter(a => ['offer','negotiating','accepted'].includes(a.status)).length
  const stale   = apps.filter(a => {
    const days = (Date.now() - new Date(a.last_activity_at).getTime()) / 86_400_000
    return days >= 14 && !['accepted','rejected','withdrawn'].includes(a.status)
  }).length

  const filtered = apps
    .filter(a => activeFilter === 'all' || a.status === activeFilter)
    .filter(a => {
      if (!search) return true
      const q = search.toLowerCase()
      return a.company.toLowerCase().includes(q) || a.role.toLowerCase().includes(q)
    })

  return (
    <div style={{ minHeight: '100vh', background: '#F6F6F6' }}>
      {/* Nav */}
      <nav style={{
        background: '#fff', borderBottom: '1px solid #EBEBEB',
        padding: '0 32px', display: 'flex', alignItems: 'center',
        justifyContent: 'space-between', height: 64,
        position: 'sticky', top: 0, zIndex: 100,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 32 }}>
          <div style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.5px', color: '#FF3008' }}>
            trackr.
          </div>
          <Link to="/analytics" style={{
            display: 'flex', alignItems: 'center', gap: 6,
            fontSize: 14, fontWeight: 700, color: '#494949',
            padding: '6px 12px', borderRadius: 8,
            background: '#F6F6F6',
          }}>
            <BarChart2 size={15} /> Analytics
          </Link>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <span style={{ fontSize: 14, fontWeight: 600, color: '#494949' }}>{user?.email}</span>
          <button onClick={() => { signOut(); navigate('/login') }} style={{
            display: 'flex', alignItems: 'center', gap: 6,
            background: '#F6F6F6', borderRadius: 8, padding: '8px 12px',
            fontSize: 13, fontWeight: 700, color: '#494949',
          }}>
            <LogOut size={15} /> Sign out
          </button>
        </div>
      </nav>

      <div style={{ maxWidth: 1100, margin: '0 auto', padding: '40px 32px' }}>
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 32 }}>
          <div>
            <h1 style={{ fontSize: 36, fontWeight: 800, letterSpacing: '-1.5px', color: '#191919' }}>
              My applications
            </h1>
            <p style={{ color: '#767676', fontSize: 15, marginTop: 4 }}>
              {apps.length} total · last updated just now
            </p>
          </div>
          <button onClick={() => setShowAdd(true)} style={{
            display: 'flex', alignItems: 'center', gap: 8,
            background: '#FF3008', color: '#fff',
            padding: '13px 22px', borderRadius: 12,
            fontWeight: 800, fontSize: 15,
            boxShadow: '0 4px 16px rgba(255,48,8,0.35)',
          }}>
            <Plus size={18} /> Log application
          </button>
        </div>

        {/* Stat cards */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16, marginBottom: 32 }}>
          {[
            { icon: <Briefcase size={20}/>, label: 'Active applications', value: active, color: '#2563EB', bg: '#EFF6FF' },
            { icon: <TrendingUp size={20}/>, label: 'Offers in play',      value: offers,  color: '#059669', bg: '#ECFDF5' },
            { icon: <Clock size={20}/>,     label: 'Need follow-up',       value: stale,   color: '#D97706', bg: '#FFFBEB' },
          ].map(({ icon, label, value, color, bg }) => (
            <div key={label} style={{
              background: '#fff', borderRadius: 16, padding: '20px 24px',
              boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
              display: 'flex', alignItems: 'center', gap: 16,
            }}>
              <div style={{ background: bg, color, borderRadius: 12, padding: 12, display: 'flex' }}>{icon}</div>
              <div>
                <div style={{ fontSize: 28, fontWeight: 800, letterSpacing: '-1px', color: '#191919' }}>{value}</div>
                <div style={{ fontSize: 13, fontWeight: 600, color: '#767676', marginTop: 1 }}>{label}</div>
              </div>
            </div>
          ))}
        </div>

        {/* Search + filters */}
        <div style={{ display: 'flex', gap: 16, marginBottom: 24, alignItems: 'center' }}>
          <div style={{ position: 'relative', flex: 1 }}>
            <Search size={16} style={{
              position: 'absolute', left: 14, top: '50%',
              transform: 'translateY(-50%)', color: '#9CA3AF',
            }}/>
            <input
              value={search} onChange={e => setSearch(e.target.value)}
              placeholder="Search company or role…"
              style={{
                width: '100%', padding: '11px 14px 11px 40px',
                borderRadius: 10, border: '1.5px solid #E8E8E8',
                fontSize: 14, outline: 'none', background: '#fff',
              }}
            />
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            {FILTER_TABS.map(tab => (
              <button key={tab.value} onClick={() => setActiveFilter(tab.value)} style={{
                padding: '10px 16px', borderRadius: 10,
                fontWeight: 700, fontSize: 13,
                background: activeFilter === tab.value ? '#FF3008' : '#fff',
                color: activeFilter === tab.value ? '#fff' : '#494949',
                border: activeFilter === tab.value ? 'none' : '1.5px solid #E8E8E8',
                transition: 'all 0.12s',
              }}>
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        {/* Applications list */}
        {isLoading ? (
          <div style={{ textAlign: 'center', padding: 80, color: '#9CA3AF', fontSize: 15 }}>Loading…</div>
        ) : filtered.length === 0 ? (
          <div style={{
            textAlign: 'center', padding: 80,
            background: '#fff', borderRadius: 16,
            boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
          }}>
            <div style={{ fontSize: 48, marginBottom: 16 }}>📋</div>
            <div style={{ fontSize: 18, fontWeight: 800, color: '#191919' }}>No applications yet</div>
            <div style={{ color: '#767676', marginTop: 8, fontSize: 14 }}>Log your first application to get started.</div>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {filtered.map(app => {
              const daysSince = Math.floor((Date.now() - new Date(app.last_activity_at).getTime()) / 86_400_000)
              const isStale = daysSince >= 14 && !['accepted','rejected','withdrawn'].includes(app.status)
              // Count how many checklist items are ticked
              const checklistCount = [
                app.cover_letter, app.cv_tailored, app.referral,
                app.video_intro, app.linkedin_connect, !!app.portfolio_link,
              ].filter(Boolean).length

              return (
                <div
                  key={app.id}
                  onClick={() => navigate(`/applications/${app.id}`)}
                  style={{
                    background: '#fff', borderRadius: 14, padding: '18px 24px',
                    boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
                    display: 'flex', alignItems: 'center', gap: 20, cursor: 'pointer',
                    border: isStale ? '1.5px solid #FDE68A' : '1.5px solid transparent',
                    transition: 'all 0.12s',
                  }}
                  onMouseEnter={e => (e.currentTarget as HTMLElement).style.boxShadow = '0 4px 16px rgba(0,0,0,0.1)'}
                  onMouseLeave={e => (e.currentTarget as HTMLElement).style.boxShadow = '0 1px 3px rgba(0,0,0,0.07)'}
                >
                  {/* Avatar */}
                  <div style={{
                    width: 44, height: 44, borderRadius: 10,
                    background: '#FF3008', color: '#fff',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: 18, fontWeight: 800, flexShrink: 0,
                  }}>
                    {app.company[0].toUpperCase()}
                  </div>

                  {/* Company + role */}
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 800, fontSize: 15, color: '#191919' }}>{app.company}</div>
                    <div style={{ color: '#767676', fontSize: 13, marginTop: 2 }}>{app.role}</div>
                  </div>

                  {/* Checklist pill */}
                  {checklistCount > 0 && (
                    <div style={{
                      fontSize: 12, fontWeight: 700, padding: '3px 8px',
                      borderRadius: 99, background: '#F0FDF4', color: '#15803D',
                      whiteSpace: 'nowrap',
                    }}>
                      ✓ {checklistCount} extras
                    </div>
                  )}

                  {app.location && (
                    <div style={{ fontSize: 13, color: '#9CA3AF', minWidth: 0 }}>{app.location}</div>
                  )}

                  <StatusBadge status={app.status} />

                  {isStale && (
                    <div style={{
                      background: '#FFFBEB', color: '#B45309',
                      padding: '4px 10px', borderRadius: 8,
                      fontSize: 12, fontWeight: 700, whiteSpace: 'nowrap',
                    }}>
                      ⚠ {daysSince}d no update
                    </div>
                  )}

                  <div style={{ fontSize: 12, color: '#9CA3AF', whiteSpace: 'nowrap' }}>
                    {formatDistanceToNow(new Date(app.applied_at), { addSuffix: true })}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>

      {showAdd && (
        <AddApplicationModal
          onClose={() => setShowAdd(false)}
          onCreated={() => {
            setShowAdd(false)
            qc.invalidateQueries({ queryKey: ['applications'] })
          }}
        />
      )}
    </div>
  )
}
