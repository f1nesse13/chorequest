import { SelectHTMLAttributes, forwardRef } from 'react'

type Props = SelectHTMLAttributes<HTMLSelectElement> & {
  invalid?: boolean
}

const Select = forwardRef<HTMLSelectElement, Props>(({ className = '', invalid, ...props }, ref) => {
  const base = 'w-full rounded border px-3 py-2 outline-none focus:ring-2 focus:ring-indigo-400 bg-white'
  const border = invalid ? 'border-rose-400' : 'border-zinc-300'
  return <select ref={ref} className={[base, border, className].join(' ')} {...props} />
})

export default Select

