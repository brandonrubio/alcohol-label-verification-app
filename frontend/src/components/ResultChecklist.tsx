import type { VerificationResult } from '@/lib/types'
import { StatusBadge } from '@/components/StatusBadge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getComparisonFieldValue, sortFieldResults } from '@/lib/verification-fields'

const fieldLabels: Record<string, string> = {
  brand_name: 'Brand name',
  class_type: 'Class/type',
  alcohol_content: 'Alcohol content',
  net_contents: 'Net contents',
  producer_address: 'Producer address',
  country_of_origin: 'Country of origin',
  government_warning: 'Government warning',
}

function formatCellValue(value: string) {
  if (value.length <= 120) {
    return value
  }
  return `${value.slice(0, 117)}…`
}

export function ResultChecklist({ result }: { result: VerificationResult }) {
  const fields = sortFieldResults(result.fields)

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <StatusBadge status={result.status} />
        <p className="text-sm text-muted-foreground">
          {result.image_name} · processed in {result.processing_ms} ms
        </p>
      </div>

      {result.status === 'fail' ? (
        <Alert variant="destructive">
          <AlertTitle>Review required</AlertTitle>
          <AlertDescription>
            One or more fields did not match the alcohol data you submitted. Failures are listed
            first.
          </AlertDescription>
        </Alert>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Field checklist</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[9rem]">Field</TableHead>
                <TableHead className="w-[7.5rem]">Status</TableHead>
                <TableHead className="min-w-[10rem]">Alcohol data</TableHead>
                <TableHead className="min-w-[10rem]">On label</TableHead>
                <TableHead className="min-w-[12rem]">Notes</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {fields.map((field) => {
                const alcoholData = getComparisonFieldValue(result.application, field.field)
                const onLabel = getComparisonFieldValue(result.extracted, field.field)

                return (
                  <TableRow key={field.field}>
                    <TableCell className="align-top font-medium">
                      {fieldLabels[field.field] ?? field.field}
                    </TableCell>
                    <TableCell className="align-top">
                      <StatusBadge status={field.status} kind="field" />
                    </TableCell>
                    <TableCell
                      className="max-w-xs align-top text-sm whitespace-normal"
                      title={alcoholData}
                    >
                      {formatCellValue(alcoholData)}
                    </TableCell>
                    <TableCell
                      className="max-w-xs align-top text-sm whitespace-normal"
                      title={onLabel}
                    >
                      {formatCellValue(onLabel)}
                    </TableCell>
                    <TableCell className="max-w-sm align-top text-sm whitespace-normal text-muted-foreground">
                      {field.message}
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
