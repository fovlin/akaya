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

	response(w, r, r.URL.Path)
	
}

func setMime(fileName string, w http.ResponseWriter) {
	ext := path.Ext(fileName)
	if mimeType := mime.TypeByExtension(ext); len(mimeType) != 0 {
		w.Header().Set("Content-Type", mimeType)
	}
}

func response(w http.ResponseWriter, r *http.Request, url string) {

	root, err := os.OpenRoot(config.Root)
	if err != nil {
		record.Warn("(open resourse directory)", err)
	}
	defer root.Close()

	url = path.Join("./", url)

	fileSata, err := os.Stat(url)
	if os.IsNotExist(err) {
		record.Warn(err)
		http.Redirect(w, r, "/", http.StatusNotFound)
		return
	}

	if fileSata.IsDir() {

		indexFile, err := root.Open(path.Join(url, "index.html"))
		if err == nil {
			defer indexFile.Close()
			io.Copy(w, indexFile)
		}

		entry, err := indexDir(url)
		if err != nil {
			record.Warn("read directory:", err)
			return
		}

		data, err := json.MarshalIndent(entry, "", "	")
		if err != nil {
			record.Warn("load directory list:", err)
		}

		w.Header().Set("Content-Type", "text/json")
		w.Write(data)

	} else {
		file, err := root.Open(url)
		if err != nil {
			record.Warn("open file", err)
			http.Redirect(w, r, "/", http.StatusForbidden)
			return
		}

		setMime(url, w)

		_, err = io.Copy(w, file)
		if err != nil {
			record.Warn("send file:", err)
			http.Redirect(w, r, "/", http.StatusForbidden)
			return
		}
	}
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