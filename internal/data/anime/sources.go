package anime

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
)

const AnimeOfflineDatabaseURL = "https://raw.githubusercontent.com/manami-project/anime-offline-database/master/anime-offline-database-minified.json"
const AniDBDumpURL = "https://anidb.net/api/anime-titles.dat.gz"

func DownloadAnimeOfflineDatabase(path string) (err error) {
	resp, err := http.Get(AnimeOfflineDatabaseURL)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	file, err := os.Create(path)

	if err != nil {
		return
	}

	defer file.Close()

	_, err = io.Copy(file, resp.Body)

	return
}

func DownloadAniDBDump(path string) (err error) {
	req, err := http.NewRequest("GET", AniDBDumpURL, nil)

	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:69.0) Gecko/20100101 Firefox/69.0")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	file, err := os.Create(path)

	if err != nil {
		return
	}

	defer file.Close()

	reader, err := gzip.NewReader(resp.Body)

	if err != nil {
		return
	}

	defer reader.Close()

	_, err = io.Copy(file, reader)

	return
}
