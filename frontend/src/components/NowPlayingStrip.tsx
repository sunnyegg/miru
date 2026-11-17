import {useState} from 'react'
import {errorMessage} from '../lib/format'
import {SeekPlayback} from '../../wailsjs/go/main/App'
import {usePlaybackStore} from '../stores/playbackStore'

type Props = {
  notice: (text: string, error?: boolean) => void
}

function formatTime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return '0:00'
  }
  const total = Math.floor(seconds)
  const hours = Math.floor(total / 3600)
  const minutes = Math.floor((total % 3600) / 60)
  const rest = total % 60
  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(rest).padStart(2, '0')}`
  }
  return `${minutes}:${String(rest).padStart(2, '0')}`
}

// NowPlayingStrip is the timeline only; pause, volume, full screen and stop
// live in the MPV window's own controls (OSC).
export function NowPlayingStrip({notice}: Props) {
  const playing = usePlaybackStore((state) => state.playing)
  const trackProgress = usePlaybackStore((state) => state.trackProgress)
  const [seekDraft, setSeekDraft] = useState<number | null>(null)

  if (!playing) {
    return null
  }

  async function commitSeek() {
    if (seekDraft === null) {
      return
    }
    const target = seekDraft
    setSeekDraft(null)
    try {
      trackProgress(await SeekPlayback(target))
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  const position = seekDraft ?? playing.position
  const max = Math.max(playing.duration, 0)
  const current = Math.min(Math.max(position, 0), max)
  const fill = max > 0 ? `${(current / max) * 100}%` : '0%'

  return (
    <div
      className="flex items-center gap-3 border-b border-border bg-bezel px-3 py-1.5"
      role="region"
      aria-label="Playback timeline"
    >
      <span className="min-w-0 max-w-[45%] shrink truncate text-sm">
        {playing.title}
      </span>
      <div className="relative h-6 min-w-[100px] flex-1">
        <div className="pointer-events-none absolute inset-x-0 top-1/2 h-0.5 -translate-y-1/2 bg-muted">
          <div className="h-full bg-primary" style={{width: fill}} />
        </div>
        <input
          type="range"
          aria-label="Seek"
          min={0}
          max={max}
          step={0.5}
          value={current}
          disabled={max <= 0}
          onChange={(event) => setSeekDraft(Number(event.target.value))}
          onPointerUp={() => void commitSeek()}
          onKeyUp={() => void commitSeek()}
          onBlur={() => void commitSeek()}
          className="absolute inset-0 h-6 w-full cursor-pointer appearance-none bg-transparent disabled:cursor-not-allowed [&::-webkit-slider-runnable-track]:h-0.5 [&::-webkit-slider-runnable-track]:bg-transparent [&::-webkit-slider-thumb]:mt-[-3px] [&::-webkit-slider-thumb]:size-2 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:bg-primary"
        />
      </div>
      <span className="shrink-0 text-xs tabular-nums text-muted-foreground">
        {formatTime(current)} / {formatTime(playing.duration)}
      </span>
    </div>
  )
}
