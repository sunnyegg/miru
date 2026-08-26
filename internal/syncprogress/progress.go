package syncprogress

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
