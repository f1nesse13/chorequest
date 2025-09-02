import { createContext, useContext, useState, ReactNode, HTMLAttributes } from 'react'

type TabsCtx = { index: number; setIndex: (i: number) => void }
const Ctx = createContext<TabsCtx | null>(null)

export function Tabs({ defaultIndex = 0, children, className = '' }: { defaultIndex?: number; children: ReactNode; className?: string }) {
  const [index, setIndex] = useState(defaultIndex)
  return <div className={className}><Ctx.Provider value={{ index, setIndex }}>{children}</Ctx.Provider></div>
}

export function TabList({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <div className={["flex gap-2 border-b", className].join(' ')}>{children}</div>
}

export function Tab({ children, idx, className = '' }: { children: ReactNode; idx: number; className?: string }) {
  const ctx = useContext(Ctx)
  if (!ctx) throw new Error('Tab must be used within Tabs')
  const selected = ctx.index === idx
  return (
    <button
      onClick={() => ctx.setIndex(idx)}
      className={[
        'px-3 py-2 -mb-px border-b-2 transition-colors',
        selected ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-zinc-600 hover:text-zinc-900',
        className,
      ].join(' ')}
    >
      {children}
    </button>
  )
}

export function TabPanels({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <div className={className}>{children}</div>
}

export function TabPanel({ children, idx, className = '' }: { children: ReactNode; idx: number; className?: string }) {
  const ctx = useContext(Ctx)
  if (!ctx) throw new Error('TabPanel must be used within Tabs')
  if (ctx.index !== idx) return null
  return <div className={className}>{children}</div>
}

