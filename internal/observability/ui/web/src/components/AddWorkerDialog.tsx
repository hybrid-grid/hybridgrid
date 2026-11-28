import { Plus, Terminal } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { api } from '@/lib/api'
import { CodeBlock } from '@/components/CodeBlock'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { useT } from '@/lib/i18n'

/**
 * There is no "provision a worker" API — a worker joins by running the
 * hg-worker binary somewhere and pointing it at this coordinator. This
 * dialog is instructional only: the exact command to copy, not a form
 * that does anything server-side.
 */
export function AddWorkerDialog() {
  const t = useT()

  const [open, setOpen] = useState(false)
  const [selected, setSelected] = useState('')
  const join = useQuery({ queryKey: ['worker-join'], queryFn: api.workerJoin, enabled: open })
  const addresses = join.data?.addresses ?? []
  const address = addresses.includes(selected) ? selected : addresses[0]
  const runCommand = address ? `hg-worker serve --coordinator='${address.replaceAll("'", "'\\''")}'` : 'hg-worker serve'

  const composeCommand = 'docker compose up -d --scale worker=3'

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="primary" size="sm" className="gap-1.5">
          <Plus className="h-3.5 w-3.5" />
          {t('workers.addWorker')}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Terminal className="h-4 w-4 text-(--color-signal)" />
            {t('workers.dialogTitle')}
          </DialogTitle>
          <DialogDescription>{t('workers.dialogDescription')}</DialogDescription>
        </DialogHeader>

        <DialogBody className="flex flex-col gap-4">
          <Step title={t('workers.step1Title')} body={t('workers.step1Body')} />

          <Step title={t('workers.step2Title')} body={t('workers.step2Body')}>
            {addresses.length > 1 && (
              <label className="mb-2 flex flex-wrap items-center gap-2 text-[12px]">
                {t('workers.coordinatorAddress')}
                <select value={address} onChange={(event) => setSelected(event.target.value)}
                  className="rounded border border-(--color-hairline) bg-(--color-surface) px-2 py-1 font-mono">
                  {addresses.map((value) => <option key={value}>{value}</option>)}
                </select>
              </label>
            )}
            {join.isPending ? <p>{t('common.loading')}</p> : (
              <>
                {!address && <p className="text-[12px]">{t('workers.joinDiscovery')}</p>}
                <CodeBlock code={runCommand} />
              </>
            )}
          </Step>

          <Step title={t('workers.step3Title')} body={t('workers.step3Body')}>
            <CodeBlock code={composeCommand} />
          </Step>

          <Step title={t('workers.step4Title')} body={t('workers.step4Body')} />
        </DialogBody>
      </DialogContent>
    </Dialog>
  )
}

function Step({ title, body, children }: { title: string; body: string; children?: React.ReactNode }) {
  return (
    <div>
      <div className="text-[12px] font-semibold text-(--color-ink)">{title}</div>
      <p className="mt-0.5 text-[12px] text-(--color-ink-muted)">{body}</p>
      {children && <div className="mt-2">{children}</div>}
    </div>
  )
}
