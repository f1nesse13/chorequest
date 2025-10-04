import { create } from 'zustand'

export type Role = 'PARENT' | 'CHILD'

type State = {
  id: string | null
  role: Role | null
  familyId: string | null
  set: (id: string, role: Role, familyId?: string | null) => void
  clear: () => void
}

const KEY = 'auth_state'

type Persisted = { id: string | null; role: Role | null; familyId: string | null }

function load(): Persisted {
  try {
    const raw = localStorage.getItem(KEY)
    if (!raw) return { id: null, role: null, familyId: null }
    return JSON.parse(raw)
  } catch {
    return { id: null, role: null, familyId: null }
  }
}

export const useAuth = create<State>((set) => ({
  ...load(),
  set: (id, role, familyId = null) => {
    localStorage.setItem(KEY, JSON.stringify({ id, role, familyId }))
    set({ id, role, familyId })
  },
  clear: () => {
    localStorage.removeItem(KEY)
    set({ id: null, role: null, familyId: null })
  },
}))
