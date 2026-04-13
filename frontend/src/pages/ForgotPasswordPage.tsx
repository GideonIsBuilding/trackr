import { useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { forgotPassword } from '@/api/client'

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [submitted, setSubmitted] = useState(false)
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    try {
      await forgotPassword(email)
      setSubmitted(true)
    } finally {
      setLoading(false)
    }
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

        <div style={{
          fontSize: 28, fontWeight: 800, color: '#FF3008',
          letterSpacing: '-0.5px', marginBottom: 8,
        }}>
          trackr.
        </div>

        {submitted ? (
          <div style={{ marginTop: 32 }}>
            <div style={{
              width: 56, height: 56, borderRadius: 16,
              background: '#ECFDF5', display: 'flex',
              alignItems: 'center', justifyContent: 'center',
              fontSize: 28, marginBottom: 24,
            }}>
              ✉️
            </div>
            <h1 style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.5px', marginBottom: 12 }}>
              Check your email
            </h1>
            <p style={{ fontSize: 15, color: '#767676', lineHeight: 1.6 }}>
              If <strong>{email}</strong> is registered with Trackr, you'll receive a password reset link shortly.
            </p>
            <p style={{ fontSize: 13, color: '#9CA3AF', marginTop: 16 }}>
              Didn't get it? Check your spam folder, or{' '}
              <button onClick={() => setSubmitted(false)} style={{
                background: 'none', color: '#FF3008', fontWeight: 700,
                fontSize: 13, padding: 0,
              }}>
                try again
              </button>.
            </p>
          </div>
        ) : (
          <>
            <h1 style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.5px', marginBottom: 8, marginTop: 32 }}>
              Forgot password?
            </h1>
            <p style={{ color: '#767676', fontSize: 15, marginBottom: 32, lineHeight: 1.6 }}>
              Enter the email address you signed up with and we'll send you a reset link.
            </p>

            <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              <div>
                <label style={{
                  display: 'block', fontSize: 13, fontWeight: 700,
                  marginBottom: 6, color: '#191919',
                }}>
                  Email address
                </label>
                <input
                  type="email" required
                  value={email} onChange={e => setEmail(e.target.value)}
                  placeholder="you@example.com"
                  style={{
                    width: '100%', padding: '14px 16px', borderRadius: 10,
                    border: '1.5px solid #E8E8E8', fontSize: 15, outline: 'none',
                  }}
                  onFocus={e => e.currentTarget.style.borderColor = '#FF3008'}
                  onBlur={e => e.currentTarget.style.borderColor = '#E8E8E8'}
                />
              </div>

              <button type="submit" disabled={loading} style={{
                padding: '15px 0', borderRadius: 12,
                background: loading ? '#ccc' : '#FF3008',
                color: '#fff', fontWeight: 800, fontSize: 16,
                boxShadow: loading ? 'none' : '0 4px 16px rgba(255,48,8,0.35)',
              }}>
                {loading ? 'Sending…' : 'Send reset link'}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  )
}
