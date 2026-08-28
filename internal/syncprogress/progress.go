package syncprogress

// ShouldAttemptRealtimeSync reports whether playback progress should trigger an
// immediate AniList sync during MPV polling (at most once per play session).
func ShouldAttemptRealtimeSync(percent, threshold float64, alreadySynced, mapFailed, attempted bool) bool {
	if alreadySynced || mapFailed || attempted {
		return false
	}
	return percent >= threshold
}

func ShouldSync(percent, threshold float64, episode, currentProgress int, alreadySynced bool) bool {
	if alreadySynced {
		return false
	}
	if percent < threshold {
		return false
	}
	if episode <= 0 {
		return false
	}
	if episode <= currentProgress {
		return false
	}
	return true
}
