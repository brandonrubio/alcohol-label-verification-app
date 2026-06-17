import { useForm } from '@tanstack/react-form'
import { useState } from 'react'

import { ResultChecklist } from '@/components/ResultChecklist'
import { ImageUploadField } from '@/components/ImageUploadField'
import { StatusBadge } from '@/components/StatusBadge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Progress } from '@/components/ui/progress'
import { Textarea } from '@/components/ui/textarea'
import { api } from '@/lib/api'
import { applicationSchema, defaultApplication } from '@/lib/schemas'
import type { VerificationBatch } from '@/lib/types'

export function BatchUpload() {
  const [images, setImages] = useState<File[]>([])
  const [error, setError] = useState<string | null>(null)
  const [batch, setBatch] = useState<VerificationBatch | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const form = useForm({
    defaultValues: defaultApplication,
    onSubmit: async ({ value }) => {
      setError(null)
      if (images.length === 0) {
        setError('Add at least one label image for batch processing.')
        return
      }

      const parsed = applicationSchema.safeParse(value)
      if (!parsed.success) {
        setError(parsed.error.issues[0]?.message ?? 'Please fix the form errors.')
        return
      }

      setSubmitting(true)
      try {
        const response = await api.createBatch(parsed.data, images)
        setBatch(response)
      } catch (submitError) {
        setError(submitError instanceof Error ? submitError.message : 'Batch verification failed')
      } finally {
        setSubmitting(false)
      }
    },
  })

  const progress = batch ? Math.round((batch.completed_count / batch.total_count) * 100) : 0
  const sortedResults = [...(batch?.results ?? [])].sort((a, b) => {
    const priority = { fail: 0, needs_review: 1, pass: 2 }
    return priority[a.status] - priority[b.status]
  })

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Batch upload</CardTitle>
        </CardHeader>
        <CardContent>
          <form
            className="space-y-4"
            onSubmit={(event) => {
              event.preventDefault()
              void form.handleSubmit()
            }}
          >
            <form.Field name="brand_name">
              {(fieldApi) => (
                <div className="space-y-2">
                  <Label htmlFor="batch-brand">Brand name (shared application)</Label>
                  <Input
                    id="batch-brand"
                    value={fieldApi.state.value}
                    onChange={(event) => fieldApi.handleChange(event.target.value)}
                  />
                </div>
              )}
            </form.Field>

            <form.Field name="class_type">
              {(fieldApi) => (
                <div className="space-y-2">
                  <Label htmlFor="batch-class">Class/type</Label>
                  <Input
                    id="batch-class"
                    value={fieldApi.state.value}
                    onChange={(event) => fieldApi.handleChange(event.target.value)}
                  />
                </div>
              )}
            </form.Field>

            <form.Field name="alcohol_content">
              {(fieldApi) => (
                <div className="space-y-2">
                  <Label htmlFor="batch-abv">Alcohol content</Label>
                  <Input
                    id="batch-abv"
                    value={fieldApi.state.value}
                    onChange={(event) => fieldApi.handleChange(event.target.value)}
                  />
                </div>
              )}
            </form.Field>

            <form.Field name="net_contents">
              {(fieldApi) => (
                <div className="space-y-2">
                  <Label htmlFor="batch-net">Net contents</Label>
                  <Input
                    id="batch-net"
                    value={fieldApi.state.value}
                    onChange={(event) => fieldApi.handleChange(event.target.value)}
                  />
                </div>
              )}
            </form.Field>

            <form.Field name="government_warning">
              {(fieldApi) => (
                <div className="space-y-2">
                  <Label htmlFor="batch-warning">Government warning</Label>
                  <Textarea
                    id="batch-warning"
                    rows={3}
                    value={fieldApi.state.value}
                    onChange={(event) => fieldApi.handleChange(event.target.value)}
                  />
                </div>
              )}
            </form.Field>

            <ImageUploadField
              id="batch-images"
              label="Label images"
              multiple
              files={images}
              onFilesChange={setImages}
              hint="Select or drop multiple label images. The upload area reacts on hover and drag-over."
            />

            {error ? (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}

            <Button disabled={submitting} type="submit">
              {submitting ? 'Processing batch…' : 'Run batch verification'}
            </Button>
          </form>
        </CardContent>
      </Card>

      {batch ? (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Batch status</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center justify-between text-sm">
                <span>
                  {batch.completed_count} of {batch.total_count} completed
                </span>
                <span className="text-muted-foreground">{batch.status}</span>
              </div>
              <Progress value={progress} />
            </CardContent>
          </Card>

          {sortedResults.map((result) => (
            <Card key={result.id}>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle className="text-base">{result.image_name}</CardTitle>
                <StatusBadge status={result.status} />
              </CardHeader>
              <CardContent>
                <ResultChecklist result={result} />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : null}
    </div>
  )
}
