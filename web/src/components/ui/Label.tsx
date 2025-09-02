import { LabelHTMLAttributes } from 'react'

export default function Label({ className = '', ...props }: LabelHTMLAttributes<HTMLLabelElement>) {
  return <label className={["block text-sm font-medium text-zinc-700", className].join(' ')} {...props} />
}

