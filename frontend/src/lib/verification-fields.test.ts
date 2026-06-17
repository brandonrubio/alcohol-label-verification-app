import { describe, expect, it } from 'vitest'

import { getComparisonFieldValue, sortFieldResults } from '@/lib/verification-fields'
import type { FieldResult } from '@/lib/types'

describe('verification-fields', () => {
  it('reads values from the application object by field key', () => {
    expect(
      getComparisonFieldValue(
        {
          brand_name: 'RINGSIDE',
          class_type: 'Kentucky Straight Bourbon Whiskey',
          alcohol_content: '45% ABV',
          net_contents: '750 mL',
          government_warning: 'GOVERNMENT WARNING:',
        },
        'brand_name',
      ),
    ).toBe('RINGSIDE')
  })

  it('sorts failures first and keeps a stable field order within each status', () => {
    const fields: FieldResult[] = [
      {
        field: 'brand_name',
        status: 'match',
        expected: 'RINGSIDE',
        found: 'RINGSIDE',
      },
      {
        field: 'country_of_origin',
        status: 'mismatch',
        expected: 'United States',
        found: 'USA',
      },
      {
        field: 'government_warning',
        status: 'mismatch',
        expected: 'GOVERNMENT WARNING:',
        found: '',
      },
    ]

    expect(sortFieldResults(fields).map((field) => field.field)).toEqual([
      'country_of_origin',
      'government_warning',
      'brand_name',
    ])
  })
})
