import { useState } from 'react'
import { setToken } from '../lib/auth'
import { useAuth } from '../store/useAuth'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'
import Label from '../components/ui/Label'
import Select from '../components/ui/Select'
import Header from '../components/Header'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'

export default function Login(){
  const [role, setRole] = useState<'PARENT'|'CHILD'>('PARENT')
  const [id, setId] = useState('parent-1')
  const setAuth = useAuth(s=>s.set)
  const [copyMsg, setCopyMsg] = useState<string>('')

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
    <div className="min-h-dvh flex flex-col bg-white text-zinc-900">
      <Header />
      <main className="flex-1 bg-gradient-to-br from-indigo-50 to-cyan-50">
        <div className="max-w-3xl mx-auto px-4 py-14">
          <div className="text-center mb-6">
            <h1 className="text-4xl font-bold tracking-tight">ChoreQuest</h1>
            <p className="text-zinc-600 mt-2">Make chores an adventure for the whole family</p>
          </div>

          <div className="grid place-items-center">
            <Card className="w-full max-w-md">
              <CardHeader>
                <CardTitle className="text-lg">Sign in</CardTitle>
                <p className="text-sm text-zinc-600 mt-1">Dev login issues a JWT via backend</p>
              </CardHeader>
              <CardContent>
                <form className="space-y-3" onSubmit={e=>{e.preventDefault(); doLogin()}}>
                  <div>
                    <Label className="mb-1">Role</Label>
                    <Select value={role} onChange={e=>setRole(e.target.value as any)}>
                      <option value="PARENT">Parent</option>
                      <option value="CHILD">Child</option>
                    </Select>
                  </div>
                  <div>
                    <Label className="mb-1">ID</Label>
                    <Input value={id} onChange={e=>setId(e.target.value)} placeholder="parent-1 or &lt;childId&gt;"/>
                  </div>
                  <Button type="submit" className="w-full">Continue</Button>
                </form>
              </CardContent>
            </Card>
          </div>

          {import.meta.env.DEV && (
            <Card className="w-full max-w-md mx-auto mt-4">
              <CardHeader>
                <CardTitle className="text-lg">Dev Tools</CardTitle>
                <p className="text-sm text-zinc-600 mt-1">Handy helpers for local development</p>
              </CardHeader>
              <CardContent>
                <div className="flex flex-col gap-2">
                  <Button variant="secondary" onClick={()=>window.open('http://localhost:8080/graphiql','_blank')}>Open GraphiQL</Button>
                  <Button onClick={async ()=>{
                    setCopyMsg('')
                    try {
                      const res = await fetch(`/auth/dev?role=${role}&sub=${encodeURIComponent(id)}`, { method: 'POST' })
                      const data = await res.json()
                      const value = `Bearer ${data.token}`
                      await navigator.clipboard.writeText(value)
                      setCopyMsg('Copied Authorization header to clipboard')
                    } catch (e) {
                      setCopyMsg('Failed to copy token')
                    }
                  }}>Copy Bearer token</Button>
                  {copyMsg && <div className="text-xs text-zinc-600">{copyMsg}</div>}
                  <div className="text-xs text-zinc-600 mt-2">
                    Tip: paste into GraphiQL’s Authorization field like
                    <code className="ml-1 px-1 py-0.5 bg-zinc-100 rounded">Bearer &lt;token&gt;</code>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </main>
    </div>
  )
}
