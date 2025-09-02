UI Kit Overview

Location: web/src/components/ui

Components

- Button: styled button with variants
  - Import: import Button from './Button'
  - Props: variant ('primary'|'secondary'|'danger'|'ghost'), size ('sm'|'md'), plus native button props
  - Example: <Button onClick={...}>Save</Button>

- Input: styled text input
  - Import: import Input from './Input'
  - Props: invalid?: boolean, plus native input props
  - Example: <Input placeholder="Your name" />

- Textarea: multi-line input
  - Import: import Textarea from './Textarea'
  - Props: invalid?: boolean, plus native textarea props
  - Example: <Textarea rows={3} placeholder="Notes" />

- Select: styled select dropdown
  - Import: import Select from './Select'
  - Props: invalid?: boolean, plus native select props
  - Example:
    <Select value={role} onChange={e=>setRole(e.target.value as any)}>
      <option value="PARENT">Parent</option>
      <option value="CHILD">Child</option>
    </Select>

- Label: form label
  - Import: import Label from './Label'
  - Example: <Label className="mb-1">Email</Label>

- Card: content container with subtle shadow
  - Import: import { Card, CardHeader, CardTitle, CardContent } from './Card'
  - Example:
    <Card>
      <CardHeader><CardTitle>Section</CardTitle></CardHeader>
      <CardContent>Body</CardContent>
    </Card>

- Badge: small status pill
  - Import: import Badge from './Badge'
  - Props: tone ('neutral'|'success'|'warning'|'danger'|'info')
  - Example: <Badge tone="success">Active</Badge>

- Tabs: simple tabbed interface
  - Import: import { Tabs, TabList, Tab, TabPanels, TabPanel } from './Tabs'
  - Example:
    <Tabs defaultIndex={0}>
      <TabList>
        <Tab idx={0}>First</Tab>
        <Tab idx={1}>Second</Tab>
      </TabList>
      <TabPanels>
        <TabPanel idx={0}>First content</TabPanel>
        <TabPanel idx={1}>Second content</TabPanel>
      </TabPanels>
    </Tabs>

Conventions

- Tailwind utility classes are baked into each component for consistency.
- Keep components small and composable; pass native props through to underlying elements.
- Prefer Label + Input/Select/Textarea for accessible forms.

Used In Pages

- Login: Button, Input, Label, Select, Card
- ParentDashboard: Button, Input, Textarea, Select, Badge, Card
- ChildView: Button, Card
