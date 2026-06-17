import type { FieldStatus, OverallStatus } from '@/lib/types'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

const overallLabels: Record<OverallStatus, string> = {
  pass: 'Pass',
  needs_review: 'Needs review',
  fail: 'Fail',
}

const fieldLabels: Record<FieldStatus, string> = {
  match: 'Match',
  needs_review: 'Needs review',
  mismatch: 'Mismatch',
}

export function StatusBadge({
  status,
  kind = 'overall',
  className,
}: {
  status: OverallStatus | FieldStatus
  kind?: 'overall' | 'field'
  className?: string
}) {
  const label =
    kind === 'overall'
      ? overallLabels[status as OverallStatus]
      : fieldLabels[status as FieldStatus]

  return (
    <Badge
      className={cn(
        'font-medium',
        status === 'pass' || status === 'match'
          ? 'bg-primary/20 text-foreground hover:bg-primary/20'
          : status === 'needs_review'
            ? 'bg-amber-100 text-amber-950 hover:bg-amber-100'
            : 'bg-destructive/15 text-destructive hover:bg-destructive/15',
        className,
      )}
      variant="secondary"
    >
      {label}
    </Badge>
  )
}
