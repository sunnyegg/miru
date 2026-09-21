import {usePlaybackStore} from '../stores/playbackStore'

export function NowPlayingStrip() {
  const playing = usePlaybackStore((state) => state.playing)
  if (!playing) {
    return null
  }

  return (
    <div
      className="border-b border-border bg-bezel px-4 py-2 text-sm"
      role="status"
    >
      Playing · {Math.round(playing.percent)}%
    </div>
  )
}
