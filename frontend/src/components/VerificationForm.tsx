import { useForm } from '@tanstack/react-form'
import { useState } from 'react'

import { ResultChecklist } from '@/components/ResultChecklist'
import { ImageUploadField } from '@/components/ImageUploadField'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { api } from '@/lib/api'
import { applicationSchema, defaultApplication } from '@/lib/schemas'
import type { VerificationResult } from '@/lib/types'

type FieldName = keyof typeof defaultApplication

const fields: Array<{ name: FieldName; label: string; multiline?: boolean }> = [
  { name: 'brand_name', label: 'Brand name' },
  { name: 'class_type', label: 'Class/type' },
  { name: 'alcohol_content', label: 'Alcohol content' },
  { name: 'net_contents', label: 'Net contents' },
  { name: 'producer_address', label: 'Producer address' },
  { name: 'country_of_origin', label: 'Country of origin' },
  { name: 'government_warning', label: 'Government warning', multiline: true },
]

export function VerificationForm() {
  const [image, setImage] = useState<File | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<VerificationResult | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const form = useForm({
    defaultValues: defaultApplication,
    onSubmit: async ({ value }) => {
      setError(null)
      if (!image) {
        setError('Please choose a label image before submitting.')
        return
      }

      const parsed = applicationSchema.safeParse(value)
      if (!parsed.success) {
        setError(parsed.error.issues[0]?.message ?? 'Please fix the form errors.')
        return
      }

      setSubmitting(true)
      try {
        const verification = await api.createVerification(parsed.data, image)
        setResult(verification)
      } catch (submitError) {
        setError(submitError instanceof Error ? submitError.message : 'Verification failed')
      } finally {
        setSubmitting(false)
      }
    },
  })

  return (
    <div className="space-y-6">
      <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Alcohol Data</CardTitle>
          </CardHeader>
          <CardContent>
            <form
              className="space-y-4"
              onSubmit={(event) => {
                event.preventDefault()
                void form.handleSubmit()
              }}
            >
              {fields.map((field) => (
                <form.Field
                  key={field.name}
                  name={field.name}
                  validators={{
                    onChange: ({ value }) => {
                      const parsed = applicationSchema.shape[field.name].safeParse(value)
                      return parsed.success ? undefined : parsed.error.issues[0]?.message
                    },
                  }}
                >
                  {(fieldApi) => (
                    <div className="space-y-2">
                      <Label htmlFor={field.name}>{field.label}</Label>
                      {field.multiline ? (
                        <Textarea
                          id={field.name}
                          value={fieldApi.state.value}
                          onBlur={fieldApi.handleBlur}
                          onChange={(event) => fieldApi.handleChange(event.target.value)}
                          rows={4}
                        />
                      ) : (
                        <Input
                          id={field.name}
                          value={fieldApi.state.value}
                          onBlur={fieldApi.handleBlur}
                          onChange={(event) => fieldApi.handleChange(event.target.value)}
                        />
                      )}
                      {fieldApi.state.meta.errors[0] ? (
                        <p className="text-sm text-destructive">{fieldApi.state.meta.errors[0]}</p>
                      ) : null}
                    </div>
                  )}
                </form.Field>
              ))}

              <ImageUploadField
                id="label-image"
                label="Label image"
                file={image}
                onFileChange={setImage}
                hint="Upload a photo of the label. The area highlights on hover and when you drag a file over it."
              />

              {error ? (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              ) : null}

              <Button disabled={submitting} type="submit">
                {submitting ? 'Verifying…' : 'Verify label'}
              </Button>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Results</CardTitle>
          </CardHeader>
          <CardContent>
            {result ? (
              <ResultChecklist result={result} />
            ) : (
              <p className="text-sm text-muted-foreground">
                Upload a label image and submit the application fields to see a field-by-field
                checklist here.
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
