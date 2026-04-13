import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, TrendingUp, Target, Award, Clock } from 'lucide-react'
import { getAnalytics } from '@/api/client'
import type { ChecklistCorrelation, FunnelStage, SourceStat } from '@/types'

export default function AnalyticsPage() {
  const navigate = useNavigate()
  const { data, isLoading, error } = useQuery({
  queryKey: ['analytics'],
  queryFn: getAnalytics,
})

if (isLoading) {
  return (
    <div style={{
      minHeight: '100vh', background: '#F6F6F6',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
    }}>
      <div style={{ color: '#9CA3AF', fontSize: 15 }}>Computing your stats…</div>
    </div>
  )
}

if (error || !data) {
  return (
    <div style={{
      minHeight: '100vh', background: '#F6F6F6',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      flexDirection: 'column', gap: 12,
    }}>
      <div style={{ fontSize: 32 }}>⚠️</div>
      <div style={{ fontSize: 16, fontWeight: 800, color: '#191919' }}>
        Could not load analytics
      </div>
      <div style={{ fontSize: 13, color: '#767676' }}>
        {error instanceof Error ? error.message : 'Check your backend terminal for errors'}
      </div>
    </div>
  )
}

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
        <span style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.5px', color: '#FF3008' }}>
          trackr.
        </span>
        <span style={{ fontWeight: 800, fontSize: 16, color: '#191919', marginLeft: 4 }}>
          Analytics
        </span>
      </nav>

      <div style={{ maxWidth: 1000, margin: '0 auto', padding: '40px 32px', display: 'flex', flexDirection: 'column', gap: 24 }}>

        {/* Header */}
        <div>
          <h1 style={{ fontSize: 32, fontWeight: 800, letterSpacing: '-1px', color: '#191919' }}>
            Your winning formula
          </h1>
          <p style={{ color: '#767676', fontSize: 15, marginTop: 6 }}>
            Based on {data.total_applications} application{data.total_applications !== 1 ? 's' : ''}
          </p>
        </div>

        {/* Top stat cards */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16 }}>
          {[
            { icon: <TrendingUp size={18}/>, label: 'Response rate',     value: `${data.response_rate}%`,      bg: '#EFF6FF', color: '#1D4ED8' },
            { icon: <Target size={18}/>,    label: 'Interview rate',     value: `${data.interview_rate}%`,     bg: '#F5F3FF', color: '#6D28D9' },
            { icon: <Award size={18}/>,     label: 'Offer rate',         value: `${data.offer_rate}%`,         bg: '#ECFDF5', color: '#047857' },
            { icon: <Clock size={18}/>,     label: 'Avg days to reply',  value: `${data.avg_days_to_response}d`, bg: '#FFFBEB', color: '#B45309' },
          ].map(({ icon, label, value, bg, color }) => (
            <div key={label} style={{
              background: '#fff', borderRadius: 16, padding: '20px',
              boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
            }}>
              <div style={{
                display: 'inline-flex', background: bg, color,
                borderRadius: 10, padding: 10, marginBottom: 12,
              }}>
                {icon}
              </div>
              <div style={{ fontSize: 28, fontWeight: 800, letterSpacing: '-1px', color: '#191919' }}>
                {value}
              </div>
              <div style={{ fontSize: 12, fontWeight: 600, color: '#767676', marginTop: 2 }}>{label}</div>
            </div>
          ))}
        </div>

        {/* Conversion funnel */}
        <div style={{
          background: '#fff', borderRadius: 16, padding: '28px',
          boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
        }}>
          <h2 style={{ fontSize: 18, fontWeight: 800, color: '#191919', marginBottom: 24, letterSpacing: '-0.3px' }}>
            Conversion funnel
          </h2>
          <FunnelChart stages={data.funnel} />
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
          {/* By source */}
          <div style={{
            background: '#fff', borderRadius: 16, padding: '28px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
          }}>
            <h2 style={{ fontSize: 18, fontWeight: 800, color: '#191919', marginBottom: 6, letterSpacing: '-0.3px' }}>
              Response rate by source
            </h2>
            <p style={{ fontSize: 13, color: '#767676', marginBottom: 20 }}>
              Where your best opportunities come from
            </p>
            <SourceChart sources={data.by_source} />
          </div>

          {/* Checklist correlation */}
          <div style={{
            background: '#fff', borderRadius: 16, padding: '28px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
          }}>
            <h2 style={{ fontSize: 18, fontWeight: 800, color: '#191919', marginBottom: 6, letterSpacing: '-0.3px' }}>
              What gets you responses
            </h2>
            <p style={{ fontSize: 13, color: '#767676', marginBottom: 20 }}>
              Response rate with vs without each extra
            </p>
            {data.total_applications < 5 ? (
              <div style={{
                padding: '32px 20px', textAlign: 'center',
                background: '#F9FAFB', borderRadius: 12,
              }}>
                <div style={{ fontSize: 32, marginBottom: 12 }}>📊</div>
                <div style={{ fontSize: 14, fontWeight: 700, color: '#374151' }}>
                  Not enough data yet
                </div>
                <div style={{ fontSize: 13, color: '#6B7280', marginTop: 6 }}>
                  Log at least 5 applications to see correlation insights.
                </div>
              </div>
            ) : (
              <ChecklistCorrelationChart items={data.checklist} />
            )}
          </div>
        </div>

        {/* Insight callout */}
        {data.total_applications >= 5 && (
          <InsightCallout data={data.checklist} responseRate={data.response_rate} />
        )}
      </div>
    </div>
  )
}

