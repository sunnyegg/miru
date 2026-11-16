import {useState} from 'react'
import {Button} from '@/components/ui/button'
import {errorMessage} from '../lib/format'
import {
  SeekPlayback,
  SetPlaybackVolume,
  StopPlayback,
  TogglePlaybackFullscreen,
  TogglePlaybackPause,
} from '../../wailsjs/go/main/App'
import {usePlaybackStore} from '../stores/playbackStore'
import {
  IconClose,
  IconFullscreen,
  IconFullscreenExit,
  IconPause,
  IconPlay,
  IconVolume,
} from './Icons'
import {cn} from '@/lib/utils'

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

type SliderProps = {
  label: string
  value: number
  max: number
  step: number
  disabled?: boolean
  className?: string
  onInput: (value: number) => void
  onCommit?: () => void
}

function Slider({
  label,
  value,
  max,
  step,
  disabled,
  className,
  onInput,
  onCommit,
}: SliderProps) {
  return (
    <input
      type="range"
      aria-label={label}
      min={0}
      max={Math.max(max, 0)}
      step={step}
      value={value}
      disabled={disabled || max <= 0}
      onChange={(event) => onInput(Number(event.target.value))}
      onPointerUp={onCommit}
      onKeyUp={onCommit}
      onBlur={onCommit}
      className={cn(
        'h-11 w-full min-w-0 cursor-pointer appearance-none bg-transparent disabled:cursor-not-allowed',
        '[&::-webkit-slider-runnable-track]:h-1 [&::-webkit-slider-runnable-track]:bg-muted',
        '[&::-webkit-slider-thumb]:mt-[-4px] [&::-webkit-slider-thumb]:size-3 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:bg-primary',
        className,
      )}
    />
  )
}

export function NowPlayingStrip({notice}: Props) {
  const playing = usePlaybackStore((state) => state.playing)
  const trackProgress = usePlaybackStore((state) => state.trackProgress)
  const clearPlaying = usePlaybackStore((state) => state.clearPlaying)
  const [seekDraft, setSeekDraft] = useState<number | null>(null)

  if (!playing) {
    return null
  }

  async function togglePause() {
    try {
      trackProgress(await TogglePlaybackPause())
    } catch (err) {
      notice(errorMessage(err), true)
    }
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

  async function changeVolume(value: number) {
    try {
      trackProgress(await SetPlaybackVolume(value))
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function toggleFullscreen() {
    try {
      trackProgress(await TogglePlaybackFullscreen())
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function stopPlayback() {
    try {
      await StopPlayback()
      clearPlaying()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  const position = seekDraft ?? playing.position

  return (
    <div
      className="flex items-center gap-2 border-b border-border bg-bezel px-3 py-1"
      role="region"
      aria-label="Playback controls"
    >
      <Button
        type="button"
        size="icon"
        aria-label={playing.paused ? 'Play' : 'Pause'}
        onClick={() => void togglePause()}
      >
        {playing.paused ? <IconPlay /> : <IconPause />}
      </Button>
      <div className="min-w-0 flex-1">
        <div className="flex items-baseline justify-between gap-3">
          <span className="truncate text-sm">{playing.title}</span>
          <span className="shrink-0 text-xs tabular-nums text-muted-foreground">
            {formatTime(position)} / {formatTime(playing.duration)}
          </span>
        </div>
        <Slider
          label="Seek"
          value={Math.min(position, playing.duration)}
          max={playing.duration}
          step={0.5}
          onInput={setSeekDraft}
          onCommit={() => void commitSeek()}
        />
      </div>
      <div className="hidden items-center gap-1 md:flex">
        <IconVolume className="size-4 text-muted-foreground" />
        <Slider
          label="Volume"
          className="w-24"
          value={playing.volume}
          max={100}
          step={1}
          onInput={(value) => void changeVolume(value)}
        />
      </div>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        aria-label={
          playing.fullscreen ? 'Exit full screen' : 'Full screen player'
        }
        onClick={() => void toggleFullscreen()}
      >
        {playing.fullscreen ? <IconFullscreenExit /> : <IconFullscreen />}
      </Button>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        aria-label="Stop playback"
        onClick={() => void stopPlayback()}
      >
        <IconClose />
      </Button>
    </div>
  )
}
