import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Trash2 } from 'lucide-react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { deleteApplication } from '@/api/client'

interface Props {
  applicationId: string
  company: string
}

export default function DeleteApplicationButton({ applicationId, company }: Props) {
  const [confirming, setConfirming] = useState(false)
  const navigate = useNavigate()
  const qc = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => deleteApplication(applicationId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['applications'] })
      navigate('/')
    },
  })

  if (confirming) {
    return (
      <div style={{
        background: '#FEF2F2', border: '1.5px solid #FECACA',
        borderRadius: 14, padding: '18px 20px',
      }}>
        <p style={{ fontSize: 14, fontWeight: 700, color: '#991B1B', marginBottom: 4 }}>
          Delete this application?
        </p>
        <p style={{ fontSize: 13, color: '#B91C1C', marginBottom: 14, lineHeight: 1.5 }}>
          This will permanently delete <strong>{company}</strong> and all its history. This cannot be undone.
        </p>
        <div style={{ display: 'flex', gap: 8 }}>
          <button onClick={() => setConfirming(false)} style={{
            flex: 1, padding: '10px 0', borderRadius: 8,
            background: '#fff', border: '1.5px solid #E8E8E8',
            fontWeight: 700, fontSize: 13, color: '#494949',
          }}>
            Cancel
          </button>
          <button
            onClick={() => mutation.mutate()}
            disabled={mutation.isPending}
            style={{
              flex: 1, padding: '10px 0', borderRadius: 8,
              background: '#DC2626', color: '#fff',
              fontWeight: 700, fontSize: 13,
            }}
          >
            {mutation.isPending ? 'Deleting…' : 'Yes, delete'}
          </button>
        </div>
      </div>
    )
  }

  return (
    <button
      onClick={() => setConfirming(true)}
      style={{
        width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center',
        gap: 8, padding: '12px 0', borderRadius: 10,
        background: '#FEF2F2', color: '#DC2626',
        fontWeight: 700, fontSize: 14,
        border: '1.5px solid #FECACA',
        transition: 'all 0.12s',
      }}
    >
      <Trash2 size={15} /> Delete application
    </button>
  )
}
