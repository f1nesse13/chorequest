import { gql } from '@apollo/client'
import { useMutation, useQuery } from '@apollo/client/react'
import { useParams, Link } from 'react-router-dom'

const Q_ASSIGNMENTS = gql`query($childId: ID!){ myAssignments(childId:$childId){ id status createdAt completedAt quest{ id title xp gold } } }`
const M_COMPLETE = gql`mutation($assignmentId: ID!){ completeAssignment(assignmentId:$assignmentId){ id status completedAt quest{ id title } } }`

export default function ChildView(){
  const { childId = '' } = useParams()
  const { data, refetch } = useQuery(Q_ASSIGNMENTS, { variables: { childId } })
  const [complete] = useMutation(M_COMPLETE, { onCompleted: () => refetch() })
  const list = (data as any)?.myAssignments ?? []
  return (
    <div className="max-w-3xl mx-auto p-6 space-y-4">
      <header className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">My Quests</h1>
        <Link to={`/parent/parent-1`} className="text-sm text-indigo-600">Parent</Link>
      </header>
      <ul className="space-y-3">
        {list.map((a:any)=> (
          <li key={a.id} className="bg-white rounded-xl shadow p-4 flex items-center justify-between">
            <div>
              <div className="font-medium">{a.quest.title}</div>
              <div className="text-xs text-zinc-500">XP {a.quest.xp} • Gold {a.quest.gold} • {a.status}</div>
            </div>
            {a.status !== 'COMPLETED' && (
              <button onClick={()=>complete({ variables: { assignmentId: a.id }})} className="px-3 py-1 rounded bg-emerald-600 text-white">Complete</button>
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}
