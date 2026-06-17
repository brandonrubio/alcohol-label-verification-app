import { z } from 'zod'

export const applicationSchema = z.object({
  brand_name: z.string().min(1, 'Brand name is required'),
  class_type: z.string().min(1, 'Class/type is required'),
  alcohol_content: z.string().min(1, 'Alcohol content is required'),
  net_contents: z.string().min(1, 'Net contents are required'),
  producer_address: z.string().optional(),
  country_of_origin: z.string().optional(),
  government_warning: z.string().min(1, 'Government warning text is required'),
})

export type ApplicationFormValues = z.infer<typeof applicationSchema>

export const defaultApplication: ApplicationFormValues = {
  brand_name: 'OLD TOM DISTILLERY',
  class_type: 'Kentucky Straight Bourbon Whiskey',
  alcohol_content: '45% Alc./Vol. (90 Proof)',
  net_contents: '750 mL',
  producer_address: 'Old Tom Distillery, Louisville, KY',
  country_of_origin: 'United States',
  government_warning:
    'GOVERNMENT WARNING: (1) According to the Surgeon General, women should not drink alcoholic beverages during pregnancy because of the risk of birth defects. (2) Consumption of alcoholic beverages impairs your ability to drive a car or operate machinery, and may cause health problems.',
}
