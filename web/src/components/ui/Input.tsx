import { InputHTMLAttributes, forwardRef } from 'react'

type Props = InputHTMLAttributes<HTMLInputElement> & {
  invalid?: boolean
}

const Input = forwardRef<HTMLInputElement, Props>(({ className = '', invalid, ...props }, ref) => {
  const base = 'w-full rounded border px-3 py-2 outline-none focus:ring-2 focus:ring-indigo-400'
  const border = invalid ? 'border-rose-400' : 'border-zinc-300'
  return <input ref={ref} className={[base, border, className].join(' ')} {...props} />
})

export default Input

