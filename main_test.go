package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const base = "http://localhost:1234"

func TestNormalizePrefix(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{
			input:  "/foo/",
			output: "/foo/",
		},
		{
			input:  "/foo",
			output: "/foo/",
		},
		{
			input:  "foo/",
			output: "/foo/",
		},
		{
			input:  "foo",
			output: "/foo/",
		},
		{
			input:  "foo/bar",
			output: "/foo/bar/",
		},
		{
			input:  "../foo",
			output: "/foo/",
		},
		{
			input:  "",
			output: "/",
		},
		{
			input:  "///foo///bar",
			output: "/foo/bar/",
		},
		{
			input:  "/foo/../bar",
			output: "/bar/",
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			prefix, err := normalizePrefix(base, c.input)
			require.NoError(t, err)

			assert.Equal(t, c.output, prefix)
		})
	}
}

func mustNormalizePrefix(prefix string) string {
	p, err := normalizePrefix(base, prefix)
	if err != nil {
		panic(err)
	}
	return p
}

func mustJoinPath(parts ...string) string {
	p, err := url.JoinPath(base, parts...)
	if err != nil {
		panic(err)
	}
	u, err := url.Parse(p)
	if err != nil {
		panic(err)
	}
	return u.Path
}

func TestServeIndex(t *testing.T) {
	testDir := "root"

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: mustNormalizePrefix("/"),
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	title := fmt.Sprintf("<title>Index of %s</title>", testDir)
	assert.Contains(t, string(body), title)
}

func TestServeIndexWithPrefix(t *testing.T) {
	testDir := "root"
	prefix := mustNormalizePrefix("/foo/bar/")

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: prefix,
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", prefix, nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	title := fmt.Sprintf("<title>Index of %s</title>", testDir)
	assert.Contains(t, string(body), title)
}

func TestServeWithPrefixNotFound(t *testing.T) {
	testDir := "root"
	prefix := mustNormalizePrefix("/foo/bar/")

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: prefix,
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestServeFile(t *testing.T) {
	testDir := "root"
	filePath := "file.txt"

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: mustNormalizePrefix("/"),
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", fmt.Sprintf("/%s", filePath), nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/plain; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("%s/%s\n", testDir, filePath), string(body))
}

func TestServeFileWithPrefix(t *testing.T) {
	testDir := "root"
	filePath := "file.txt"
	prefix := mustNormalizePrefix("/some/prefix/")

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: prefix,
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", mustJoinPath(s.Prefix, filePath), nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/plain; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("%s/%s\n", testDir, filePath), string(body))
}

func TestServeFileWithPrefixNotFound(t *testing.T) {
	testDir := "root"
	filePath := "file.txt"
	prefix := mustNormalizePrefix("/some/prefix/")

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: prefix,
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", "/"+filePath, nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestServeCSS(t *testing.T) {
	testDir := "root"
	filePath := "style.css"

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: mustNormalizePrefix("/"),
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", fmt.Sprintf("/%s", filePath), nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/css; charset=utf-8", response.Header.Get("Content-Type"))
}

func TestServeSubIndex(t *testing.T) {
	testDir := "root"
	dirPath := "sub"

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: mustNormalizePrefix("/"),
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", fmt.Sprintf("/%s/", dirPath), nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	title := fmt.Sprintf("<title>Index of %s/%s</title>", testDir, dirPath)
	assert.Contains(t, string(body), title)
}

func TestServeSubIndexRedirect(t *testing.T) {
	testDir := "root"
	dirPath := "sub"

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: mustNormalizePrefix("/"),
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", fmt.Sprintf("/%s/index.html", dirPath), nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusMovedPermanently, response.StatusCode)
	assert.Equal(t, "./", response.Header.Get("Location"))
}

func TestServeSubDirRedirect(t *testing.T) {
	testDir := "root"
	dirPath := "sub"

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: mustNormalizePrefix("/"),
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", fmt.Sprintf("/%s", dirPath), nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusMovedPermanently, response.StatusCode)
	assert.Equal(t, fmt.Sprintf("%s/", dirPath), response.Header.Get("Location"))
}

func TestServeSubFile(t *testing.T) {
	testDir := "root"
	filePath := "sub/file.txt"

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: mustNormalizePrefix("/"),
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", fmt.Sprintf("/%s", filePath), nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/plain; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("%s/%s\n", testDir, filePath), string(body))
}

func TestServeSubFilePrefix(t *testing.T) {
	testDir := "root"
	filePath := "sub/file.txt"
	prefix := mustNormalizePrefix("some/prefix")

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: prefix,
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", mustJoinPath(prefix, filePath), nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/plain; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("%s/%s\n", testDir, filePath), string(body))
}

func TestServeCustomIndex(t *testing.T) {
	testDir := "root-with-index"

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: mustNormalizePrefix("/"),
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, "root-with-index/index.html\n", string(body))
}

func TestServeCustomSubIndex(t *testing.T) {
	testDir := "root-with-index"
	dirPath := "sub-with-index"

	s := &Serve{
		Dir:    fmt.Sprintf("testdata/%s", testDir),
		Prefix: mustNormalizePrefix("/"),
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", fmt.Sprintf("/%s/", dirPath), nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, "root-with-index/sub-with-index/index.html\n", string(body))
}

func TestServeExplicitIndexNotInPath(t *testing.T) {
	testDir := "root-with-index"

	s := &Serve{
		Dir:           fmt.Sprintf("testdata/%s", testDir),
		Prefix:        mustNormalizePrefix("/"),
		ExplicitIndex: true,
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	title := fmt.Sprintf("<title>Index of %s</title>", testDir)
	assert.Contains(t, string(body), title)
}

func TestServeExplicitIndexInPath(t *testing.T) {
	testDir := "root-with-index"

	s := &Serve{
		Dir:           fmt.Sprintf("testdata/%s", testDir),
		Prefix:        mustNormalizePrefix("/"),
		ExplicitIndex: true,
	}

	handler := s.handler()

	request := httptest.NewRequest("GET", "/index.html", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, "root-with-index/index.html\n", string(body))
}
