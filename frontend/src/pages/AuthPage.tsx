import { useState, useMemo } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { login, register } from '@/api/client'
import { useAuth } from '@/hooks/useAuth'

// ── Password strength ─────────────────────────────────────────────────────────

interface PasswordChecks {
  length: boolean
  upper: boolean
  lower: boolean
  digit: boolean
  special: boolean
}

function checkPassword(p: string): PasswordChecks {
  return {
    length:  p.length >= 12,
    upper:   /[A-Z]/.test(p),
    lower:   /[a-z]/.test(p),
    digit:   /[0-9]/.test(p),
    special: /[^A-Za-z0-9]/.test(p),
  }
}

type StrengthLevel = 0 | 1 | 2 | 3 | 4

function strengthLevel(checks: PasswordChecks): StrengthLevel {
  const score = Object.values(checks).filter(Boolean).length
  if (score <= 1) return 0
  if (score === 2) return 1
  if (score === 3) return 2
  if (score === 4) return 3
  return 4
}

const STRENGTH_LABEL = ['Very weak', 'Weak', 'Fair', 'Good', 'Strong'] as const
const STRENGTH_COLOR = ['#EF4444', '#F97316', '#EAB308', '#22C55E', '#16A34A'] as const

function PasswordStrength({ password }: { password: string }) {
  const checks = useMemo(() => checkPassword(password), [password])
  const level  = strengthLevel(checks)
  const color  = STRENGTH_COLOR[level]

  if (!password) return null

  const requirements: [keyof PasswordChecks, string][] = [
    ['length',  'At least 12 characters'],
    ['upper',   'Uppercase letter (A–Z)'],
    ['lower',   'Lowercase letter (a–z)'],
    ['digit',   'Number (0–9)'],
    ['special', 'Special character (!@#$…)'],
  ]

  return (
    <div style={{ marginTop: 12 }}>
      {/* Label */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
        <span style={{ fontSize: 12, fontWeight: 600, color: '#767676' }}>Password strength</span>
        <span style={{ fontSize: 12, fontWeight: 700, color }}>{STRENGTH_LABEL[level]}</span>
      </div>
      {/* Bar */}
      <div style={{ display: 'flex', gap: 4, marginBottom: 12 }}>
        {([0, 1, 2, 3] as const).map(i => (
          <div key={i} style={{
            flex: 1, height: 8, borderRadius: 99,
            background: i < level ? color : '#E8E8E8',
            transition: 'background 0.25s',
          }} />
        ))}
      </div>

      {/* Checklist */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        {requirements.map(([key, label]) => (
          <div key={key} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <span style={{
              width: 16, height: 16, borderRadius: '50%', flexShrink: 0,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              background: checks[key] ? '#22C55E' : '#E8E8E8',
              transition: 'background 0.15s',
            }}>
              {checks[key] && (
                <svg width="9" height="7" viewBox="0 0 9 7" fill="none">
                  <path d="M1 3.5L3.5 6L8 1" stroke="#fff" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              )}
            </span>
            <span style={{
              fontSize: 12, fontWeight: 500,
              color: checks[key] ? '#191919' : '#9CA3AF',
              transition: 'color 0.15s',
            }}>
              {label}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function AuthPage() {
  const [mode, setMode]       = useState<'login' | 'register'>('login')
  const [email, setEmail]     = useState('')
  const [password, setPassword] = useState('')
  const [error, setError]     = useState('')
  const [loading, setLoading] = useState(false)
  const { signIn } = useAuth()
  const navigate   = useNavigate()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const fn = mode === 'login' ? login : register
      const { user } = await fn(email, password, Intl.DateTimeFormat().resolvedOptions().timeZone)
      signIn(user)
      navigate('/')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      minHeight: '100vh', background: '#fff',
      display: 'grid', gridTemplateColumns: '1fr 1fr',
    }}>
      {/* Left panel */}
      <div style={{
        background: '#FF3008', display: 'flex', flexDirection: 'column',
        justifyContent: 'center', padding: '80px 72px', color: '#fff',
      }}>
        <div style={{ fontSize: 28, fontWeight: 800, letterSpacing: '-0.5px', marginBottom: 48 }}>
          trackr.
        </div>
        <h1 style={{ fontSize: 52, fontWeight: 800, lineHeight: 1.1, letterSpacing: '-2px', marginBottom: 24 }}>
          Never let an<br />opportunity<br />slip away.
        </h1>
        <p style={{ fontSize: 18, fontWeight: 500, opacity: 0.85, lineHeight: 1.6 }}>
          Track every application, follow up at the right time, and land your next role faster.
        </p>
        <div style={{ display: 'flex', gap: 40, marginTop: 64 }}>
          {[['14', 'day follow-up alerts'], ['100%', 'your data'], ['0', 'missed callbacks']].map(([num, label]) => (
            <div key={num}>
              <div style={{ fontSize: 32, fontWeight: 800, letterSpacing: '-1px' }}>{num}</div>
              <div style={{ fontSize: 13, fontWeight: 500, opacity: 0.75, marginTop: 2 }}>{label}</div>
            </div>
          ))}
        </div>
      </div>

      {/* Right panel */}
      <div style={{
        display: 'flex', flexDirection: 'column',
        justifyContent: 'center', alignItems: 'center',
        padding: '80px 72px', background: '#fff',
      }}>
        <div style={{ width: '100%', maxWidth: 400 }}>
          {/* Tab switcher */}
          <div style={{
            display: 'flex', background: '#F6F6F6',
            borderRadius: 12, padding: 4, marginBottom: 40,
          }}>
            {(['login', 'register'] as const).map(m => (
              <button key={m} onClick={() => { setMode(m); setError(''); setPassword('') }} style={{
                flex: 1, padding: '10px 0', borderRadius: 9,
                fontWeight: 700, fontSize: 14,
                background: mode === m ? '#fff' : 'transparent',
                color: mode === m ? '#191919' : '#767676',
                boxShadow: mode === m ? '0 1px 3px rgba(0,0,0,0.1)' : 'none',
                transition: 'all 0.15s',
              }}>
                {m === 'login' ? 'Sign in' : 'Create account'}
              </button>
            ))}
          </div>

          <h2 style={{ fontSize: 28, fontWeight: 800, letterSpacing: '-0.5px', marginBottom: 8 }}>
            {mode === 'login' ? 'Welcome back' : 'Get started'}
          </h2>
          <p style={{ color: '#767676', fontSize: 15, marginBottom: 32 }}>
            {mode === 'login' ? 'Sign in to your account to continue.' : 'Create a free account to start tracking.'}
          </p>

          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <label style={{ display: 'block', fontSize: 13, fontWeight: 700, marginBottom: 6, color: '#191919' }}>
                Email address
              </label>
              <input type="email" required value={email} onChange={e => setEmail(e.target.value)}
                placeholder="you@example.com"
                style={{ width: '100%', padding: '14px 16px', borderRadius: 10, border: '1.5px solid #E8E8E8', fontSize: 15, outline: 'none' }}
                onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
                onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'}
              />
            </div>

            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                <label style={{ fontSize: 13, fontWeight: 700, color: '#191919' }}>Password</label>
                {mode === 'login' && (
                  <Link to="/forgot-password" style={{ fontSize: 13, fontWeight: 700, color: '#FF3008' }}>
                    Forgot password?
                  </Link>
                )}
              </div>
              <input type="password" required value={password} onChange={e => setPassword(e.target.value)}
                placeholder={mode === 'register' ? 'Min. 12 characters' : '••••••••'}
                style={{ width: '100%', padding: '14px 16px', borderRadius: 10, border: '1.5px solid #E8E8E8', fontSize: 15, outline: 'none' }}
                onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
                onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'}
              />
              {mode === 'register' && <PasswordStrength password={password} />}
            </div>

            {error && (
              <div style={{ background: '#FEF2F2', color: '#B91C1C', padding: '12px 16px', borderRadius: 10, fontSize: 14, fontWeight: 500 }}>
                {error}
              </div>
            )}

            <button type="submit" disabled={loading} style={{
              marginTop: 8, padding: '15px 0', borderRadius: 12,
              background: loading ? '#ccc' : '#FF3008', color: '#fff',
              fontWeight: 800, fontSize: 16, letterSpacing: '-0.2px',
              boxShadow: loading ? 'none' : '0 4px 16px rgba(255,48,8,0.35)',
            }}>
              {loading ? 'Loading…' : mode === 'login' ? 'Sign in' : 'Create account'}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
