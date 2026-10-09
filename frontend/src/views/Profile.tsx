import {useEffect, useState} from 'react'
import {
  GetAnilistProfile,
  ListAnimeList,
  SaveShareBanner,
  UpdateAnilistProfile,
} from '../../wailsjs/go/main/App'
import {IconShare} from '../components/Icons'
import {LibraryPosterCard} from '../components/LibraryPosterCard'
import {
  ShareBannerDialog,
  type ShareBannerSession,
} from '../components/ShareBannerDialog'
import {WatchingEditSheet} from '../components/WatchingEditSheet'
import {SettingsCheckboxRow} from '../components/settings/SettingsCheckboxRow'
import {SettingsField} from '../components/settings/SettingsField'
import {errorMessage} from '../lib/format'
import type {
  AnilistProfileInput,
  AnilistProfileView,
  AnimeListEntryInput,
  WatchingEntryView,
} from '../lib/types'
import {useNavigationStore} from '../stores/navigationStore'
import {useWatchingStore, type ListFilter} from '../stores/watchingStore'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'
import {Skeleton} from '@/components/ui/skeleton'
import {Textarea} from '@/components/ui/textarea'
import {cn} from '@/lib/utils'

type Props = {
  notice: (msg: string, isError?: boolean) => void
}

const listFilters: {value: ListFilter; label: string}[] = [
  {value: 'CURRENT', label: 'Watching'},
  {value: 'COMPLETED', label: 'Completed'},
  {value: 'PLANNING', label: 'Planning'},
  {value: 'PAUSED', label: 'Paused'},
  {value: 'DROPPED', label: 'Dropped'},
  {value: 'REPEATING', label: 'Repeating'},
]

const titleLanguageOptions: {value: string; label: string}[] = [
  {value: 'ROMAJI', label: 'Romaji'},
  {value: 'ENGLISH', label: 'English'},
  {value: 'NATIVE', label: 'Native'},
]

const scoreFormatOptions: {value: string; label: string}[] = [
  {value: 'POINT_100', label: '100 Point (55/100)'},
  {value: 'POINT_10_DECIMAL', label: '10 Point Decimal (5.5/10)'},
  {value: 'POINT_10', label: '10 Point (5/10)'},
  {value: 'POINT_5', label: '5 Star (3/5)'},
  {value: 'POINT_3', label: '3 Point Smileys'},
]

const emptyProfileForm: AnilistProfileInput = {
  about: '',
  titleLanguage: 'ROMAJI',
  displayAdultContent: false,
  scoreFormat: 'POINT_100',
}

function listStatusLabel(status: string): string {
  const match = listFilters.find((filter) => filter.value === status)
  if (match) {
    return match.label
  }
  return status
}

const mediaStatusLabels: Record<string, string> = {
  RELEASING: 'Airing',
  FINISHED: 'Finished',
  NOT_YET_RELEASED: 'Not yet aired',
  CANCELLED: 'Cancelled',
  HIATUS: 'On hiatus',
}

function mediaStatusLabel(status: string): string {
  if (!status) {
    return ''
  }
  return mediaStatusLabels[status] ?? status
}

function profileFormFromView(profile: AnilistProfileView): AnilistProfileInput {
  return {
    about: profile.about ?? '',
    titleLanguage: profile.titleLanguage || 'ROMAJI',
    displayAdultContent: profile.displayAdultContent,
    scoreFormat: profile.scoreFormat || 'POINT_100',
  }
}

function entryPosterCaption(entry: WatchingEntryView): string {
  if (entry.totalEpisodes > 0) {
    return `${entry.progress} / ${entry.totalEpisodes}`
  }
  return `${entry.progress} watched`
}

function entryPosterSubcaption(entry: WatchingEntryView): string {
  if (entry.mediaStatus) {
    return mediaStatusLabel(entry.mediaStatus)
  }
  return 'Click to edit'
}

const posterGridClassName =
  'grid grid-cols-[repeat(auto-fill,minmax(min(12rem,100%),1fr))] gap-5 p-1'

