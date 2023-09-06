package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeIndex(t *testing.T) {
	testDir := "root"

	s := &Serve{
		Dir: fmt.Sprintf("testdata/%s", testDir),
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

func TestServeFile(t *testing.T) {
	testDir := "root"
	filePath := "file.txt"

	s := &Serve{
		Dir: fmt.Sprintf("testdata/%s", testDir),
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

func TestServeSubIndex(t *testing.T) {
	testDir := "root"
	dirPath := "sub"

	s := &Serve{
		Dir: fmt.Sprintf("testdata/%s", testDir),
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
		Dir: fmt.Sprintf("testdata/%s", testDir),
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
		Dir: fmt.Sprintf("testdata/%s", testDir),
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
		Dir: fmt.Sprintf("testdata/%s", testDir),
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

func TestServeCustomIndex(t *testing.T) {
	testDir := "root-with-index"

	s := &Serve{
		Dir: fmt.Sprintf("testdata/%s", testDir),
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
		Dir: fmt.Sprintf("testdata/%s", testDir),
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
