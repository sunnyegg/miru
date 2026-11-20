import {useEffect, useState} from 'react'
import {
  advancedFilterCount,
  advancedFiltersValid,
  defaultAdvancedFilters,
  formatLabels,
  sortLabels,
  statusLabels,
  type AnimeAdvancedFilters,
} from '../lib/animeFilters'
import {mediaSeasons, seasonDisplayName} from '../lib/season'
import {useSearchStore} from '../stores/searchStore'
import {Button} from '@/components/ui/button'
import {Dialog} from '@/components/ui/dialog'
import {Input} from '@/components/ui/input'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'

type Props = {
  open: boolean
  onClose: () => void
}

type ChipField = 'genres' | 'excludeGenres' | 'formats' | 'statuses'

type ChipGroupProps = {
  label: string
  field: ChipField
  options: string[]
  selected: string[]
  labels?: Record<string, string>
  onToggle: (field: ChipField, value: string) => void
}

function ChipGroup({
  label,
  field,
  options,
  selected,
  labels,
  onToggle,
}: ChipGroupProps) {
  return (
    <fieldset>
      <legend className="text-sm font-medium">{label}</legend>
      <div className="mt-2 flex flex-wrap gap-2">
        {options.map((option) => {
          const active = selected.includes(option)
          return (
            <Button
              key={option}
              type="button"
              variant={active ? 'default' : 'muted'}
              aria-pressed={active}
              onClick={() => onToggle(field, option)}
            >
              {labels?.[option] ?? option}
            </Button>
          )
        })}
      </div>
    </fieldset>
  )
}

