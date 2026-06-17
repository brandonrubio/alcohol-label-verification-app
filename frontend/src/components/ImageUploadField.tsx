import { ImageIcon, Upload } from 'lucide-react'
import { useEffect, useId, useRef, useState } from 'react'

import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'

type BaseProps = {
  id?: string
  label: string
  accept?: string
  hint?: string
}

type SingleImageUploadFieldProps = BaseProps & {
  multiple?: false
  file: File | null
  onFileChange: (file: File | null) => void
}

type MultipleImageUploadFieldProps = BaseProps & {
  multiple: true
  files: File[]
  onFilesChange: (files: File[]) => void
}

type ImageUploadFieldProps = SingleImageUploadFieldProps | MultipleImageUploadFieldProps

export function ImageUploadField(props: ImageUploadFieldProps) {
  const { id, label, accept = 'image/*', hint } = props
  const multiple = props.multiple === true

  const inputRef = useRef<HTMLInputElement>(null)
  const [isDragging, setIsDragging] = useState(false)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const generatedId = useId()
  const inputId = id ?? generatedId

  const selectedFiles = multiple ? props.files : props.file ? [props.file] : []
  const hasSelection = selectedFiles.length > 0
  const previewFile = multiple ? null : props.file

  useEffect(() => {
    if (!previewFile) {
      setPreviewUrl(null)
      return
    }

    const url = URL.createObjectURL(previewFile)
    setPreviewUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [previewFile])

  const applySelection = (nextFiles: File[]) => {
    if (multiple) {
      props.onFilesChange(nextFiles)
      return
    }

    props.onFileChange(nextFiles[0] ?? null)
  }

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    applySelection(Array.from(event.target.files ?? []))
    event.target.value = ''
  }

  const handleDragOver = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    event.stopPropagation()
    setIsDragging(true)
  }

  const handleDragLeave = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    event.stopPropagation()
    setIsDragging(false)
  }

  const handleDrop = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    event.stopPropagation()
    setIsDragging(false)

    const dropped = Array.from(event.dataTransfer.files ?? []).filter((item) =>
      item.type.startsWith('image/'),
    )
    applySelection(multiple ? dropped : dropped.slice(0, 1))
  }

  const openPicker = () => inputRef.current?.click()

  return (
    <div className="space-y-2">
      <Label htmlFor={inputId}>{label}</Label>
      <div
        role="button"
        tabIndex={0}
        aria-label={label}
        onClick={openPicker}
        onKeyDown={(event) => {
          if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault()
            openPicker()
          }
        }}
        onDragEnter={handleDragOver}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        className={cn(
          'flex min-h-36 cursor-pointer flex-col items-center justify-center gap-3 rounded-lg border-2 border-dashed px-4 py-6 text-center transition-all duration-150',
          'hover:border-primary hover:bg-accent/60 hover:shadow-sm',
          'focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none',
          isDragging && 'scale-[1.01] border-primary bg-accent shadow-sm',
          hasSelection && 'border-solid border-primary/50 bg-accent/40',
        )}
      >
        <input
          ref={inputRef}
          id={inputId}
          type="file"
          accept={accept}
          multiple={multiple}
          className="sr-only"
          onChange={handleInputChange}
        />

        {previewUrl ? (
          <img
            src={previewUrl}
            alt="Selected label preview"
            className="max-h-28 max-w-full rounded-md border border-border object-contain"
          />
        ) : (
          <div
            className={cn(
              'flex size-12 items-center justify-center rounded-full bg-muted transition-colors',
              (isDragging || hasSelection) && 'bg-primary/20',
            )}
          >
            {hasSelection ? (
              <ImageIcon className="size-5 text-foreground" />
            ) : (
              <Upload
                className={cn(
                  'size-5 text-muted-foreground transition-colors',
                  isDragging && 'text-foreground',
                )}
              />
            )}
          </div>
        )}

        <div className="space-y-1">
          <p className="text-sm font-medium text-foreground">
            {hasSelection
              ? multiple
                ? `${selectedFiles.length} image(s) selected`
                : selectedFiles[0]?.name
              : 'Click to upload or drag a label image here'}
          </p>
          <p className="text-xs text-muted-foreground">
            {hint ?? 'PNG, JPG, or WEBP. Hover or drop to add a file.'}
          </p>
        </div>
      </div>
    </div>
  )
}
