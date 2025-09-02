import { Link } from 'react-router-dom'
import { useAuth } from '../store/useAuth'

export default function Header() {
  const auth = useAuth()
  return (
    <header className="sticky top-0 z-40 bg-gradient-to-r from-indigo-600 to-cyan-500 text-white shadow">
      <div className="max-w-6xl mx-auto px-4 h-12 flex items-center justify-between">
        <Link to={auth.role === 'PARENT' && auth.id ? `/parent/${auth.id}` : '/login'} className="font-semibold">
          ChoreQuest
        </Link>
        <nav className="text-sm flex items-center gap-4">
          <Link to="/login" className="underline/50 hover:underline">Login</Link>
          <Link to="/parent/parent-1" className="underline/50 hover:underline">Parent</Link>
          <Link to="/child/child-1" className="underline/50 hover:underline">Child</Link>
          {auth.id && auth.role && (
            <span className="opacity-90">{auth.role} · {auth.id}</span>
          )}
          {import.meta.env.DEV && (
            <a href="http://localhost:8080/graphiql" className="underline/50 hover:underline" target="_blank" rel="noreferrer">GraphiQL</a>
          )}
        </nav>
      </div>
    </header>
  )
}
