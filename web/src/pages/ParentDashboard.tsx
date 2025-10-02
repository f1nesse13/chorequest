import { useMemo, useRef, useState } from 'react'
import type { SVGProps } from 'react'
import { gql } from '@apollo/client'
import { useLazyQuery, useMutation, useQuery } from '@apollo/client/react'
import { Link, useParams } from 'react-router-dom'
import { useAuth } from '../store/useAuth'
import Header from '../components/Header'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'
import Textarea from '../components/ui/Textarea'
import Select from '../components/ui/Select'
import Badge from '../components/ui/Badge'
import Label from '../components/ui/Label'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import NotificationModal from '../components/ui/NotificationModal'

const Q_CHILDREN = gql`query($parentId: ID!) { children(parentId:$parentId){ id name xp gold parentId } }`
const Q_QUESTS = gql`query($parentId: ID!) { quests(parentId:$parentId){ id title description xp gold parentId } }`
const Q_REWARDS = gql`query($parentId: ID!) { rewards(parentId:$parentId){ id name xpThreshold parentId } }`

const M_CREATE_CHILD = gql`mutation($parentId: ID!, $name: String!){ createChild(input:{parentId:$parentId,name:$name}){ id name xp gold parentId } }`
const M_UPDATE_CHILD = gql`mutation($parentId: ID!, $childId: ID!, $name: String!){ updateChild(parentId:$parentId, childId:$childId, input:{ name:$name }){ id name xp gold parentId } }`
const M_DELETE_CHILD = gql`mutation($parentId: ID!, $childId: ID!){ deleteChild(parentId:$parentId, childId:$childId) }`
const M_CREATE_QUEST = gql`mutation($parentId: ID!, $title: String!, $description: String, $xp: Int!, $gold: Int!){ createQuest(input:{parentId:$parentId,title:$title,description:$description,xp:$xp,gold:$gold}){ id title } }`
const M_CREATE_REWARD = gql`mutation($parentId: ID!, $name: String!, $xpThreshold: Int!){ createReward(input:{parentId:$parentId,name:$name,xpThreshold:$xpThreshold}){ id name } }`
const M_ASSIGN = gql`mutation($questId: ID!, $childId: ID!){ assignQuest(questId:$questId, childId:$childId){ id status childId quest{ id title } } }`
const M_UPDATE_QUEST = gql`mutation($parentId: ID!, $questId: ID!, $input: UpdateQuestInput!){ updateQuest(parentId:$parentId, questId:$questId, input:$input){ id title description xp gold parentId } }`
const M_DELETE_QUEST = gql`mutation($parentId: ID!, $questId: ID!){ deleteQuest(parentId:$parentId, questId:$questId) }`
const M_UPDATE_REWARD = gql`mutation($parentId: ID!, $rewardId: ID!, $input: UpdateRewardInput!){ updateReward(parentId:$parentId, rewardId:$rewardId, input:$input){ id name xpThreshold parentId } }`
const M_DELETE_REWARD = gql`mutation($parentId: ID!, $rewardId: ID!){ deleteReward(parentId:$parentId, rewardId:$rewardId) }`
const Q_SUB = gql`query($parentId: ID!){ subscriptionStatus(parentId:$parentId){ active currentPeriodEnd } }`
const M_CHECKOUT = gql`mutation($parentId: ID!, $success: String!, $cancel: String!){ createCheckoutSession(parentId:$parentId, successUrl:$success, cancelUrl:$cancel) }`
const Q_CHILD_ASSIGNMENTS = gql`query($childId: ID!){ myAssignments(childId:$childId){ id status quest{ id title } } }`
const M_DELETE_ASSIGNMENT = gql`mutation($assignmentId: ID!){ deleteAssignment(assignmentId:$assignmentId) }`

function PencilIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M4 21h4l11-11a2.828 2.828 0 0 0-4-4L4 17v4z" />
      <path d="M14 7l3 3" />
    </svg>
  )
}

function TrashIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="3 6 5 6 21 6" />
      <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" />
      <path d="M10 11v6" />
      <path d="M14 11v6" />
      <path d="M15 6V4a2 2 0 0 0-2-2h-2a2 2 0 0 0-2 2v2" />
    </svg>
  )
}

function CheckIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M20 6L9 17l-5-5" />
    </svg>
  )
}

function XIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M18 6L6 18" />
      <path d="M6 6l12 12" />
    </svg>
  )
}

type NotificationState = {
  title: string
  description: string
  confirmLabel?: string
  onConfirm?: () => Promise<void> | void
}

type AssignmentSummary = {
  id: string
  questId: string
  questTitle: string
  status: string
}

export default function ParentDashboard(){
  const { parentId = 'parent-1' } = useParams()
  const auth = useAuth()
  const [fetchChildAssignments] = useLazyQuery(Q_CHILD_ASSIGNMENTS, { fetchPolicy: 'network-only' })
  const [childName, setChildName] = useState('')
  const [editingChildId, setEditingChildId] = useState<string | null>(null)
  const [childEditName, setChildEditName] = useState('')
  const [questTitle, setQuestTitle] = useState('')
  const [questDesc, setQuestDesc] = useState('')
  const [questXP, setQuestXP] = useState(50)
  const [questGold, setQuestGold] = useState(10)
  const [rewardName, setRewardName] = useState('Movie Night')
  const [rewardXP, setRewardXP] = useState(200)
  const [editingQuestId, setEditingQuestId] = useState<string | null>(null)
  const [questEditDraft, setQuestEditDraft] = useState({ title: '', description: '', xp: 0, gold: 0 })
  const [editingRewardId, setEditingRewardId] = useState<string | null>(null)
  const [rewardEditDraft, setRewardEditDraft] = useState({ name: '', xpThreshold: 0 })
  const [openQuestActionsId, setOpenQuestActionsId] = useState<string | null>(null)
  const [openRewardActionsId, setOpenRewardActionsId] = useState<string | null>(null)
  const [showAddChildForm, setShowAddChildForm] = useState(false)
  const [showAddQuestForm, setShowAddQuestForm] = useState(false)
  const [showAddRewardForm, setShowAddRewardForm] = useState(false)
  const [notification, setNotification] = useState<NotificationState | null>(null)
  const assignmentsCache = useRef<Record<string, { quests: Set<string>; items: AssignmentSummary[] }>>({})
  const [activeAssignmentsChildId, setActiveAssignmentsChildId] = useState<string | null>(null)
  const [activeAssignments, setActiveAssignments] = useState<AssignmentSummary[]>([])

  const { data: dc, refetch: refetchChildren } = useQuery(Q_CHILDREN, { variables: { parentId } })
  const { data: dq, refetch: refetchQuests } = useQuery(Q_QUESTS, { variables: { parentId } })
  const { data: dr, refetch: refetchRewards } = useQuery(Q_REWARDS, { variables: { parentId } })
  const { data: ds } = useQuery(Q_SUB, { variables: { parentId } })

  const [createChild] = useMutation(M_CREATE_CHILD, { onCompleted: () => { setChildName(''); setShowAddChildForm(false); refetchChildren() } })
  const [updateChild] = useMutation(M_UPDATE_CHILD, { onCompleted: () => { setEditingChildId(null); setChildEditName(''); refetchChildren() } })
  const [deleteChildMutation] = useMutation(M_DELETE_CHILD, { onCompleted: () => { refetchChildren() } })
  const [createQuest] = useMutation(M_CREATE_QUEST, { onCompleted: () => { setQuestTitle(''); setQuestDesc(''); setShowAddQuestForm(false); refetchQuests() } })
  const [createReward] = useMutation(M_CREATE_REWARD, { onCompleted: () => { setShowAddRewardForm(false); setRewardName('Movie Night'); setRewardXP(200); refetchRewards() } })
  const [assignQuest] = useMutation(M_ASSIGN, { onCompleted: () => {} })
  const [updateQuest] = useMutation(M_UPDATE_QUEST, { onCompleted: () => { setEditingQuestId(null); setQuestEditDraft({ title: '', description: '', xp: 0, gold: 0 }); refetchQuests() } })
  const [deleteQuestMutation] = useMutation(M_DELETE_QUEST, { onCompleted: () => { refetchQuests() } })
  const [updateReward] = useMutation(M_UPDATE_REWARD, { onCompleted: () => { setEditingRewardId(null); setRewardEditDraft({ name: '', xpThreshold: 0 }); refetchRewards() } })
  const [deleteRewardMutation] = useMutation(M_DELETE_REWARD, { onCompleted: () => { refetchRewards() } })
  const [checkout] = useMutation(M_CHECKOUT)
  const [deleteAssignmentMutation] = useMutation(M_DELETE_ASSIGNMENT)

  const children = useMemo(() => (dc as any)?.children ?? [], [dc])
  const quests = useMemo(() => (dq as any)?.quests ?? [], [dq])
  const rewards = useMemo(() => (dr as any)?.rewards ?? [], [dr])
  const sub = (ds as any)?.subscriptionStatus

  const pendingResetRef = useRef<(() => void) | null>(null)

  const hideNotification = () => {
    if (pendingResetRef.current) {
      pendingResetRef.current()
      pendingResetRef.current = null
    }
    setNotification(null)
  }

  const loadAssignmentsForChild = async (childId: string) => {
    const cached = assignmentsCache.current[childId]
    if (cached) return cached
    const res = await fetchChildAssignments({ variables: { childId } })
    const list = (res.data as any)?.myAssignments ?? []
    const items: AssignmentSummary[] = []
    const quests = new Set<string>()
    list.forEach((item: any) => {
      const questId = item?.quest?.id
      if (questId) {
        quests.add(questId)
      }
      items.push({
        id: item?.id ?? '',
        questId: questId ?? '',
        questTitle: item?.quest?.title ?? 'Unknown quest',
        status: item?.status ?? 'ASSIGNED',
      })
    })
    const entry = { quests, items }
    assignmentsCache.current[childId] = entry
    return entry
  }

  const handleRequestAssign = async (quest: any, childId: string, childName: string, resetSelect: () => void) => {
    if (!childId) return
    try {
      const assignmentsEntry = await loadAssignmentsForChild(childId)
      if (assignmentsEntry.quests.has(quest.id)) {
        pendingResetRef.current = null
        setNotification({
          title: 'Already assigned',
          description: `${childName} already has "${quest.title}" assigned.`,
        })
        resetSelect()
        return
      }
      pendingResetRef.current = resetSelect
      setNotification({
        title: 'Assign quest?',
        description: `Assign "${quest.title}" to ${childName}?`,
        confirmLabel: 'Assign quest',
        onConfirm: async () => {
          try {
            const result = await assignQuest({ variables: { questId: quest.id, childId } })
            const payload = (result.data as any)?.assignQuest
            const cacheEntry = assignmentsCache.current[childId] ?? { quests: new Set<string>(), items: [] as AssignmentSummary[] }
            cacheEntry.quests.add(quest.id)
            const summary: AssignmentSummary = {
              id: payload?.id ?? `assignment-${Date.now()}`,
              questId: quest.id,
              questTitle: quest.title ?? 'Quest',
              status: payload?.status ?? 'ASSIGNED',
            }
            cacheEntry.items = [summary, ...cacheEntry.items.filter(item => item.id !== summary.id)]
            assignmentsCache.current[childId] = cacheEntry
            if (activeAssignmentsChildId === childId) {
              setActiveAssignments([...cacheEntry.items])
            }
            resetSelect()
            pendingResetRef.current = null
            hideNotification()
          } catch (err: any) {
            resetSelect()
            pendingResetRef.current = null
            const message = typeof err?.message === 'string' ? err.message : 'Failed to assign quest.'
            setNotification({
              title: 'Assignment failed',
              description: message,
            })
          }
        },
      })
    } catch (err: any) {
      pendingResetRef.current = null
      const message = typeof err?.message === 'string' ? err.message : 'Unable to check assignments.'
      setNotification({
        title: 'Unable to assign quest',
        description: message,
      })
      resetSelect()
    }
  }

  const toggleAssignmentsForChild = async (child: any) => {
    if (!child?.id) return
    if (activeAssignmentsChildId === child.id) {
      setActiveAssignmentsChildId(null)
      setActiveAssignments([])
      return
    }
    try {
      const data = await loadAssignmentsForChild(child.id)
      setActiveAssignmentsChildId(child.id)
      setActiveAssignments([...data.items])
    } catch (err: any) {
      const message = typeof err?.message === 'string' ? err.message : 'Unable to load assignments.'
      setNotification({
        title: 'Unable to load assignments',
        description: message,
      })
    }
  }

  const requestRemoveAssignment = (childId: string, childName: string, assignment: AssignmentSummary) => {
    pendingResetRef.current = null
    setNotification({
      title: 'Remove assignment?',
      description: `Remove "${assignment.questTitle}" from ${childName}?`,
      confirmLabel: 'Remove',
      onConfirm: async () => {
        try {
          await deleteAssignmentMutation({ variables: { assignmentId: assignment.id } })
          const cacheEntry = assignmentsCache.current[childId]
          if (cacheEntry) {
            if (assignment.questId) {
              cacheEntry.quests.delete(assignment.questId)
            }
            cacheEntry.items = cacheEntry.items.filter(item => item.id !== assignment.id)
            assignmentsCache.current[childId] = cacheEntry
            if (activeAssignmentsChildId === childId) {
              setActiveAssignments([...cacheEntry.items])
            }
          }
          hideNotification()
        } catch (err: any) {
          const message = typeof err?.message === 'string' ? err.message : 'Failed to remove assignment.'
          setNotification({
            title: 'Remove assignment failed',
            description: message,
          })
        }
      },
    })
  }

  const startChildEdit = (child: any) => {
    setEditingChildId(child.id)
    setChildEditName(child.name ?? '')
    setShowAddChildForm(false)
  }

  const cancelChildEdit = () => {
    setEditingChildId(null)
    setChildEditName('')
  }

  const requestDeleteChild = (child: any) => {
    if (!child?.id) return
    pendingResetRef.current = null
    setNotification({
      title: 'Delete child?',
      description: `Delete ${child.name ?? 'this child'}? This will remove their quests and assignments.`,
      confirmLabel: 'Delete child',
      onConfirm: async () => {
        try {
          if (editingChildId === child.id) {
            cancelChildEdit()
          }
          await deleteChildMutation({ variables: { parentId, childId: child.id } })
          delete assignmentsCache.current[child.id]
          if (activeAssignmentsChildId === child.id) {
            setActiveAssignmentsChildId(null)
            setActiveAssignments([])
          }
          hideNotification()
        } catch (err: any) {
          const message = typeof err?.message === 'string' ? err.message : 'Failed to delete child.'
          setNotification({
            title: 'Delete child failed',
            description: message,
          })
        }
      },
    })
  }

  const startQuestEdit = (quest: any) => {
    setOpenQuestActionsId(null)
    setEditingQuestId(quest.id)
    setQuestEditDraft({
      title: quest.title ?? '',
      description: quest.description ?? '',
      xp: quest.xp ?? 0,
      gold: quest.gold ?? 0,
    })
  }

  const cancelQuestEdit = () => {
    setEditingQuestId(null)
    setQuestEditDraft({ title: '', description: '', xp: 0, gold: 0 })
  }

  const handleDeleteQuest = async (questId: string) => {
    if (!questId) return
    if (typeof window !== 'undefined' && !window.confirm('Delete this quest?')) return
    setOpenQuestActionsId(null)
    await deleteQuestMutation({ variables: { parentId, questId } })
  }

  const startRewardEdit = (reward: any) => {
    setOpenRewardActionsId(null)
    setEditingRewardId(reward.id)
    setRewardEditDraft({
      name: reward.name ?? '',
      xpThreshold: reward.xpThreshold ?? 0,
    })
  }

  const cancelRewardEdit = () => {
    setEditingRewardId(null)
    setRewardEditDraft({ name: '', xpThreshold: 0 })
  }

  const handleDeleteReward = async (rewardId: string) => {
    if (!rewardId) return
    if (typeof window !== 'undefined' && !window.confirm('Delete this reward?')) return
    setOpenRewardActionsId(null)
    await deleteRewardMutation({ variables: { parentId, rewardId } })
  }

  return (
    <div>
      <Header />
      <div className="max-w-6xl mx-auto p-6 space-y-8">
      <header className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Parent Dashboard</h1>
        <div className="text-sm text-zinc-600">Parent: {parentId}{auth?.id?` • signed in as ${auth.role} ${auth.id}`:''}</div>
      </header>

      <section className="grid md:grid-cols-3 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Subscription</CardTitle>
          </CardHeader>
          <CardContent>
          {sub ? (
            <div className="text-sm">Status: {sub.active ? <Badge tone="success">Active</Badge> : <Badge tone="warning">Inactive</Badge>} {sub.currentPeriodEnd && `(until ${sub.currentPeriodEnd})`}</div>
          ) : null}
          <Button className="mt-3" onClick={async ()=>{
            const success = location.origin + '/parent/' + parentId
            const cancel = location.href
            const res = await checkout({ variables: { parentId, success, cancel } })
            const url = (res as any).data?.createCheckoutSession || 'https://example.com/checkout'
            location.assign(url)
          }}>Subscribe</Button>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Children</CardTitle>
          </CardHeader>
          <CardContent>
          <ul className="mt-3 space-y-2">
            {children.map((c:any) => (
              <li key={c.id} className="rounded border border-transparent p-3 transition-colors hover:border-zinc-200">
                {editingChildId === c.id ? (
                  <form
                    className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between"
                    onSubmit={async e => {
                      e.preventDefault()
                      if (!editingChildId) return
                      await updateChild({ variables: { parentId, childId: editingChildId, name: childEditName } })
                    }}
                  >
                    <Input value={childEditName} onChange={e=>setChildEditName(e.target.value)} placeholder="Child name" autoFocus />
                    <div className="flex gap-2">
                      <button
                        type="submit"
                        className="flex h-9 w-9 items-center justify-center rounded bg-emerald-600 text-white transition-colors hover:bg-emerald-500 focus:outline-none focus:ring-2 focus:ring-emerald-300"
                        aria-label="Save child"
                      >
                        <CheckIcon className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        className="flex h-9 w-9 items-center justify-center rounded bg-rose-600 text-white transition-colors hover:bg-rose-500 focus:outline-none focus:ring-2 focus:ring-rose-300"
                        aria-label="Cancel edit child"
                        onClick={cancelChildEdit}
                      >
                        <XIcon className="h-4 w-4" />
                      </button>
                    </div>
                  </form>
                ) : (
                  <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                    <div className="text-sm">
                      <span className="font-medium">{c.name}</span>{' '}
                      <span className="text-xs text-zinc-500">(xp {c.xp}, gold {c.gold})</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Link className="text-indigo-600 text-sm" to={`/child/${c.id}`}>View</Link>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="h-8 w-8 p-0 text-zinc-600 hover:text-indigo-600"
                        aria-label="Edit child"
                        onClick={() => startChildEdit(c)}
                      >
                        <PencilIcon className="h-4 w-4" />
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="h-8 px-2 text-zinc-600 hover:text-indigo-600"
                        onClick={() => void toggleAssignmentsForChild(c)}
                      >
                        {activeAssignmentsChildId === c.id ? 'Hide assignments' : 'Manage assignments'}
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="h-8 w-8 p-0 text-rose-600 hover:text-rose-700"
                        aria-label="Delete child"
                        onClick={() => requestDeleteChild(c)}
                      >
                        <TrashIcon className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                )}
                {activeAssignmentsChildId === c.id && (
                  <div className="mt-3 space-y-2 rounded border border-zinc-200 bg-zinc-50 p-3">
                    <div className="text-xs font-medium uppercase tracking-wide text-zinc-600">Assignments</div>
                    {activeAssignments.length === 0 ? (
                      <div className="text-xs text-zinc-500">No quests assigned.</div>
                    ) : (
                      <ul className="space-y-2">
                        {activeAssignments.map((assignment) => (
                          <li key={assignment.id} className="flex items-center justify-between gap-3">
                            <div>
                              <div className="text-sm font-medium text-zinc-800">{assignment.questTitle}</div>
                              <div className="text-xs uppercase tracking-wide text-zinc-500">{assignment.status}</div>
                            </div>
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              className="text-rose-600 hover:text-rose-700"
                              onClick={() => requestRemoveAssignment(c.id, c.name, assignment)}
                            >
                              Remove
                            </Button>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                )}
              </li>
            ))}
          </ul>
          {!showAddChildForm ? (
            <Button type="button" className="mt-3" onClick={() => setShowAddChildForm(true)}>Add Child</Button>
          ) : (
            <form className="mt-3 space-y-2" onSubmit={e=>{e.preventDefault(); createChild({ variables: { parentId, name: childName }})}}>
              <Label className="mb-1">New child name</Label>
              <Input value={childName} onChange={e=>setChildName(e.target.value)} placeholder="e.g. Alex"/>
              <div className="flex gap-2">
                <button
                  type="submit"
                  className="flex h-9 w-9 items-center justify-center rounded bg-emerald-600 text-white transition-colors hover:bg-emerald-500 focus:outline-none focus:ring-2 focus:ring-emerald-300"
                  aria-label="Confirm add child"
                >
                  <CheckIcon className="h-4 w-4" />
                </button>
                <button
                  type="button"
                  className="flex h-9 w-9 items-center justify-center rounded bg-rose-600 text-white transition-colors hover:bg-rose-500 focus:outline-none focus:ring-2 focus:ring-rose-300"
                  aria-label="Cancel add child"
                  onClick={() => {
                    setChildName('')
                    setShowAddChildForm(false)
                  }}
                >
                  <XIcon className="h-4 w-4" />
                </button>
              </div>
            </form>
          )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Quests</CardTitle>
          </CardHeader>
          <CardContent>
          <ul className="mt-3 space-y-3">
            {quests.map((q:any) => (
              <li key={q.id} className="space-y-3 rounded border border-zinc-200 p-3">
                <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <div>
                    <div className="font-medium">{q.title}</div>
                    <div className="text-xs text-zinc-500">XP {q.xp} • Gold {q.gold}</div>
                    {q.description ? <div className="mt-1 text-xs text-zinc-500">{q.description}</div> : null}
                  </div>
                  <div className="flex flex-wrap items-center gap-2 md:justify-end">
                    <Select
                      defaultValue=""
                      className="md:w-40"
                      onChange={e => {
                        const selectEl = e.target as HTMLSelectElement
                        const childId = selectEl.value
                        if (!childId) return
                        const child = children.find((c: any) => c.id === childId)
                        const childName = child?.name ?? 'this child'
                        void handleRequestAssign(q, childId, childName, () => {
                          selectEl.selectedIndex = 0
                        })
                      }}
                    >
                      <option value="">Assign to…</option>
                      {children.map((c:any)=>(<option key={c.id} value={c.id}>{c.name}</option>))}
                    </Select>
                    <div className="relative">
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="h-8 w-8 p-0 text-zinc-600 hover:text-indigo-600"
                        aria-haspopup="menu"
                        aria-expanded={openQuestActionsId === q.id}
                        aria-label="Quest actions"
                        onClick={() => {
                          setOpenRewardActionsId(null)
                          setOpenQuestActionsId(prev => prev === q.id ? null : q.id)
                        }}
                      >
                        <PencilIcon className="h-4 w-4" />
                      </Button>
                      {openQuestActionsId === q.id ? (
                        <div className="absolute right-0 top-full z-10 mt-1 w-36 rounded border border-zinc-200 bg-white p-1 shadow-lg">
                          <button
                            type="button"
                            className="flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-zinc-700 transition-colors hover:bg-indigo-50 focus:outline-none focus:ring-2 focus:ring-indigo-200"
                            onClick={() => startQuestEdit(q)}
                          >
                            <PencilIcon className="h-4 w-4" />
                            Edit
                          </button>
                          <button
                            type="button"
                            className="flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-rose-600 transition-colors hover:bg-rose-50 focus:outline-none focus:ring-2 focus:ring-rose-200"
                            onClick={() => handleDeleteQuest(q.id)}
                          >
                            <TrashIcon className="h-4 w-4" />
                            Delete
                          </button>
                        </div>
                      ) : null}
                    </div>
                  </div>
                </div>
                {editingQuestId === q.id && (
                  <form
                    className="grid gap-2 md:grid-cols-4 md:items-end"
                    onSubmit={async e => {
                      e.preventDefault()
                      if (!editingQuestId) return
                      await updateQuest({
                        variables: {
                          parentId,
                          questId: editingQuestId,
                          input: {
                            title: questEditDraft.title,
                            description: questEditDraft.description ? questEditDraft.description : null,
                            xp: questEditDraft.xp,
                            gold: questEditDraft.gold,
                          },
                        },
                      })
                    }}
                  >
                    <div className="md:col-span-4">
                      <Label className="mb-1">Title</Label>
                      <Input value={questEditDraft.title} onChange={e=>setQuestEditDraft(prev => ({ ...prev, title: e.target.value }))} required />
                    </div>
                    <div className="md:col-span-4">
                      <Label className="mb-1">Description</Label>
                      <Textarea rows={3} value={questEditDraft.description} onChange={e=>setQuestEditDraft(prev => ({ ...prev, description: e.target.value }))} placeholder="Optional description" />
                    </div>
                    <div>
                      <Label className="mb-1">XP</Label>
                      <Input type="number" value={questEditDraft.xp} onChange={e=>setQuestEditDraft(prev => ({ ...prev, xp: parseInt(e.target.value || '0', 10) }))} />
                    </div>
                    <div>
                      <Label className="mb-1">Gold</Label>
                      <Input type="number" value={questEditDraft.gold} onChange={e=>setQuestEditDraft(prev => ({ ...prev, gold: parseInt(e.target.value || '0', 10) }))} />
                    </div>
                    <div className="flex gap-2 md:col-span-4">
                      <Button type="submit">Save</Button>
                      <Button type="button" variant="secondary" onClick={cancelQuestEdit}>Cancel</Button>
                    </div>
                  </form>
                )}
              </li>
            ))}
            {quests.length === 0 && <li className="text-sm text-zinc-500">No quests yet.</li>}
          </ul>
          {!showAddQuestForm ? (
            <Button type="button" className="mt-3" onClick={() => setShowAddQuestForm(true)}>Add Quest</Button>
          ) : (
            <form className="mt-3 space-y-2" onSubmit={e=>{e.preventDefault(); createQuest({ variables: { parentId, title: questTitle, description: questDesc || null, xp: questXP, gold: questGold }})}}>
              <Label className="mb-1">Quest title</Label>
              <Input value={questTitle} onChange={e=>setQuestTitle(e.target.value)} placeholder="e.g. Clean Room"/>
              <Label className="mb-1">Description (optional)</Label>
              <Textarea value={questDesc} onChange={e=>setQuestDesc(e.target.value)} placeholder="e.g. Tidy up and vacuum" rows={3}/>
              <div className="grid gap-2 md:grid-cols-2">
                <div>
                  <Label className="mb-1">XP Reward</Label>
                  <Input type="number" value={questXP} onChange={e=>setQuestXP(parseInt(e.target.value||'0'))} placeholder="e.g. 50"/>
                </div>
                <div>
                  <Label className="mb-1">Gold Reward</Label>
                  <Input type="number" value={questGold} onChange={e=>setQuestGold(parseInt(e.target.value||'0'))} placeholder="e.g. 10"/>
                </div>
              </div>
              <div className="flex gap-2">
                <button
                  type="submit"
                  className="flex h-9 w-9 items-center justify-center rounded bg-emerald-600 text-white transition-colors hover:bg-emerald-500 focus:outline-none focus:ring-2 focus:ring-emerald-300"
                  aria-label="Confirm add quest"
                >
                  <CheckIcon className="h-4 w-4" />
                </button>
                <button
                  type="button"
                  className="flex h-9 w-9 items-center justify-center rounded bg-rose-600 text-white transition-colors hover:bg-rose-500 focus:outline-none focus:ring-2 focus:ring-rose-300"
                  aria-label="Cancel add quest"
                  onClick={() => {
                    setQuestTitle('')
                    setQuestDesc('')
                    setQuestXP(50)
                    setQuestGold(10)
                    setShowAddQuestForm(false)
                  }}
                >
                  <XIcon className="h-4 w-4" />
                </button>
              </div>
            </form>
          )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Rewards</CardTitle>
          </CardHeader>
          <CardContent>
          <ul className="mt-3 space-y-3">
            {rewards.map((r:any) => (
              <li key={r.id} className="space-y-3 rounded border border-zinc-200 p-3">
                <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <div>
                    <div className="font-medium">{r.name}</div>
                    <div className="text-xs text-zinc-500">Requires {r.xpThreshold}+ XP</div>
                  </div>
                  <div className="flex flex-wrap items-center gap-2 md:justify-end">
                    <div className="relative">
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="h-8 w-8 p-0 text-zinc-600 hover:text-indigo-600"
                        aria-haspopup="menu"
                        aria-expanded={openRewardActionsId === r.id}
                        aria-label="Reward actions"
                        onClick={() => {
                          setOpenQuestActionsId(null)
                          setOpenRewardActionsId(prev => prev === r.id ? null : r.id)
                        }}
                      >
                        <PencilIcon className="h-4 w-4" />
                      </Button>
                      {openRewardActionsId === r.id ? (
                        <div className="absolute right-0 top-full z-10 mt-1 w-36 rounded border border-zinc-200 bg-white p-1 shadow-lg">
                          <button
                            type="button"
                            className="flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-zinc-700 transition-colors hover:bg-indigo-50 focus:outline-none focus:ring-2 focus:ring-indigo-200"
                            onClick={() => startRewardEdit(r)}
                          >
                            <PencilIcon className="h-4 w-4" />
                            Edit
                          </button>
                          <button
                            type="button"
                            className="flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-rose-600 transition-colors hover:bg-rose-50 focus:outline-none focus:ring-2 focus:ring-rose-200"
                            onClick={() => handleDeleteReward(r.id)}
                          >
                            <TrashIcon className="h-4 w-4" />
                            Delete
                          </button>
                        </div>
                      ) : null}
                    </div>
                  </div>
                </div>
                {editingRewardId === r.id && (
                  <form
                    className="grid gap-2 md:grid-cols-3 md:items-end"
                    onSubmit={async e => {
                      e.preventDefault()
                      if (!editingRewardId) return
                      await updateReward({
                        variables: {
                          parentId,
                          rewardId: editingRewardId,
                          input: {
                            name: rewardEditDraft.name,
                            xpThreshold: rewardEditDraft.xpThreshold,
                          },
                        },
                      })
                    }}
                  >
                    <div className="md:col-span-2">
                      <Label className="mb-1">Name</Label>
                      <Input value={rewardEditDraft.name} onChange={e=>setRewardEditDraft(prev => ({ ...prev, name: e.target.value }))} required />
                    </div>
                    <div>
                      <Label className="mb-1">XP Threshold</Label>
                      <Input type="number" value={rewardEditDraft.xpThreshold} onChange={e=>setRewardEditDraft(prev => ({ ...prev, xpThreshold: parseInt(e.target.value || '0', 10) }))} />
                    </div>
                    <div className="flex gap-2 md:col-span-3">
                      <Button type="submit">Save</Button>
                      <Button type="button" variant="secondary" onClick={cancelRewardEdit}>Cancel</Button>
                    </div>
                  </form>
                )}
              </li>
            ))}
            {rewards.length === 0 && <li className="text-sm text-zinc-500">No rewards yet.</li>}
          </ul>
          {!showAddRewardForm ? (
            <Button type="button" className="mt-3" onClick={() => setShowAddRewardForm(true)}>Add Reward</Button>
          ) : (
            <form className="mt-3 space-y-2" onSubmit={e=>{e.preventDefault(); createReward({ variables: { parentId, name: rewardName, xpThreshold: rewardXP }})}}>
              <Label className="mb-1">Reward name</Label>
              <Input value={rewardName} onChange={e=>setRewardName(e.target.value)} placeholder="e.g. Movie Night"/>
              <Label className="mb-1">XP threshold</Label>
              <Input type="number" value={rewardXP} onChange={e=>setRewardXP(parseInt(e.target.value||'0'))} placeholder="XP threshold"/>
              <div className="flex gap-2">
                <button
                  type="submit"
                  className="flex h-9 w-9 items-center justify-center rounded bg-emerald-600 text-white transition-colors hover:bg-emerald-500 focus:outline-none focus:ring-2 focus:ring-emerald-300"
                  aria-label="Confirm add reward"
                >
                  <CheckIcon className="h-4 w-4" />
                </button>
                <button
                  type="button"
                  className="flex h-9 w-9 items-center justify-center rounded bg-rose-600 text-white transition-colors hover:bg-rose-500 focus:outline-none focus:ring-2 focus:ring-rose-300"
                  aria-label="Cancel add reward"
                  onClick={() => {
                    setRewardName('Movie Night')
                    setRewardXP(200)
                    setShowAddRewardForm(false)
                  }}
                >
                  <XIcon className="h-4 w-4" />
                </button>
              </div>
            </form>
          )}
          </CardContent>
        </Card>
      </section>
      </div>
      <NotificationModal
        open={!!notification}
        title={notification?.title ?? ''}
        description={notification?.description ?? ''}
        confirmLabel={notification?.confirmLabel}
        onConfirm={notification?.onConfirm}
        onClose={hideNotification}
      />
    </div>
  )
}