export function AnimeAdvancedSearchSheet({open, onClose}: Props) {
  const filters = useSearchStore((state) => state.advancedFilters)
  const filterOptions = useSearchStore((state) => state.filterOptions)
  const searching = useSearchStore((state) => state.animeLoading)
  const loadFilterOptions = useSearchStore((state) => state.loadFilterOptions)
  const setAdvancedFilters = useSearchStore((state) => state.setAdvancedFilters)
  const searchAnimeAdvanced = useSearchStore(
    (state) => state.searchAnimeAdvanced,
  )
  const [draft, setDraft] = useState<AnimeAdvancedFilters>(filters)

  useEffect(() => {
    if (!open) {
      return
    }
    setDraft(filters)
    void loadFilterOptions()
  }, [open, filters, loadFilterOptions])

  const currentYear = new Date().getFullYear()
  const years: number[] = []
  for (let year = currentYear + 2; year >= 1940; year -= 1) {
    years.push(year)
  }
  if (draft.year > 0 && !years.includes(draft.year)) {
    years.unshift(draft.year)
  }

  const genres = filterOptions?.genres ?? []
  const formats = filterOptions?.formats ?? []
  const statuses = filterOptions?.statuses ?? []
  const sorts = filterOptions?.sorts ?? Object.keys(sortLabels)
  const seasonNeedsYear = draft.season !== '' && draft.year === 0
  const isEmpty = advancedFilterCount(draft) === 0
  const canApply = advancedFiltersValid(draft) && !searching

  function toggleChip(field: ChipField, value: string) {
    setDraft((current) => {
      const selected = current[field]
      const next = selected.includes(value)
        ? selected.filter((item) => item !== value)
        : [...selected, value]
      return {...current, [field]: next}
    })
  }

  function apply() {
    if (!advancedFiltersValid(draft)) {
      return
    }
    setAdvancedFilters({...draft, search: draft.search.trim()})
    void searchAnimeAdvanced()
    onClose()
  }

  return (
    <Dialog.Root
      open={open}
      onOpenChange={(next) => {
        if (!next) {
          onClose()
        }
      }}
    >
      <Dialog.Portal>
        <Dialog.Backdrop />
        <Dialog.Viewport>
          <Dialog.Panel className="max-w-2xl">
            <form
              className="flex flex-col gap-4"
              onSubmit={(event) => {
                event.preventDefault()
                apply()
              }}
            >
              <div>
                <Dialog.Title>Advanced anime search</Dialog.Title>
                <Dialog.Description>
                  Filters combine on AniList. Pick at least one, and a year
                  whenever you pick a season.
                </Dialog.Description>
              </div>

              <div>
                <label className="text-sm font-medium" htmlFor="advanced-title">
                  Title
                </label>
                <Input
                  id="advanced-title"
                  className="mt-2"
                  value={draft.search}
                  maxLength={200}
                  placeholder="Optional title, for example Frieren"
                  onChange={(event) =>
                    setDraft((current) => ({
                      ...current,
                      search: event.target.value,
                    }))
                  }
                />
              </div>

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <label
                    className="text-sm font-medium"
                    htmlFor="advanced-year"
                  >
                    Year
                  </label>
                  <NativeSelect
                    id="advanced-year"
                    className="mt-2 w-full"
                    value={String(draft.year)}
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        year: Number(event.target.value),
                      }))
                    }
                  >
                    <NativeSelectOption value="0">Any year</NativeSelectOption>
                    {years.map((year) => (
                      <NativeSelectOption key={year} value={String(year)}>
                        {year}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>
                </div>
                <div>
                  <label
                    className="text-sm font-medium"
                    htmlFor="advanced-season"
                  >
                    Season
                  </label>
                  <NativeSelect
                    id="advanced-season"
                    className="mt-2 w-full"
                    value={draft.season}
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        season: event.target.value,
                      }))
                    }
                  >
                    <NativeSelectOption value="">Any season</NativeSelectOption>
                    {mediaSeasons.map((season) => (
                      <NativeSelectOption key={season} value={season}>
                        {seasonDisplayName(season)}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>
                </div>
                <div>
                  <label
                    className="text-sm font-medium"
                    htmlFor="advanced-score"
                  >
                    Minimum score
                  </label>
                  <Input
                    id="advanced-score"
                    className="mt-2"
                    type="number"
                    min={0}
                    max={100}
                    step={5}
                    value={draft.minScore > 0 ? String(draft.minScore) : ''}
                    placeholder="Any score"
                    onChange={(event) => {
                      const parsed = Number(event.target.value)
                      const minScore = Number.isFinite(parsed)
                        ? Math.min(100, Math.max(0, Math.round(parsed)))
                        : 0
                      setDraft((current) => ({...current, minScore}))
                    }}
                  />
                </div>
                <div>
                  <label
                    className="text-sm font-medium"
                    htmlFor="advanced-sort"
                  >
                    Sort by
                  </label>
                  <NativeSelect
                    id="advanced-sort"
                    className="mt-2 w-full"
                    value={draft.sort}
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        sort: event.target.value,
                      }))
                    }
                  >
                    {sorts.map((sort) => (
                      <NativeSelectOption key={sort} value={sort}>
                        {sortLabels[sort] ?? sort}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>
                </div>
              </div>

              <ChipGroup
                label="Genres"
                field="genres"
                options={genres}
                selected={draft.genres}
                onToggle={toggleChip}
              />
              <ChipGroup
                label="Exclude genres"
                field="excludeGenres"
                options={genres}
                selected={draft.excludeGenres}
                onToggle={toggleChip}
              />
              <ChipGroup
                label="Format"
                field="formats"
                options={formats}
                selected={draft.formats}
                labels={formatLabels}
                onToggle={toggleChip}
              />
              <ChipGroup
                label="Status"
                field="statuses"
                options={statuses}
                selected={draft.statuses}
                labels={statusLabels}
                onToggle={toggleChip}
              />

              {(isEmpty || seasonNeedsYear) && (
                <p className="text-sm text-destructive">
                  {seasonNeedsYear
                    ? 'Pick a year to use a season filter.'
                    : 'Set at least one filter to search.'}
                </p>
              )}

              <div className="flex flex-wrap items-center justify-between gap-2">
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() => setDraft(defaultAdvancedFilters)}
                >
                  Reset
                </Button>
                <div className="flex flex-wrap gap-2">
                  <Button type="button" variant="muted" onClick={onClose}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={!canApply}>
                    {searching ? 'Searching…' : 'Apply filters'}
                  </Button>
                </div>
              </div>
            </form>
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
