import { HTMLAttributes } from 'react'

type Props = HTMLAttributes<HTMLSpanElement> & {
  tone?: 'neutral' | 'success' | 'warning' | 'danger' | 'info'
}

export default function Badge({ className = '', tone = 'neutral', ...props }: Props) {
  const tones = {
    neutral: 'bg-zinc-100 text-zinc-800',
    success: 'bg-emerald-100 text-emerald-800',
    warning: 'bg-amber-100 text-amber-800',
    danger: 'bg-rose-100 text-rose-800',
    info: 'bg-sky-100 text-sky-800',
  } as const
  return <span className={["inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium", tones[tone], className].join(' ')} {...props} />
}

