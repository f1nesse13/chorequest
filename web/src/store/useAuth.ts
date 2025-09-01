import { create } from 'zustand'

export type Role = 'PARENT' | 'CHILD'

type State = {
  id: string | null
  role: Role | null
  set: (id: string, role: Role) => void
  clear: () => void
}

const KEY = 'auth_state'

function load(): { id: string | null; role: Role | null } {
  try {
    const raw = localStorage.getItem(KEY)
    if (!raw) return { id: null, role: null }
    return JSON.parse(raw)
  } catch {
    return { id: null, role: null }
  }
}

export const useAuth = create<State>((set) => ({
  ...load(),
  set: (id, role) => {
    localStorage.setItem(KEY, JSON.stringify({ id, role }))
    set({ id, role })
  },
  clear: () => {
    localStorage.removeItem(KEY)
    set({ id: null, role: null })
  },
}))

