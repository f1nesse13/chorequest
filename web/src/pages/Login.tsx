import { useState } from 'react'
import { setToken } from '../lib/auth'
import { useAuth } from '../store/useAuth'

export default function Login(){
  const [role, setRole] = useState<'PARENT'|'CHILD'>('PARENT')
  const [id, setId] = useState('parent-1')
  const setAuth = useAuth(s=>s.set)

  const doLogin = async () => {
    const res = await fetch(`/auth/dev?role=${role}&sub=${encodeURIComponent(id)}`, { method: 'POST' })
    const data = await res.json()
    await setToken(data.token)
    setAuth(id, role)
    if (role === 'PARENT') {
      location.assign(`/parent/${id}`)
    } else {
      location.assign(`/child/${id}`)
    }
  }

  return (
    <main className="min-h-dvh grid place-items-center bg-zinc-50 text-zinc-900">
      <div className="p-6 rounded-xl shadow bg-white w-full max-w-sm">
        <h1 className="text-xl font-semibold">Chorequest Login</h1>
        <p className="text-sm text-zinc-600 mt-1">Dev login issues a JWT via backend</p>
        <div className="mt-4 space-y-3">
          <div>
            <label className="block text-sm mb-1">Role</label>
            <select className="w-full border rounded px-2 py-1" value={role} onChange={e=>setRole(e.target.value as any)}>
              <option value="PARENT">Parent</option>
              <option value="CHILD">Child</option>
            </select>
          </div>
          <div>
            <label className="block text-sm mb-1">ID</label>
            <input className="w-full border rounded px-2 py-1" value={id} onChange={e=>setId(e.target.value)} placeholder="parent-1 or <childId>"/>
          </div>
          <button onClick={doLogin} className="w-full px-3 py-2 rounded bg-indigo-600 text-white">Login</button>
        </div>
      </div>
    </main>
  )
}
