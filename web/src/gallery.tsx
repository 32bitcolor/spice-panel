import * as React from 'react'
import { createRoot } from 'react-dom/client'
import './ui/gallery.css'
import { Button, Panel, Chip, Spinner, TextField, Switch, Checkbox } from './ui'

const Section: React.FC<{ title: string, children: React.ReactNode }> = ({
  title,
  children,
}): React.ReactElement => (
  <section className="mb-10">
    <div className="mb-4 flex items-center gap-3">
      <span className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">{title}</span>
      <span className="h-px flex-1 bg-gradient-to-r from-border to-transparent" />
    </div>
    {children}
  </section>
)

const Gallery: React.FC = (): React.ReactElement => {
  const [pvp, setPvp] = React.useState(true)
  const [maint, setMaint] = React.useState(false)
  const [agree, setAgree] = React.useState(true)

  return (
    <div className="mx-auto max-w-4xl px-7 py-10">
      <h1 className="mb-8 text-xl font-bold uppercase tracking-[0.34em]">
        <span className="text-accent">SPICE</span>
        -PANEL · UI
      </h1>

      <Section title="Buttons">
        <div className="flex flex-wrap items-center gap-3">
          <Button variant="primary">Start Server</Button>
          <Button variant="solid">Restart</Button>
          <Button variant="ghost">Broadcast</Button>
          <Button variant="danger">Stop</Button>
          <Button variant="primary" size="sm">
            Grant Item
          </Button>
          <Button variant="solid" size="sm" isDisabled>
            Locked
          </Button>
        </div>
      </Section>

      <Section title="Status chips">
        <div className="flex flex-wrap items-center gap-3">
          <Chip color="success" dot>
            Online
          </Chip>
          <Chip color="warning" dot>
            Restarting
          </Chip>
          <Chip color="danger" dot>
            Crashed
          </Chip>
          <Chip color="muted" dot>
            Idle
          </Chip>
          <Chip color="accent">Spice</Chip>
        </div>
      </Section>

      <Section title="Form controls">
        <Panel className="flex flex-col gap-5">
          <div className="grid gap-4 sm:grid-cols-2">
            <TextField label="Player search" placeholder="name or account id…" type="search" />
            <TextField label="Spice grant" placeholder="0" defaultValue="50000" />
          </div>
          <div className="flex flex-wrap items-center gap-8">
            <Switch isSelected={pvp} onChange={setPvp}>
              PvP enabled
            </Switch>
            <Switch isSelected={maint} onChange={setMaint}>
              Maintenance mode
            </Switch>
            <Checkbox isSelected={agree} onChange={setAgree}>
              Confirm destructive action
            </Checkbox>
          </div>
        </Panel>
      </Section>

      <Section title="Panel + loading">
        <div className="grid gap-4 sm:grid-cols-2">
          <Panel>
            <div className="mb-2 font-mono text-[11px] uppercase tracking-[0.22em] text-muted">
              Arrakis-01
            </div>
            <div className="text-3xl font-semibold text-focus">47</div>
            <div className="mt-1 text-sm text-muted">players online</div>
          </Panel>
          <Panel className="flex items-center gap-3">
            <Spinner />
            <span className="text-sm text-muted">Fetching server status…</span>
          </Panel>
        </div>
      </Section>
    </div>
  )
}

const root = document.getElementById('gallery')
if (root) createRoot(root).render(<Gallery />)
