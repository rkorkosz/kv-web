package main

import (
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutSuccess(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testput.db")
	kv := NewBoltKV(dbPath)
	h := NewHandler(kv)

	buf := strings.NewReader("testdata")
	req := httptest.NewRequest("PUT", "/test/put/data", buf)
	w := httptest.NewRecorder()
	h.put(w, req)
	resp := w.Result()
	assert.Equal(t, 204, resp.StatusCode)
}

func TestGetSuccess(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testget.db")
	kv := NewBoltKV(dbPath)
	h := NewHandler(kv)

	data := "testdata"
	buf := strings.NewReader(data)
	req := httptest.NewRequest("PUT", "/test/get/data", buf)
	w := httptest.NewRecorder()
	h.put(w, req)
	resp := w.Result()
	require.Equal(t, 204, resp.StatusCode)

	req = httptest.NewRequest("GET", "/test/get/data", nil)
	w = httptest.NewRecorder()
	h.get(w, req)
	resp = w.Result()
	assert.Equal(t, 200, resp.StatusCode)
	body := make([]byte, len(data))
	_, err := resp.Body.Read(body)
	require.NoError(t, err)
	assert.Equal(t, data, string(body))
}

func TestGetNotExisting(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testget.db")
	kv := NewBoltKV(dbPath)
	h := NewHandler(kv)

	req := httptest.NewRequest("GET", "/test/get/data", nil)
	w := httptest.NewRecorder()
	h.get(w, req)
	resp := w.Result()
	assert.Equal(t, 404, resp.StatusCode)
}

func TestDeleteSuccess(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testdelete.db")
	kv := NewBoltKV(dbPath)
	h := NewHandler(kv)

	buf := strings.NewReader("testdata")
	req := httptest.NewRequest("PUT", "/test/delete/data", buf)
	w := httptest.NewRecorder()
	h.put(w, req)
	resp := w.Result()
	require.Equal(t, 204, resp.StatusCode)

	req = httptest.NewRequest("DELETE", "/test/delete/data", nil)
	w = httptest.NewRecorder()
	h.delete(w, req)
	resp = w.Result()
	assert.Equal(t, 204, resp.StatusCode)

	req = httptest.NewRequest("GET", "/test/delete/data", nil)
	w = httptest.NewRecorder()
	h.get(w, req)
	resp = w.Result()
	assert.Equal(t, 404, resp.StatusCode)
}

func TestDeleteNotExisting(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testdelete.db")
	kv := NewBoltKV(dbPath)
	h := NewHandler(kv)

	req := httptest.NewRequest("DELETE", "/test/delete/data", nil)
	w := httptest.NewRecorder()
	h.get(w, req)
	resp := w.Result()
	assert.Equal(t, 404, resp.StatusCode)
}
