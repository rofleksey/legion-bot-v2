package util

func IsStaticAsset(path string) bool {
	extensions := []string{".mp4", ".webm", ".js", ".css", ".png", ".jpg", ".woff2"}
	for _, ext := range extensions {
		if len(path) > len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}
	return false
}
