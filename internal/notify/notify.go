package notify

import "github.com/gen2brain/beeep"

func DownloadComplete(fileName string) error {
	body := fileName
	if body == "" {
		body = "Download finished"
	}
	return beeep.Notify("Download complete", body, "")
}
