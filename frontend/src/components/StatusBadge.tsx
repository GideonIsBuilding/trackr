import type { ApplicationStatus } from '@/types'
import { STATUS_LABELS } from '@/types'

const style: Record<ApplicationStatus, { bg: string; color: string }> = {
  applied:              { bg: '#EFF6FF', color: '#1D4ED8' },
  phone_screen:         { bg: '#F5F3FF', color: '#6D28D9' },
  interview:            { bg: '#FFFBEB', color: '#B45309' },
  technical_assessment: { bg: '#ECFEFF', color: '#0E7490' },
  offer:                { bg: '#ECFDF5', color: '#047857' },
  negotiating:          { bg: '#F0FDF4', color: '#15803D' },
  accepted:             { bg: '#ECFDF5', color: '#047857' },
  rejected:             { bg: '#FEF2F2', color: '#B91C1C' },
  withdrawn:            { bg: '#F9FAFB', color: '#4B5563' },
  ghosted:              { bg: '#F3F4F6', color: '#6B7280' },
}

export function StatusBadge({ status }: { status: ApplicationStatus }) {
  const s = style[status]
  return (
    <span style={{
      display: 'inline-flex',
      alignItems: 'center',
      padding: '3px 10px',
      borderRadius: 99,
      fontSize: 12,
      fontWeight: 700,
      letterSpacing: '0.02em',
      background: s.bg,
      color: s.color,
      whiteSpace: 'nowrap',
    }}>
      {STATUS_LABELS[status]}
    </span>
  )
}
