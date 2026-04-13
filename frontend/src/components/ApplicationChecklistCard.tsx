import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateChecklist } from '@/api/client'
import type { Application, ApplicationChecklist } from '@/types'

interface Props {
  app: Application
}

const ITEMS: { key: keyof ApplicationChecklist; label: string; desc: string }[] = [
  { key: 'cover_letter',    label: 'Cover letter',       desc: 'Included a tailored cover letter' },
  { key: 'cv_tailored',     label: 'Tailored CV',        desc: 'CV was customised for this role' },
  { key: 'referral',        label: 'Referral',           desc: 'Had an internal referral' },
  { key: 'video_intro',     label: 'Video introduction', desc: 'Included a video introduction' },
  { key: 'linkedin_connect',label: 'LinkedIn connection',desc: 'Connected with the hiring manager' },
]

export default function ApplicationChecklistCard({ app }: Props) {
  const qc = useQueryClient()
  const [checklist, setChecklist] = useState<ApplicationChecklist>({
    cover_letter:     app.cover_letter,
    cv_tailored:      app.cv_tailored,
    referral:         app.referral,
    portfolio_link:   app.portfolio_link,
    video_intro:      app.video_intro,
    linkedin_connect: app.linkedin_connect,
  })
  const [portfolioLink, setPortfolioLink] = useState(app.portfolio_link ?? '')
  const [saved, setSaved] = useState(false)

  const mutation = useMutation({
    mutationFn: () => updateChecklist(app.id, {
      ...checklist, portfolio_link: portfolioLink || undefined,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['application', app.id] })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    },
  })

  function toggle(key: keyof ApplicationChecklist) {
    if (key === 'portfolio_link') return
    setChecklist(prev => ({ ...prev, [key]: !prev[key] }))
  }

  const checkedCount = ITEMS.filter(i => checklist[i.key]).length

  return (
    <div style={{
      background: '#fff', borderRadius: 14, padding: '20px',
      boxShadow: '0 1px 3px rgba(0,0,0,0.07)',
    }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h3 style={{ fontSize: 14, fontWeight: 800, color: '#191919' }}>
          Submission checklist
        </h3>
        <span style={{
          fontSize: 12, fontWeight: 700, padding: '3px 10px',
          borderRadius: 99, background: checkedCount > 0 ? '#ECFDF5' : '#F6F6F6',
          color: checkedCount > 0 ? '#047857' : '#767676',
        }}>
          {checkedCount}/{ITEMS.length}
        </span>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {ITEMS.map(({ key, label, desc }) => (
          <button
            key={key}
            onClick={() => toggle(key)}
            style={{
              display: 'flex', alignItems: 'center', gap: 12,
              padding: '10px 12px', borderRadius: 10,
              background: checklist[key] ? '#F0FDF4' : '#F9FAFB',
              border: `1.5px solid ${checklist[key] ? '#86EFAC' : '#E8E8E8'}`,
              textAlign: 'left', transition: 'all 0.12s', cursor: 'pointer',
            }}
          >
            {/* Checkbox */}
            <div style={{
              width: 20, height: 20, borderRadius: 6, flexShrink: 0,
              background: checklist[key] ? '#22C55E' : '#fff',
              border: `2px solid ${checklist[key] ? '#22C55E' : '#D1D5DB'}`,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              transition: 'all 0.12s',
            }}>
              {checklist[key] && (
                <svg width="11" height="9" viewBox="0 0 11 9" fill="none">
                  <path d="M1 4L4 7.5L10 1" stroke="white" strokeWidth="2"
                    strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              )}
            </div>
            <div>
              <div style={{ fontSize: 13, fontWeight: 700, color: '#191919' }}>{label}</div>
              <div style={{ fontSize: 12, color: '#767676', marginTop: 1 }}>{desc}</div>
            </div>
          </button>
        ))}

        {/* Portfolio link */}
        <div style={{
          padding: '10px 12px', borderRadius: 10,
          background: portfolioLink ? '#F0FDF4' : '#F9FAFB',
          border: `1.5px solid ${portfolioLink ? '#86EFAC' : '#E8E8E8'}`,
        }}>
          <div style={{ fontSize: 13, fontWeight: 700, color: '#191919', marginBottom: 6 }}>
            Portfolio / project link
          </div>
          <input
            value={portfolioLink}
            onChange={e => setPortfolioLink(e.target.value)}
            placeholder="https://github.com/you/project"
            style={{
              width: '100%', padding: '8px 10px', borderRadius: 8,
              border: '1.5px solid #E8E8E8', fontSize: 13, outline: 'none',
              background: '#fff',
            }}
          />
        </div>
      </div>

      <button
        onClick={() => mutation.mutate()}
        disabled={mutation.isPending}
        style={{
          marginTop: 14, width: '100%', padding: '11px 0',
          borderRadius: 10, fontWeight: 800, fontSize: 13,
          background: saved ? '#22C55E' : '#191919',
          color: '#fff', transition: 'background 0.2s',
        }}
      >
        {mutation.isPending ? 'Saving…' : saved ? '✓ Saved' : 'Save checklist'}
      </button>
    </div>
  )
}
