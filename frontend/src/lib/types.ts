export type FieldStatus = 'match' | 'mismatch' | 'needs_review'

export type OverallStatus = 'pass' | 'needs_review' | 'fail'

export interface FieldResult {
  field: string
  status: FieldStatus
  expected: string
  found: string
  evidence?: string
  message?: string
}

export interface ExtractedFields {
  brand_name: string
  class_type: string
  alcohol_content: string
  net_contents: string
  producer_address?: string
  country_of_origin?: string
  government_warning: string
  evidence?: Record<string, string>
  confidence?: Record<string, number>
}

export interface ApplicationData {
  brand_name: string
  class_type: string
  alcohol_content: string
  net_contents: string
  producer_address?: string
  country_of_origin?: string
  government_warning: string
}

export interface VerificationResult {
  id: string
  user_id?: string
  status: OverallStatus
  image_name: string
  application: ApplicationData
  extracted: ExtractedFields
  fields: FieldResult[]
  processing_ms: number
  created_at: string
}

export interface UserProfile {
  id: string
  email?: string
  name?: string
}
