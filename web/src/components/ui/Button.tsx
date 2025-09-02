import { ButtonHTMLAttributes } from 'react'

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
  size?: 'sm' | 'md'
}

export default function Button({ variant = 'primary', size = 'md', className = '', ...props }: Props) {
  const base = 'inline-flex items-center justify-center rounded transition-colors disabled:opacity-60 disabled:cursor-not-allowed'
  const sizes = {
    sm: 'px-2.5 py-1.5 text-sm',
    md: 'px-3 py-2',
  } as const
  const variants = {
    primary: 'bg-indigo-600 text-white hover:bg-indigo-500',
    secondary: 'bg-zinc-200 text-zinc-900 hover:bg-zinc-300',
    danger: 'bg-rose-600 text-white hover:bg-rose-500',
    ghost: 'bg-transparent text-indigo-600 hover:bg-indigo-50',
  } as const
  return <button className={[base, sizes[size], variants[variant], className].join(' ')} {...props} />
}

