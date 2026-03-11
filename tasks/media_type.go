package tasks

import "strings"

func isVideoFileName(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	return strings.HasSuffix(lower, ".mp4") ||
		strings.HasSuffix(lower, ".mov") ||
		strings.HasSuffix(lower, ".m4v") ||
		strings.HasSuffix(lower, ".webm") ||
		strings.HasSuffix(lower, ".avi")
}

func containsVideoFile(files []string) bool {
	for _, file := range files {
		if isVideoFileName(file) {
			return true
		}
	}
	return false
}
