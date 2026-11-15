import { Plus, Terminal } from 'lucide-react'
import { useMemo } from 'react'
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

  const coordinatorAddress = useMemo(() => `${window.location.hostname}:9000`, [])

  const runCommand = `hg-worker serve \\
  --coordinator=${coordinatorAddress} \\
  --advertise-address=<this-machine-ip>:50052`

  const composeCommand = 'docker compose up -d --scale worker=3'

  return (
    <Dialog>
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
            <CodeBlock code={runCommand} />
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
