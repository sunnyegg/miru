import {libraryHeroBackgroundImage} from '../lib/anilistImage'
import type {ShowGroup} from '../lib/groupEpisodes'
import {LibraryAddToWatchingBanner} from './LibraryAddToWatchingBanner'
import {Button} from '@/components/ui/button'

type Props = {
  show: ShowGroup
  bannerImage?: string
  showAddToWatching: boolean
  saving: boolean
  unmatching: boolean
  onAddToWatching: () => void
  onMatchAnilist: () => void
  onUnmatchAnilist: () => void
}

function progressSummary(show: ShowGroup): string {
  if (!show.bound) {
    const count = show.episodes.length
    return `${count} local episode${count === 1 ? '' : 's'} · not linked to AniList`
  }
  if (show.totalEpisodes > 0) {
    const remaining = show.totalEpisodes - show.progress
    if (remaining > 0) {
      return `Episode ${show.progress} / ${show.totalEpisodes} · ${remaining} left`
    }
    return `Episode ${show.progress} / ${show.totalEpisodes}`
  }
  return `${show.progress} watched`
}

export function LibraryShowDetailHero({
  show,
  bannerImage = '',
  showAddToWatching,
  saving,
  unmatching,
  onAddToWatching,
  onMatchAnilist,
  onUnmatchAnilist,
}: Props) {
  const heroBackground = libraryHeroBackgroundImage(
    bannerImage ? {bannerImage, coverImage: show.coverImage} : null,
    show,
  )

  return (
    <div className="relative -mx-5 mb-6 overflow-hidden px-5">
      <div className="relative flex min-h-52 flex-col justify-end sm:min-h-64">
        {heroBackground ? (
          <img
            src={heroBackground}
            alt=""
            referrerPolicy="no-referrer"
            className="absolute inset-0 size-full object-cover object-top opacity-50"
          />
        ) : (
          <div className="absolute inset-0 bg-muted" />
        )}
        <div className="absolute inset-0 bg-gradient-to-r from-background via-background/85 to-background/40" />
        <div className="relative pt-20 pb-6 sm:pt-24">
          <h2 className="max-w-2xl text-2xl font-semibold tracking-tight sm:text-3xl">
            {show.title}
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {progressSummary(show)}
          </p>
          {showAddToWatching && (
            <div className="mt-4">
              <LibraryAddToWatchingBanner
                show={show}
                saving={saving}
                onAddToWatching={onAddToWatching}
                onMatchAnilist={onMatchAnilist}
              />
            </div>
          )}
          {show.bound && (
            <div className="mt-4">
              <Button
                type="button"
                variant="secondary"
                disabled={unmatching}
                onClick={onUnmatchAnilist}
              >
                {unmatching ? 'Removing match…' : 'Unmatch AniList'}
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
