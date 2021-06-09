package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

const (
	ServiceName = "mega"
	ServiceUrl  = "http://localhost:5984/libree"
)

type Storage struct {
	Account string `json:"account,omitempty"`
	Service string `json:"service,omitempty"`
}

type FileDoc struct {
	ID        string  `json:"_id"`
	DocType   string  `json:"docType"`
	BasePath  string  `json:"basePath"`
	Filename  string  `json:"filename"`
	Extension string  `json:"ext"`
	Storage   Storage `json:"storage,omitempty"`
}

func (fd FileDoc) Buffer() (*bytes.Buffer, error) {
	data, err := json.Marshal(fd)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)
	return buf, nil
}

type Service struct {
	Url      *url.URL
	Username string
	Password string
}

func (s Service) Post(f *FileDoc) {
	buf, err := f.Buffer()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", s.Url.String(), buf)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(s.Username, s.Password)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}

func main() {
	// define flags
	apiUrl := flag.String("u", ServiceUrl, "Target API service url")
	directory := flag.String("d", "", "Path to directory to process")
	flag.Parse()

	Url, err := url.Parse(*apiUrl)
	if err != nil {
		log.Fatal((err))
	}

	service := Service{Url: Url, Username: "admin", Password: "c0uch"}
	stats := make(map[string]int)
	count := 0

	filepath.WalkDir(*directory, func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}

		hash := sha1.New()
		io.WriteString(hash, strings.Join([]string{ServiceName, path}, "/"))

		fd := FileDoc{
			ID:        hex.EncodeToString(hash.Sum(nil)),
			DocType:   "file",
			BasePath:  filepath.Dir(path),
			Filename:  entry.Name(),
			Extension: filepath.Ext(path),
			Storage:   Storage{Service: ServiceName},
		}

		service.Post(&fd)

		// track duplicates
		stats[entry.Name()]++
		count++

		// display progress
		if count%10 == 0 {
			fmt.Print(".")
			if count%1000 == 0 {
				fmt.Println("")
			}
		}

		return nil
	})

	data, err := json.MarshalIndent(stats, "", "  ")
	if err == nil {
		fmt.Println(string(data))
	}
}
