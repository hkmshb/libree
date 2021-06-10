package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
)

const (
	Program     = "libree"
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

var config struct {
	apiEndpoint    string
	directory      string
	excludePattern string
	help           bool
}

func displayUsage(cmdText string, description string, cmd *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s %s", Program, cmdText)
	if description != "" {
		fmt.Fprint(os.Stderr, "\n\n", description)
	}

	fmt.Fprintln(os.Stderr, "\n\nOptions:")
	cmd.VisitAll(func(f *flag.Flag) {
		args := []string{f.Shorthand, f.Name}
		if f.DefValue == "" {
			args = append(args, f.Usage)
		} else {
			args = append(args, fmt.Sprintf("%s (%s)", f.Usage, f.DefValue))
		}

		fmt.Fprintf(os.Stderr, "  -%s, --%-7s %s\n", args[0], args[1], args[2])
	})
}

func handleIndex() error {
	cmd := flag.NewFlagSet("index", flag.ExitOnError)
	cmd.BoolVarP(&config.help, "help", "h", false, "Show this message and exit")
	cmd.StringVarP(&config.apiEndpoint, "url", "u", ServiceUrl, "url to api endpoint")

	cmd.Usage = func() {
		displayUsage("index directory [OPTIONS]", "Index entries to the database", cmd)
	}

	cmd.Parse(os.Args[2:])
	if cmd.NArg() != 1 || config.help {
		if !config.help {
			fmt.Fprint(os.Stderr, "error: directory argument missing\n\n")
		}

		cmd.Usage()
		os.Exit(0)
	}

	Url, err := url.Parse(config.apiEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	config.directory = os.ExpandEnv(cmd.Args()[0])
	if _, err := os.Stat(config.directory); os.IsNotExist(err) {
		log.Fatal(err)
	}

	service := Service{Url: Url, Username: "admin", Password: "c0uch"}
	stats := make(map[string]int)
	count := 0

	filepath.WalkDir(config.directory, func(path string, entry fs.DirEntry, err error) error {
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
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func handleTrim() error {
	cmd := flag.NewFlagSet("trim", flag.ExitOnError)
	cmd.BoolVarP(&config.help, "help", "h", false, "Show this message and exit")
	cmd.StringVarP(&config.excludePattern, "exclude", "x", "", "pattern for files to exclude")

	cmd.Usage = func() {
		displayUsage("trim [OPTIONS]", "Remove duplicate entries from filesystem and database", cmd)
	}

	cmd.Parse(os.Args[2:])
	if config.help {
		cmd.Usage()
		os.Exit(0)
	}

	log.Fatal("Not implemented!")
	return nil
}

func main() {
	commands := map[string]string{
		"index": "Index entries to the database",
		"trim":  "Remove duplicate entries from filesystem and database",
	}

	flag.BoolVarP(&config.help, "help", "h", false, "Show this message and exit")
	flag.Usage = func() {
		displayUsage("[OPTIONS] COMMANDS", "", flag.CommandLine)

		fmt.Fprintln(os.Stderr, "\nCommands:")
		for cmd, usage := range commands {
			fmt.Fprintf(os.Stderr, "  %-7s %s\n", cmd, usage)
		}
		fmt.Fprintln(os.Stderr, "")
	}

	flag.Parse()
	if flag.NArg() == 0 || (flag.NArg() == 0 && config.help) {
		flag.Usage()
		os.Exit(0)
	}

	if _, ok := commands[os.Args[1]]; ok {
		switch os.Args[1] {
		case "index":
			handleIndex()
		case "trim":
			handleTrim()
		}
		return
	}

	flag.Usage()
}
