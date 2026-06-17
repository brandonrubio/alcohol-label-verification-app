import { describe, expect, it } from 'vitest'

import { applicationSchema } from '@/lib/schemas'

describe('application schema', () => {
  it('requires core label fields', () => {
    const parsed = applicationSchema.safeParse({
      brand_name: 'OLD TOM DISTILLERY',
      class_type: 'Kentucky Straight Bourbon Whiskey',
      alcohol_content: '45% Alc./Vol.',
      net_contents: '750 mL',
      government_warning: 'GOVERNMENT WARNING:',
    })

    expect(parsed.success).toBe(true)
  })
})
