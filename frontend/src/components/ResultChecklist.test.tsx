import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { ResultChecklist } from '@/components/ResultChecklist'
import type { VerificationResult } from '@/lib/types'

const result: VerificationResult = {
  id: '1',
  status: 'fail',
  image_name: 'label.jpg',
  application: {
    brand_name: 'OLD TOM DISTILLERY',
    class_type: 'Kentucky Straight Bourbon Whiskey',
    alcohol_content: '45% Alc./Vol.',
    net_contents: '750 mL',
    government_warning: 'GOVERNMENT WARNING:',
  },
  extracted: {
    brand_name: 'WRONG BRAND',
    class_type: 'Kentucky Straight Bourbon Whiskey',
    alcohol_content: '45% Alc./Vol.',
    net_contents: '750 mL',
    government_warning: '',
  },
  fields: [
    {
      field: 'brand_name',
      status: 'mismatch',
      expected: 'OLD TOM DISTILLERY',
      found: 'WRONG BRAND',
      message: 'Brand name on the label does not match the application.',
    },
    {
      field: 'alcohol_content',
      status: 'match',
      expected: '45% Alc./Vol.',
      found: '45% Alc./Vol.',
      message: 'Alcohol content matches.',
    },
  ],
  processing_ms: 900,
  created_at: new Date().toISOString(),
}

describe('ResultChecklist', () => {
  it('shows failures before matches', () => {
    render(<ResultChecklist result={result} />)
    const rows = screen.getAllByRole('row')
    expect(rows[1]).toHaveTextContent('Brand name')
    expect(screen.getByText('Fail')).toBeInTheDocument()
  })
})
