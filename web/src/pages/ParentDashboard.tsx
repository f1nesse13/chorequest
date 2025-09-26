import { useMemo, useState } from 'react'
import type { SVGProps } from 'react'
import { gql } from '@apollo/client'
import { useMutation, useQuery } from '@apollo/client/react'
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

const Q_CHILDREN = gql`query($parentId: ID!) { children(parentId:$parentId){ id name xp gold parentId } }`
const Q_QUESTS = gql`query($parentId: ID!) { quests(parentId:$parentId){ id title description xp gold parentId } }`
const Q_REWARDS = gql`query($parentId: ID!) { rewards(parentId:$parentId){ id name xpThreshold parentId } }`

const M_CREATE_CHILD = gql`mutation($parentId: ID!, $name: String!){ createChild(input:{parentId:$parentId,name:$name}){ id name } }`
const M_CREATE_QUEST = gql`mutation($parentId: ID!, $title: String!, $description: String, $xp: Int!, $gold: Int!){ createQuest(input:{parentId:$parentId,title:$title,description:$description,xp:$xp,gold:$gold}){ id title } }`
const M_CREATE_REWARD = gql`mutation($parentId: ID!, $name: String!, $xpThreshold: Int!){ createReward(input:{parentId:$parentId,name:$name,xpThreshold:$xpThreshold}){ id name } }`
const M_ASSIGN = gql`mutation($questId: ID!, $childId: ID!){ assignQuest(questId:$questId, childId:$childId){ id status childId quest{ id title } } }`
const M_UPDATE_QUEST = gql`mutation($parentId: ID!, $questId: ID!, $input: UpdateQuestInput!){ updateQuest(parentId:$parentId, questId:$questId, input:$input){ id title description xp gold parentId } }`
const M_DELETE_QUEST = gql`mutation($parentId: ID!, $questId: ID!){ deleteQuest(parentId:$parentId, questId:$questId) }`
const M_UPDATE_REWARD = gql`mutation($parentId: ID!, $rewardId: ID!, $input: UpdateRewardInput!){ updateReward(parentId:$parentId, rewardId:$rewardId, input:$input){ id name xpThreshold parentId } }`
const M_DELETE_REWARD = gql`mutation($parentId: ID!, $rewardId: ID!){ deleteReward(parentId:$parentId, rewardId:$rewardId) }`
const Q_SUB = gql`query($parentId: ID!){ subscriptionStatus(parentId:$parentId){ active currentPeriodEnd } }`
const M_CHECKOUT = gql`mutation($parentId: ID!, $success: String!, $cancel: String!){ createCheckoutSession(parentId:$parentId, successUrl:$success, cancelUrl:$cancel) }`

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

export default function ParentDashboard(){
  const { parentId = 'parent-1' } = useParams()
  const auth = useAuth()
  const [childName, setChildName] = useState('')
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

  const { data: dc, refetch: refetchChildren } = useQuery(Q_CHILDREN, { variables: { parentId } })
  const { data: dq, refetch: refetchQuests } = useQuery(Q_QUESTS, { variables: { parentId } })
  const { data: dr, refetch: refetchRewards } = useQuery(Q_REWARDS, { variables: { parentId } })
  const { data: ds } = useQuery(Q_SUB, { variables: { parentId } })

  const [createChild] = useMutation(M_CREATE_CHILD, { onCompleted: () => { setChildName(''); refetchChildren() } })
  const [createQuest] = useMutation(M_CREATE_QUEST, { onCompleted: () => { setQuestTitle(''); setQuestDesc(''); refetchQuests() } })
  const [createReward] = useMutation(M_CREATE_REWARD, { onCompleted: () => { refetchRewards() } })
  const [assignQuest] = useMutation(M_ASSIGN, { onCompleted: () => {} })
  const [updateQuest] = useMutation(M_UPDATE_QUEST, { onCompleted: () => { setEditingQuestId(null); setQuestEditDraft({ title: '', description: '', xp: 0, gold: 0 }); refetchQuests() } })
  const [deleteQuestMutation] = useMutation(M_DELETE_QUEST, { onCompleted: () => { refetchQuests() } })
  const [updateReward] = useMutation(M_UPDATE_REWARD, { onCompleted: () => { setEditingRewardId(null); setRewardEditDraft({ name: '', xpThreshold: 0 }); refetchRewards() } })
  const [deleteRewardMutation] = useMutation(M_DELETE_REWARD, { onCompleted: () => { refetchRewards() } })
  const [checkout] = useMutation(M_CHECKOUT)

  const children = useMemo(() => (dc as any)?.children ?? [], [dc])
  const quests = useMemo(() => (dq as any)?.quests ?? [], [dq])
  const rewards = useMemo(() => (dr as any)?.rewards ?? [], [dr])
  const sub = (ds as any)?.subscriptionStatus

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
              <li key={c.id} className="flex items-center justify-between">
                <span>{c.name} <span className="text-xs text-zinc-500">(xp {c.xp}, gold {c.gold})</span></span>
                <Link className="text-indigo-600 text-sm" to={`/child/${c.id}`}>View</Link>
              </li>
            ))}
          </ul>
          <form className="mt-3 space-y-2" onSubmit={e=>{e.preventDefault(); createChild({ variables: { parentId, name: childName }})}}>
            <Label className="mb-1">New child name</Label>
            <Input value={childName} onChange={e=>setChildName(e.target.value)} placeholder="e.g. Alex"/>
            <Button>Add Child</Button>
          </form>
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
                      onChange={async e => {
                        const target = e.target
                        const childId = target.value
                        if (childId) {
                          await assignQuest({ variables: { questId: q.id, childId } })
                          target.selectedIndex = 0
                        }
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
          <form className="mt-3 space-y-2" onSubmit={e=>{e.preventDefault(); createQuest({ variables: { parentId, title: questTitle, description: questDesc || null, xp: questXP, gold: questGold }})}}>
            <Label className="mb-1">Quest title</Label>
            <Input value={questTitle} onChange={e=>setQuestTitle(e.target.value)} placeholder="e.g. Clean Room"/>
            <Label className="mb-1">Description (optional)</Label>
            <Textarea value={questDesc} onChange={e=>setQuestDesc(e.target.value)} placeholder="e.g. Tidy up and vacuum" rows={3}/>
            <div className="flex gap-2">
              <Input type="number" value={questXP} onChange={e=>setQuestXP(parseInt(e.target.value||'0'))} placeholder="XP"/>
              <Input type="number" value={questGold} onChange={e=>setQuestGold(parseInt(e.target.value||'0'))} placeholder="Gold"/>
            </div>
            <Button>Add Quest</Button>
          </form>
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
          <form className="mt-3 space-y-2" onSubmit={e=>{e.preventDefault(); createReward({ variables: { parentId, name: rewardName, xpThreshold: rewardXP }})}}>
            <Label className="mb-1">Reward name</Label>
            <Input value={rewardName} onChange={e=>setRewardName(e.target.value)} placeholder="e.g. Movie Night"/>
            <Label className="mb-1">XP threshold</Label>
            <Input type="number" value={rewardXP} onChange={e=>setRewardXP(parseInt(e.target.value||'0'))} placeholder="XP threshold"/>
            <Button>Add Reward</Button>
          </form>
          </CardContent>
        </Card>
      </section>
      </div>
    </div>
  )
}
