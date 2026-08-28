import {useEffect, useState} from 'react'
import type {AnimeListEntryInput, WatchingEntryView} from '../lib/types'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {Label} from '@/components/ui/label'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'
import {Textarea} from '@/components/ui/textarea'

type Props = {
  entry: WatchingEntryView
  saving: boolean
  onClose: () => void
  onSave: (input: AnimeListEntryInput) => void
}

const listStatusOptions = [
  {value: 'CURRENT', label: 'Watching'},
  {value: 'COMPLETED', label: 'Completed'},
  {value: 'PLANNING', label: 'Planning'},
  {value: 'PAUSED', label: 'Paused'},
  {value: 'DROPPED', label: 'Dropped'},
  {value: 'REPEATING', label: 'Repeating'},
]

function fuzzyDateToInput(year: number, month: number, day: number): string {
  if (year <= 0 || month <= 0 || day <= 0) {
    return ''
  }
  const paddedMonth = String(month).padStart(2, '0')
  const paddedDay = String(day).padStart(2, '0')
  return `${year}-${paddedMonth}-${paddedDay}`
}

function inputDateToParts(value: string): {year: number; month: number; day: number} {
  if (!value) {
    return {year: 0, month: 0, day: 0}
  }
  const [yearText, monthText, dayText] = value.split('-')
  return {
    year: Number(yearText) || 0,
    month: Number(monthText) || 0,
    day: Number(dayText) || 0,
  }
}

function entryToForm(entry: WatchingEntryView) {
  return {
    status: entry.listStatus || 'CURRENT',
    progress: String(entry.progress),
    scoreRaw: String(entry.scoreRaw),
    notes: entry.notes,
    repeat: String(entry.repeat),
    privateEntry: entry.private ? 'true' : 'false',
    startedAt: fuzzyDateToInput(entry.startedAt.year, entry.startedAt.month, entry.startedAt.day),
    completedAt: fuzzyDateToInput(entry.completedAt.year, entry.completedAt.month, entry.completedAt.day),
  }
}

export function WatchingEditSheet({entry, saving, onClose, onSave}: Props) {
  const title = entry.titleEnglish || entry.titleRomaji
  const [form, setForm] = useState(() => entryToForm(entry))

  useEffect(() => {
    setForm(entryToForm(entry))
  }, [entry])

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        onClose()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [onClose])

  function submitSave() {
    const startedAt = inputDateToParts(form.startedAt)
    const completedAt = inputDateToParts(form.completedAt)
    onSave({
      mediaId: entry.mediaId,
      status: form.status,
      progress: Number(form.progress) || 0,
      scoreRaw: Number(form.scoreRaw) || 0,
      notes: form.notes,
      repeat: Number(form.repeat) || 0,
      private: form.privateEntry === 'true',
      startedYear: startedAt.year,
      startedMonth: startedAt.month,
      startedDay: startedAt.day,
      completedYear: completedAt.year,
      completedMonth: completedAt.month,
      completedDay: completedAt.day,
    })
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-bezel/80 p-6"
      role="presentation"
      onClick={onClose}
    >
      <Card
        className="max-h-[calc(100vh-3rem)] w-full max-w-lg overflow-y-auto border border-border/40 p-4"
        role="dialog"
        aria-labelledby="watching-edit-title"
        aria-modal="true"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <h3 id="watching-edit-title" className="text-base font-medium">
              Edit list entry
            </h3>
            <p className="mt-1 truncate text-sm text-muted-foreground">{title}</p>
          </div>
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancel
          </Button>
        </div>

        <div className="mt-4 grid gap-4">
          <div className="grid gap-2">
            <Label htmlFor="watching-edit-status">Status</Label>
            <NativeSelect
              id="watching-edit-status"
              value={form.status}
              onChange={(event) => setForm((current) => ({...current, status: event.target.value}))}
            >
              {listStatusOptions.map((option) => (
                <NativeSelectOption key={option.value} value={option.value}>
                  {option.label}
                </NativeSelectOption>
              ))}
            </NativeSelect>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="grid gap-2">
              <Label htmlFor="watching-edit-progress">Progress (episodes)</Label>
              <Input
                id="watching-edit-progress"
                type="number"
                min={0}
                value={form.progress}
                onChange={(event) => setForm((current) => ({...current, progress: event.target.value}))}
                className="border-border/40"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="watching-edit-score">Score (0–100)</Label>
              <Input
                id="watching-edit-score"
                type="number"
                min={0}
                max={100}
                value={form.scoreRaw}
                onChange={(event) => setForm((current) => ({...current, scoreRaw: event.target.value}))}
                className="border-border/40"
              />
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="grid gap-2">
              <Label htmlFor="watching-edit-repeat">Repeat</Label>
              <Input
                id="watching-edit-repeat"
                type="number"
                min={0}
                max={1000}
                value={form.repeat}
                onChange={(event) => setForm((current) => ({...current, repeat: event.target.value}))}
                className="border-border/40"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="watching-edit-private">Visibility</Label>
              <NativeSelect
                id="watching-edit-private"
                value={form.privateEntry}
                onChange={(event) => setForm((current) => ({...current, privateEntry: event.target.value}))}
              >
                <NativeSelectOption value="false">Public</NativeSelectOption>
                <NativeSelectOption value="true">Private</NativeSelectOption>
              </NativeSelect>
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="grid gap-2">
              <Label htmlFor="watching-edit-started">Started</Label>
              <Input
                id="watching-edit-started"
                type="date"
                value={form.startedAt}
                onChange={(event) => setForm((current) => ({...current, startedAt: event.target.value}))}
                className="border-border/40"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="watching-edit-completed">Completed</Label>
              <Input
                id="watching-edit-completed"
                type="date"
                value={form.completedAt}
                onChange={(event) => setForm((current) => ({...current, completedAt: event.target.value}))}
                className="border-border/40"
              />
            </div>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="watching-edit-notes">Notes</Label>
            <Textarea
              id="watching-edit-notes"
              value={form.notes}
              rows={4}
              maxLength={6000}
              onChange={(event) => setForm((current) => ({...current, notes: event.target.value}))}
              className="border-border/40"
            />
          </div>
        </div>

        <div className="mt-4 flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={onClose} disabled={saving}>
            Cancel
          </Button>
          <Button type="button" onClick={submitSave} disabled={saving} aria-busy={saving}>
            {saving ? 'Saving…' : 'Save'}
          </Button>
        </div>
      </Card>
    </div>
  )
}