// --- Funnel chart ---
function FunnelChart({ stages }: { stages: FunnelStage[] }) {
  const max = stages[0]?.count || 1
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {stages.filter(s => s.count > 0 || s.status === 'applied').map((stage, i) => (
        <div key={stage.status} style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div style={{ width: 120, fontSize: 13, fontWeight: 600, color: '#494949', textAlign: 'right', flexShrink: 0 }}>
            {stage.label}
          </div>
          <div style={{ flex: 1, position: 'relative', height: 36 }}>
            <div style={{
              position: 'absolute', left: 0, top: 0, bottom: 0,
              width: `${(stage.count / max) * 100}%`,
              background: i === 0 ? '#FF3008' : i === stages.length - 1 ? '#22C55E' : '#FFB3A7',
              borderRadius: 8,
              minWidth: stage.count > 0 ? 8 : 0,
              transition: 'width 0.5s ease',
            }} />
          </div>
          <div style={{ width: 80, display: 'flex', alignItems: 'center', gap: 8, flexShrink: 0 }}>
            <span style={{ fontSize: 16, fontWeight: 800, color: '#191919' }}>{stage.count}</span>
            {i > 0 && (
              <span style={{ fontSize: 12, color: '#9CA3AF' }}>
                {stage.rate.toFixed(0)}%
              </span>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

// --- Source chart ---
function SourceChart({ sources }: { sources: SourceStat[] }) {
  if (!sources.length) {
    return <div style={{ color: '#9CA3AF', fontSize: 14, textAlign: 'center', padding: 32 }}>No data yet</div>
  }
  const max = Math.max(...sources.map(s => s.response_rate), 1)
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
      {sources.map(s => (
        <div key={s.source}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}>
            <span style={{ fontSize: 13, fontWeight: 700, color: '#191919' }}>{s.label}</span>
            <span style={{ fontSize: 13, color: '#767676' }}>
              {s.responded}/{s.total} · <strong style={{ color: '#191919' }}>{s.response_rate}%</strong>
            </span>
          </div>
          <div style={{ height: 8, background: '#F3F4F6', borderRadius: 99 }}>
            <div style={{
              height: '100%',
              width: `${(s.response_rate / max) * 100}%`,
              background: s.response_rate >= 30 ? '#22C55E' : s.response_rate >= 15 ? '#F59E0B' : '#FF3008',
              borderRadius: 99, transition: 'width 0.5s ease',
            }} />
          </div>
        </div>
      ))}
    </div>
  )
}

// --- Checklist correlation chart ---
function ChecklistCorrelationChart({ items }: { items: ChecklistCorrelation[] }) {
  const sorted = [...items].sort((a, b) => b.lift - a.lift)
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
      {sorted.map(item => (
        <div key={item.field}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}>
            <span style={{ fontSize: 13, fontWeight: 700, color: '#191919' }}>{item.label}</span>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              {item.lift > 0 && (
                <span style={{
                  fontSize: 11, fontWeight: 700, padding: '2px 7px',
                  borderRadius: 99, background: '#ECFDF5', color: '#047857',
                }}>
                  +{item.lift.toFixed(0)}%
                </span>
              )}
              <span style={{ fontSize: 12, color: '#9CA3AF' }}>
                {item.sample_size} apps
              </span>
            </div>
          </div>
          {/* With vs without bars */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            {[
              { label: 'With', value: item.with_item, color: '#22C55E' },
              { label: 'Without', value: item.without_item, color: '#E5E7EB' },
            ].map(bar => (
              <div key={bar.label} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <span style={{ fontSize: 11, color: '#9CA3AF', width: 44, flexShrink: 0 }}>{bar.label}</span>
                <div style={{ flex: 1, height: 8, background: '#F3F4F6', borderRadius: 99 }}>
                  <div style={{
                    height: '100%', width: `${bar.value}%`,
                    background: bar.color, borderRadius: 99,
                    transition: 'width 0.5s ease',
                  }} />
                </div>
                <span style={{ fontSize: 12, fontWeight: 700, color: '#374151', width: 32, textAlign: 'right' }}>
                  {bar.value}%
                </span>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

// --- Insight callout ---
function InsightCallout({ data, responseRate }: { data: ChecklistCorrelation[]; responseRate: number }) {
  const topItem = [...data].sort((a, b) => b.lift - a.lift)[0]
  if (!topItem || topItem.lift <= 0) return null

  return (
    <div style={{
      background: 'linear-gradient(135deg, #FF3008 0%, #FF6B47 100%)',
      borderRadius: 16, padding: '24px 28px', color: '#fff',
    }}>
      <div style={{ fontSize: 12, fontWeight: 800, opacity: 0.8, letterSpacing: '0.08em', marginBottom: 8 }}>
        YOUR TOP INSIGHT
      </div>
      <div style={{ fontSize: 20, fontWeight: 800, lineHeight: 1.4, letterSpacing: '-0.3px' }}>
        Applications with a <span style={{ textDecoration: 'underline' }}>{topItem.label.toLowerCase()}</span> get{' '}
        {topItem.lift.toFixed(0)}% more responses than those without.
      </div>
      <div style={{ fontSize: 14, opacity: 0.85, marginTop: 10 }}>
        Your overall response rate is {responseRate}%. With {topItem.label.toLowerCase()}, it's {topItem.with_item}%.
      </div>
    </div>
  )
}
