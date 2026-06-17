import type { ExtractedFields, FieldResult } from '@/lib/types'

export const comparisonFieldOrder = [
  'brand_name',
  'class_type',
  'alcohol_content',
  'net_contents',
  'producer_address',
  'country_of_origin',
  'government_warning',
] as const

export type ComparisonField = (typeof comparisonFieldOrder)[number]

type FieldRecord = Partial<Record<ComparisonField, string | undefined>>

export function getComparisonFieldValue(
  record: FieldRecord | Record<string, unknown>,
  field: string,
): string {
  const value = record[field as ComparisonField]
  if (typeof value !== 'string') {
    return '—'
  }

  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : '—'
}

export function sortFieldResults(fields: FieldResult[]) {
  const priority = { mismatch: 0, needs_review: 1, match: 2 }

  return [...fields].sort((a, b) => {
    const statusDiff = priority[a.status] - priority[b.status]
    if (statusDiff !== 0) {
      return statusDiff
    }

    const aIndex = comparisonFieldOrder.indexOf(a.field as ComparisonField)
    const bIndex = comparisonFieldOrder.indexOf(b.field as ComparisonField)
    return (aIndex === -1 ? 99 : aIndex) - (bIndex === -1 ? 99 : bIndex)
  })
}
