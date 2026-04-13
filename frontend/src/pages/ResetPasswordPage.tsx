import { useState } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { resetPassword } from '@/api/client'

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') ?? ''
  const navigate = useNavigate()

  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    if (password !== confirm) {
      setError('Passwords do not match')
      return
    }
    if (password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }
    if (!token) {
      setError('Invalid reset link. Please request a new one.')
      return
    }

    setLoading(true)
    try {
      await resetPassword(token, password)
      setSuccess(true)
      setTimeout(() => navigate('/login'), 3000)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to reset password')
    } finally {
      setLoading(false)
    }
  }

  if (!token) {
    return (
      <div style={{
        minHeight: '100vh', display: 'flex',
        alignItems: 'center', justifyContent: 'center',
        flexDirection: 'column', gap: 16, padding: 24,
      }}>
        <div style={{ fontSize: 48 }}>🔗</div>
        <h2 style={{ fontSize: 20, fontWeight: 800 }}>Invalid reset link</h2>
        <p style={{ color: '#767676' }}>This link is missing a token.</p>
        <Link to="/forgot-password" style={{ color: '#FF3008', fontWeight: 700 }}>
          Request a new link
        </Link>
      </div>
    )
  }

  return (
    <div style={{
      minHeight: '100vh', background: '#fff',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      padding: 24,
    }}>
      <div style={{ width: '100%', maxWidth: 420 }}>
        <Link to="/login" style={{
          display: 'inline-flex', alignItems: 'center', gap: 6,
          fontSize: 13, fontWeight: 700, color: '#767676', marginBottom: 40,
        }}>
          <ArrowLeft size={15} /> Back to sign in
        </Link>

        <div style={{ fontSize: 28, fontWeight: 800, color: '#FF3008', letterSpacing: '-0.5px', marginBottom: 8 }}>
          trackr.
        </div>

        {success ? (
          <div style={{ marginTop: 32 }}>
            <div style={{ fontSize: 48, marginBottom: 24 }}>✅</div>
            <h1 style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.5px', marginBottom: 12 }}>
              Password updated!
            </h1>
            <p style={{ color: '#767676', fontSize: 15 }}>
              Your password has been changed. Redirecting you to sign in…
            </p>
          </div>
        ) : (
          <>
            <h1 style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.5px', marginBottom: 8, marginTop: 32 }}>
              Set new password
            </h1>
            <p style={{ color: '#767676', fontSize: 15, marginBottom: 32 }}>
              Choose a strong password for your Trackr account.
            </p>

            <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              {[
                { label: 'New password', value: password, set: setPassword, placeholder: 'Min. 8 characters' },
                { label: 'Confirm password', value: confirm, set: setConfirm, placeholder: 'Repeat your password' },
              ].map(({ label, value, set, placeholder }) => (
                <div key={label}>
                  <label style={{
                    display: 'block', fontSize: 13, fontWeight: 700,
                    marginBottom: 6, color: '#191919',
                  }}>
                    {label}
                  </label>
                  <input
                    type="password" required
                    value={value} onChange={e => set(e.target.value)}
                    placeholder={placeholder}
                    style={{
                      width: '100%', padding: '14px 16px', borderRadius: 10,
                      border: '1.5px solid #E8E8E8', fontSize: 15, outline: 'none',
                    }}
                    onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
                    onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'}
                  />
                </div>
              ))}

              {error && (
                <div style={{
                  background: '#FEF2F2', color: '#B91C1C',
                  padding: '12px 16px', borderRadius: 10, fontSize: 14,
                }}>
                  {error}
                </div>
              )}

              <button type="submit" disabled={loading} style={{
                padding: '15px 0', borderRadius: 12,
                background: loading ? '#ccc' : '#FF3008',
                color: '#fff', fontWeight: 800, fontSize: 16,
                boxShadow: loading ? 'none' : '0 4px 16px rgba(255,48,8,0.35)',
              }}>
                {loading ? 'Updating…' : 'Update password'}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  )
}
