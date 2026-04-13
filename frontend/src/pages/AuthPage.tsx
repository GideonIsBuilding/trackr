import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { login, register } from '@/api/client'
import { useAuth } from '@/hooks/useAuth'

export default function AuthPage() {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { signIn } = useAuth()
  const navigate = useNavigate()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const fn = mode === 'login' ? login : register
      const { user, token } = await fn(email, password, Intl.DateTimeFormat().resolvedOptions().timeZone)
      signIn(user, token)
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
              <button key={m} onClick={() => { setMode(m); setError('') }} style={{
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
                placeholder={mode === 'register' ? 'Min. 8 characters' : '••••••••'}
                style={{ width: '100%', padding: '14px 16px', borderRadius: 10, border: '1.5px solid #E8E8E8', fontSize: 15, outline: 'none' }}
                onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
                onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'}
              />
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
