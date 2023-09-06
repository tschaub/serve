package main

import (
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/rs/cors"
)

func main() {
	ctx := kong.Parse(&ServeCmd{}, kong.UsageOnError())
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

type ServeCmd struct {
	Port int    `help:"Listen on this port." default:"4000"`
	Dir  string `help:"Serve files from this directory." arg:"" type:"existingdir"`
	Cors bool   `help:"Include CORS support (on by default)." default:"true" negatable:""`
	Dot  bool   `help:"Serve dot files (files prefixed with a '.')" default:"false"`
}

func (c *ServeCmd) Run() error {
	server := &Server{
		dir:  c.Dir,
		port: c.Port,
		cors: c.Cors,
		dot:  c.Dot,
	}

	return server.Start()
}

type Server struct {
	dir  string
	port int
	cors bool
	dot  bool
}

func excludeDot(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		parts := strings.Split(request.URL.Path, "/")
		for _, part := range parts {
			if strings.HasPrefix(part, ".") {
				http.NotFound(response, request)
				return
			}
		}

		handler.ServeHTTP(response, request)
	})
}

type IndexData struct {
	Dir     string
	Parents []*Entry
	Entries []*Entry
}

type Entry struct {
	Name string
	Path string
	Type string
}

const (
	fileType   = "file"
	folderType = "folder"
)

//go:embed index.html
var indexHtml string

func withIndex(dir string, dot bool, handler http.Handler) http.Handler {
	indexTemplate := template.Must(template.New("index").Parse(indexHtml))
	base := filepath.Base(dir)
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		urlPath := request.URL.Path
		if !strings.HasSuffix(urlPath, "/") {
			handler.ServeHTTP(response, request)
			return
		}

		dirPath := filepath.Join(dir, urlPath)
		list, dirErr := os.ReadDir(dirPath)
		if dirErr != nil {
			if errors.Is(dirErr, os.ErrNotExist) {
				http.NotFound(response, request)
				return
			}
			http.Error(response, dirErr.Error(), http.StatusInternalServerError)
			return
		}

		entries := []*Entry{}
		for _, item := range list {
			name := item.Name()
			if !dot && strings.HasPrefix(name, ".") {
				continue
			}
			entry := &Entry{
				Name: name,
				Path: path.Join(urlPath, name),
			}
			if item.IsDir() {
				entry.Type = folderType
				entry.Path = entry.Path + "/"
			} else {
				entry.Type = fileType
			}
			entries = append(entries, entry)
		}
		sort.Slice(entries, func(i int, j int) bool {
			iEntry := entries[i]
			jEntry := entries[j]
			if iEntry.Type == folderType && jEntry.Type != folderType {
				return true
			}
			if jEntry.Type == folderType && iEntry.Type != folderType {
				return false
			}
			return iEntry.Name < jEntry.Name
		})

		if urlPath != "/" {
			parentEntry := &Entry{
				Name: "..",
				Path: path.Join(urlPath, ".."),
				Type: folderType,
			}
			entries = append([]*Entry{parentEntry}, entries...)
		}

		parentParts := strings.Split(urlPath, "/")
		parentParts = parentParts[:len(parentParts)-1]
		parentEntries := make([]*Entry, len(parentParts))
		for i, part := range parentParts {
			entry := &Entry{
				Name: part,
				Path: strings.Join(parentParts[:i+1], "/") + "/",
				Type: folderType,
			}
			if part == "" {
				entry.Name = base
			}
			parentEntries[i] = entry
		}

		data := &IndexData{
			Dir:     filepath.Join(base, urlPath),
			Entries: entries,
			Parents: parentEntries,
		}

		response.WriteHeader(http.StatusOK)
		if err := indexTemplate.Execute(response, data); err != nil {
			fmt.Printf("trouble executing template: %s\n", err)
		}
	})
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	dir := http.Dir(s.dir)
	mux.Handle("/", http.FileServer(dir))

	handler := withIndex(string(dir), s.dot, http.Handler(mux))
	if !s.dot {
		handler = excludeDot(handler)
	}
	if s.cors {
		handler = cors.Default().Handler(handler)
	}

	fmt.Printf("Serving %s on http://localhost:%d/\n", s.dir, s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), handler)
}
