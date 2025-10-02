import { ReactNode, useState } from 'react'
import Button from './Button'

interface Props {
  open: boolean
  title: string
  description: ReactNode
  confirmLabel?: string
  onConfirm?: () => Promise<void> | void
  onClose: () => void
}

export default function NotificationModal({ open, title, description, confirmLabel, onConfirm, onClose }: Props) {
  const [isProcessing, setIsProcessing] = useState(false)

  if (!open) return null

  const handleConfirm = async () => {
    if (!onConfirm) {
      onClose()
      return
    }
    try {
      setIsProcessing(true)
      await onConfirm()
    } finally {
      setIsProcessing(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4">
      <div className="w-full max-w-sm rounded-lg bg-white p-5 shadow-xl">
        <h2 className="text-lg font-semibold text-zinc-900">{title}</h2>
        <div className="mt-2 text-sm text-zinc-600">{description}</div>
        <div className="mt-4 flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose} disabled={isProcessing}>
            {confirmLabel ? 'Cancel' : 'Close'}
          </Button>
          {confirmLabel ? (
            <Button type="button" onClick={handleConfirm} disabled={isProcessing}>
              {isProcessing ? 'Working…' : confirmLabel}
            </Button>
          ) : null}
        </div>
      </div>
    </div>
  )
}
