import type {RefObject} from 'react'
import type {ShowGroup} from '../lib/groupEpisodes'
import type {EpisodeView, PlaybackEvent} from '../lib/types'
import {IconPlay} from './Icons'
import {Button} from '@/components/ui/button'

type Props = {
  selectedShow: ShowGroup | null
  selectedEpisode: EpisodeView | null
  playing: PlaybackEvent | null
  busyId: number | null
  loading: boolean
  loadError: string
  selectedEpisodeButtonRef: RefObject<HTMLButtonElement | null>
  onSelectEpisode: (episodeId: number) => void
  onPlay: (episodeId: number) => void
}

export function LibraryOsc({
  selectedShow,
  selectedEpisode,
  playing,
  busyId,
  loading,
  loadError,
  selectedEpisodeButtonRef,
  onSelectEpisode,
  onPlay,
}: Props) {
  const selectedEpisodeIsPlaying = Boolean(
    playing && selectedEpisode && playing.episodeId === selectedEpisode.id,
  )
  const progress = selectedEpisodeIsPlaying && Number.isFinite(playing?.percent)
    ? Math.min(100, Math.max(0, playing?.percent ?? 0))
    : 0
  const fileLabel = selectedEpisode
    ? selectedEpisode.filePath.split(/[\\/]/).pop() || selectedEpisode.displayTitle
    : ''

  return (
    <div
      className="relative shrink-0 border-t border-border bg-bezel py-3"
      role="region"
      aria-label="Playback"
      style={{paddingInline: 16}}
    >
      <div
        className="pointer-events-none absolute inset-x-0 top-0 h-0.5 bg-muted"
        role={selectedEpisodeIsPlaying ? 'progressbar' : undefined}
        aria-hidden={selectedEpisodeIsPlaying ? undefined : true}
        aria-label={selectedEpisodeIsPlaying ? `Playback progress for ${selectedEpisode?.displayTitle}` : undefined}
        aria-valuemin={selectedEpisodeIsPlaying ? 0 : undefined}
        aria-valuemax={selectedEpisodeIsPlaying ? 100 : undefined}
        aria-valuenow={selectedEpisodeIsPlaying ? Math.round(progress) : undefined}
        aria-valuetext={selectedEpisodeIsPlaying ? `${Math.round(progress)}% played` : undefined}
      >
        <div
          className="osc-progress-fill h-full bg-accent transition-[width] duration-200"
          style={{width: `${progress}%`}}
        />
      </div>
      {selectedShow && selectedEpisode ? (
        <div key={selectedShow.key} className="osc-drop flex min-w-0 items-center" style={{gap: 12}}>
          <p className="min-w-0 max-w-[28%] shrink-0 truncate text-sm text-foreground" title={fileLabel}>
            {fileLabel}
          </p>
          <div className="min-w-0 flex-1 overflow-x-auto py-1" role="group" aria-label="Episodes">
            <div className="relative flex w-max min-w-full items-center justify-between px-1">
              <div className="pointer-events-none absolute inset-x-1 top-1/2 h-px bg-border" aria-hidden="true" />
              {selectedShow.episodes.map((episode) => {
                const current = episode.id === selectedEpisode.id
                const label = episode.episodeNumber > 0 ? `Episode ${episode.episodeNumber}` : episode.displayTitle
                return (
                  <button
                    key={episode.id}
                    ref={current ? selectedEpisodeButtonRef : undefined}
                    type="button"
                    aria-label={label}
                    aria-pressed={current}
                    onClick={() => onSelectEpisode(episode.id)}
                    className="relative flex h-11 min-w-6 shrink-0 cursor-pointer items-center justify-center"
                  >
                    <span
                      className={`block h-3 w-0.5 ${current ? 'h-4 bg-accent' : 'bg-muted-foreground'}`}
                    />
                  </button>
                )
              })}
            </div>
          </div>
          <Button
            type="button"
            onClick={() => onPlay(selectedEpisode.id)}
            disabled={busyId === selectedEpisode.id}
            style={{minWidth: 96, gap: 8, paddingInline: 20}}
          >
            <span className="shrink-0" style={{width: 16, height: 16}}>
              <IconPlay className="size-full" />
            </span>
            {busyId === selectedEpisode.id ? 'Starting…' : 'Play'}
          </Button>
        </div>
      ) : (
        <div className="osc-drop flex min-w-0 items-center" style={{gap: 12}}>
          <p className="min-w-0 flex-1 truncate text-sm text-muted-foreground">
            {loading ? 'Loading library…' : loadError ? 'Library unavailable' : 'No local shows yet'}
          </p>
          <Button type="button" disabled style={{minWidth: 96, gap: 8, paddingInline: 20}}>
            <span className="shrink-0" style={{width: 16, height: 16}}>
              <IconPlay className="size-full" />
            </span>
            Play
          </Button>
        </div>
      )}
    </div>
  )
}