export function ProfileView({notice}: Props) {
  const goToSettings = useNavigationStore((state) => state.setTab)
  const listFilter = useWatchingStore((state) => state.listFilter)
  const entries = useWatchingStore((state) => state.entries)
  const counts = useWatchingStore((state) => state.counts)
  const countsLoading = useWatchingStore((state) => state.countsLoading)
  const countsError = useWatchingStore((state) => state.countsError)
  const loading = useWatchingStore((state) => state.loading)
  const listNotConnected = useWatchingStore((state) => state.notConnected)
  const error = useWatchingStore((state) => state.error)
  const selectFilter = useWatchingStore((state) => state.selectFilter)
  const loadList = useWatchingStore((state) => state.loadList)
  const loadCounts = useWatchingStore((state) => state.loadCounts)
  const saveEntry = useWatchingStore((state) => state.saveEntry)

  const [username, setUsername] = useState('')
  const [profileForm, setProfileForm] =
    useState<AnilistProfileInput>(emptyProfileForm)
  const [profileLoading, setProfileLoading] = useState(true)
  const [profileNotConnected, setProfileNotConnected] = useState(false)
  const [profileError, setProfileError] = useState('')
  const [savingProfile, setSavingProfile] = useState(false)

  const [editingEntry, setEditingEntry] = useState<
    (typeof entries)[number] | null
  >(null)
  const [savingEntry, setSavingEntry] = useState(false)
  const [sharingFilter, setSharingFilter] = useState<ListFilter | null>(null)
  const [shareSession, setShareSession] = useState<ShareBannerSession | null>(
    null,
  )
  const [savingBanner, setSavingBanner] = useState(false)

  const notConnected = listNotConnected || profileNotConnected

  async function loadProfile() {
    setProfileLoading(true)
    setProfileError('')
    setProfileNotConnected(false)
    try {
      const profile = await GetAnilistProfile()
      setUsername(profile.username)
      setProfileForm(profileFormFromView(profile))
    } catch (err) {
      const message = errorMessage(err)
      if (message === 'AniList not connected') {
        setProfileNotConnected(true)
        setUsername('')
        setProfileForm(emptyProfileForm)
      } else {
        setProfileError(message)
      }
    } finally {
      setProfileLoading(false)
    }
  }

  useEffect(() => {
    void loadProfile()
    void loadList()
    void loadCounts()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function saveProfile() {
    setSavingProfile(true)
    try {
      const updated = await UpdateAnilistProfile(profileForm)
      setUsername(updated.username)
      setProfileForm(profileFormFromView(updated))
      notice('Profile saved')
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setSavingProfile(false)
    }
  }

  async function saveEntryWithState(input: AnimeListEntryInput) {
    setSavingEntry(true)
    try {
      await saveEntry(input, notice)
      setEditingEntry(null)
    } catch {
      // notice handled in store
    } finally {
      setSavingEntry(false)
    }
  }

  async function shareCategory(filter: ListFilter) {
    const categoryLabel = listStatusLabel(filter)
    setSharingFilter(filter)
    try {
      let shareEntries: WatchingEntryView[]
      if (listFilter === filter) {
        shareEntries = entries
      } else {
        shareEntries = (await ListAnimeList(filter)) ?? []
      }
      if (shareEntries.length === 0) {
        notice(`Nothing to share in ${categoryLabel}`, true)
        return
      }

      setShareSession({
        username,
        categoryLabel,
        categoryCount: counts[filter] ?? shareEntries.length,
        defaultFilename: `miru-${categoryLabel.toLowerCase()}.png`,
        posters: shareEntries.map((entry) => ({
          mediaId: entry.mediaId,
          title: entry.titleEnglish || entry.titleRomaji,
          coverImage: entry.coverImage,
        })),
      })
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setSharingFilter(null)
    }
  }

  async function saveShareBanner(pngBase64: string, defaultFilename: string) {
    setSavingBanner(true)
    try {
      const savedPath = await SaveShareBanner(pngBase64, defaultFilename)
      if (savedPath) {
        notice('Banner saved')
        setShareSession(null)
      }
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setSavingBanner(false)
    }
  }

  const filterLabel = listStatusLabel(listFilter)
  const emptyCopy = `Nothing on your ${filterLabel} list. Switch lists, or search AniList to add a title.`

  return (
    <section className="flex h-full flex-col gap-6">
      <header>
        <p className="text-xs font-medium tracking-[0.16em] text-muted-foreground uppercase">
          AniList
        </p>
        <h2 className="text-3xl font-semibold tracking-tight">Profile</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Update your AniList account preferences and anime lists.
        </p>
      </header>

      {profileLoading ? (
        <div className="flex flex-col gap-4 border border-border bg-card p-5">
          <Skeleton className="h-4 w-40 animate-pulse" />
          <Skeleton className="h-24 w-full animate-pulse" />
          <Skeleton className="h-11 w-48 animate-pulse" />
        </div>
      ) : notConnected ? (
        <section className="border border-border bg-card p-5">
          <h3 className="font-medium">
            Connect AniList to manage your profile
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Sign in with AniList from Settings, then return here to edit your
            profile and anime list.
          </p>
          <Button
            type="button"
            variant="secondary"
            className="mt-4"
            onClick={() => goToSettings('settings')}
          >
            Open Settings
          </Button>
        </section>
      ) : profileError ? (
        <Alert variant="destructive">
          <AlertDescription>{profileError}</AlertDescription>
          <AlertAction>
            <Button
              type="button"
              variant="secondary"
              onClick={() => void loadProfile()}
            >
              Try again
            </Button>
          </AlertAction>
        </Alert>
      ) : (
        <form
          className="border border-border bg-card p-5"
          onSubmit={(event) => {
            event.preventDefault()
            void saveProfile()
          }}
        >
          <h3 className="font-medium">Account</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Signed in as {username || 'AniList user'}
          </p>

          <SettingsField label="About" htmlFor="profile-about">
            <Textarea
              id="profile-about"
              rows={4}
              value={profileForm.about}
              onChange={(event) =>
                setProfileForm((current) => ({
                  ...current,
                  about: event.target.value,
                }))
              }
              className="min-h-24 bg-card"
            />
          </SettingsField>

          <SettingsField
            label="Title language"
            htmlFor="profile-title-language"
          >
            <NativeSelect
              id="profile-title-language"
              value={profileForm.titleLanguage}
              onChange={(event) =>
                setProfileForm((current) => ({
                  ...current,
                  titleLanguage: event.target.value,
                }))
              }
              className="max-w-xs bg-card"
            >
              {titleLanguageOptions.map((option) => (
                <NativeSelectOption key={option.value} value={option.value}>
                  {option.label}
                </NativeSelectOption>
              ))}
            </NativeSelect>
          </SettingsField>

          <SettingsField label="Scoring system" htmlFor="profile-score-format">
            <NativeSelect
              id="profile-score-format"
              value={profileForm.scoreFormat}
              onChange={(event) =>
                setProfileForm((current) => ({
                  ...current,
                  scoreFormat: event.target.value,
                }))
              }
              className="max-w-xs bg-card"
            >
              {scoreFormatOptions.map((option) => (
                <NativeSelectOption key={option.value} value={option.value}>
                  {option.label}
                </NativeSelectOption>
              ))}
            </NativeSelect>
          </SettingsField>

          <SettingsCheckboxRow
            id="profile-display-adult"
            label="Display adult content"
            checked={profileForm.displayAdultContent}
            onChange={(checked) =>
              setProfileForm((current) => ({
                ...current,
                displayAdultContent: checked,
              }))
            }
            className="mt-4"
          />

          <Button
            type="submit"
            variant="secondary"
            className="mt-5"
            disabled={savingProfile}
          >
            {savingProfile ? 'Saving…' : 'Save profile'}
          </Button>
        </form>
      )}

      {!notConnected && (
        <div className="flex flex-col gap-6">
          <div>
            <h3 className="text-xl font-semibold tracking-tight">Anime list</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              Track what you are watching and where you left off.
            </p>
            <div
              className="mt-5 -mx-1 flex gap-1 overflow-x-auto border-b border-border px-1"
              role="group"
              aria-label="Anime list status"
            >
              {listFilters.map((filter) => {
                const selected = listFilter === filter.value
                const sharing = sharingFilter === filter.value
                return (
                  <div
                    key={filter.value}
                    className="relative flex shrink-0 items-center"
                  >
                    <Button
                      type="button"
                      variant="ghost"
                      className={cn(
                        'relative h-11 min-h-11 px-3 text-muted-foreground hover:text-foreground',
                        selected && 'text-foreground',
                      )}
                      aria-pressed={selected}
                      onClick={() => void selectFilter(filter.value)}
                    >
                      {filter.label}
                      {!countsLoading && !countsError && (
                        <span className="text-xs tabular-nums text-muted-foreground">
                          {counts[filter.value] ?? 0}
                        </span>
                      )}
                      {selected && (
                        <span
                          className="absolute inset-x-3 bottom-0 h-0.5 bg-accent"
                          aria-hidden="true"
                        />
                      )}
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="h-11 w-11 text-muted-foreground hover:text-foreground"
                      aria-label={
                        sharing
                          ? `Sharing ${filter.label}`
                          : `Share ${filter.label}`
                      }
                      disabled={sharingFilter !== null}
                      onClick={() => void shareCategory(filter.value)}
                    >
                      <IconShare className="size-4" />
                    </Button>
                  </div>
                )
              })}
            </div>
          </div>

          {countsError && !error && (
            <Alert variant="destructive">
              <AlertDescription>
                Could not load list counts. {countsError}
              </AlertDescription>
              <AlertAction>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => void loadCounts()}
                >
                  Try again
                </Button>
              </AlertAction>
            </Alert>
          )}

          {loading ? (
            <ul
              className={posterGridClassName}
              aria-busy="true"
              aria-label="Loading your list"
            >
              {Array.from({length: 8}, (_, index) => (
                <li key={index}>
                  <Skeleton className="aspect-[2/3] w-full animate-pulse" />
                </li>
              ))}
            </ul>
          ) : error ? (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
              <AlertAction>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => void loadList()}
                >
                  Try again
                </Button>
              </AlertAction>
            </Alert>
          ) : entries.length === 0 ? (
            <p className="border border-dashed border-border/60 p-8 text-sm text-muted-foreground">
              {emptyCopy}
            </p>
          ) : (
            <section aria-label="Anime list entries">
              <ul className={posterGridClassName}>
                {entries.map((entry) => {
                  const title = entry.titleEnglish || entry.titleRomaji
                  return (
                    <li key={entry.mediaId}>
                      <LibraryPosterCard
                        title={title}
                        coverImage={entry.coverImage}
                        caption={entryPosterCaption(entry)}
                        subcaption={entryPosterSubcaption(entry)}
                        size="grid"
                        onClick={() => setEditingEntry(entry)}
                      />
                    </li>
                  )
                })}
              </ul>
            </section>
          )}
        </div>
      )}

      {editingEntry && (
        <WatchingEditSheet
          entry={editingEntry}
          saving={savingEntry}
          onClose={() => setEditingEntry(null)}
          onSave={(input) => void saveEntryWithState(input)}
        />
      )}

      <ShareBannerDialog
        key={
          shareSession
            ? `${shareSession.categoryLabel}-${shareSession.posters.map((poster) => poster.mediaId).join('-')}`
            : 'share-banner-closed'
        }
        open={shareSession !== null}
        session={shareSession}
        saving={savingBanner}
        onClose={() => setShareSession(null)}
        onSave={(pngBase64, defaultFilename) =>
          void saveShareBanner(pngBase64, defaultFilename)
        }
        onError={(message) => notice(message, true)}
      />
    </section>
  )
}
