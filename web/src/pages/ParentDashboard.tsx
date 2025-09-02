import { useMemo, useState } from 'react'
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
const Q_SUB = gql`query($parentId: ID!){ subscriptionStatus(parentId:$parentId){ active currentPeriodEnd } }`
const M_CHECKOUT = gql`mutation($parentId: ID!, $success: String!, $cancel: String!){ createCheckoutSession(parentId:$parentId, successUrl:$success, cancelUrl:$cancel) }`

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

  const { data: dc, refetch: refetchChildren } = useQuery(Q_CHILDREN, { variables: { parentId } })
  const { data: dq, refetch: refetchQuests } = useQuery(Q_QUESTS, { variables: { parentId } })
  const { data: dr, refetch: refetchRewards } = useQuery(Q_REWARDS, { variables: { parentId } })
  const { data: ds } = useQuery(Q_SUB, { variables: { parentId } })

  const [createChild] = useMutation(M_CREATE_CHILD, { onCompleted: () => { setChildName(''); refetchChildren() } })
  const [createQuest] = useMutation(M_CREATE_QUEST, { onCompleted: () => { setQuestTitle(''); setQuestDesc(''); refetchQuests() } })
  const [createReward] = useMutation(M_CREATE_REWARD, { onCompleted: () => { refetchRewards() } })
  const [assignQuest] = useMutation(M_ASSIGN, { onCompleted: () => {} })
  const [checkout] = useMutation(M_CHECKOUT)

  const children = useMemo(() => (dc as any)?.children ?? [], [dc])
  const quests = useMemo(() => (dq as any)?.quests ?? [], [dq])
  const rewards = useMemo(() => (dr as any)?.rewards ?? [], [dr])
  const sub = (ds as any)?.subscriptionStatus

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
          <ul className="mt-3 space-y-2">
            {quests.map((q:any) => (
              <li key={q.id} className="flex items-center justify-between">
                <span>{q.title} <span className="text-xs text-zinc-500">(xp {q.xp}, gold {q.gold})</span></span>
                <Select onChange={e=>{ const childId=e.target.value; if(childId) assignQuest({ variables: { questId: q.id, childId }}) }}>
                  <option value="">Assign to…</option>
                  {children.map((c:any)=>(<option key={c.id} value={c.id}>{c.name}</option>))}
                </Select>
              </li>
            ))}
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
          <ul className="mt-3 space-y-2">
            {rewards.map((r:any) => (
              <li key={r.id}>{r.name} <span className="text-xs text-zinc-500">(xp {r.xpThreshold}+)</span></li>
            ))}
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
