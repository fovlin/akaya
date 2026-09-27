package main

import (
	"fmt"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/fovlin/record"
)

type handler struct{}

func (handler handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	record.Info("receive request: url:", r.RequestURI, "adddress:", r.RemoteAddr)
	if f, ok := specURL[r.RequestURI]; ok {
		f(w, r)
		return
	}

	err := response(w, r.URL.Path)
	if err != nil {
		record.Warn(err)
		fmt.Fprint(w, err)
	}
}

func setMime(fileName string, w http.ResponseWriter) {
	ext := path.Ext(fileName)
	if mimeType := mime.TypeByExtension(ext); len(mimeType) != 0 {
		w.Header().Set("Content-Type", mimeType)
	}
}

func response(w http.ResponseWriter, url string) error {

	url = path.Join(config.Root, url)

	fileSata, err := os.Stat(url)
	if err != nil {
		return err
	}

	if !fileSata.IsDir() {
		w.Header().Set("connection", "keep-alive")
		w.Header().Set("Content-Length", fmt.Sprint(fileSata.Size()))
		w.Header().Set("Accept-Ranges", "bytes")
		err := fileResponse(url, w)
		if err != nil {
			return err
		}
	} else {
		dirResponse(url, w)
	}
	return nil
}

func dirResponse(url string, w http.ResponseWriter) error {
	entry, err := indexDir(url)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(entry, "", "	")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write(data)

	return nil
}

func fileResponse(url string, w http.ResponseWriter) error {
	file, err := os.Open(url)
	if err != nil {
		return err
	}

	setMime(url, w)

	_, err = io.Copy(w, file)
	if err != nil {
		return err
	}
	return nil
}

func indexDir(url string) (map[string]DirEntry, error) {
	var entry map[string]DirEntry = map[string]DirEntry{}

	dirList, err := os.ReadDir(url)
	if err != nil {
		record.Warn(err)
		return nil, err
	}

	for _, dirEntry := range(dirList) {

		stat, err := os.Stat(path.Join(url, dirEntry.Name()))
		if err != nil {
			record.Warn("read directory:", err)
			continue
		}

		var filetype string
		if !stat.IsDir() {
			filetype = "file"
		} else {
			filetype = "directory"
		}
		modTime := stat.ModTime().Format(time.DateTime)
		size := fmt.Sprint(stat.Size())

		entry[stat.Name()] = DirEntry{
			Type: filetype,
			Time: modTime,
			Size: size,
		}
	}

	return entry, nil
}