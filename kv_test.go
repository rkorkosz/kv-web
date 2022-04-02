package main

import (
	"bufio"
	"fmt"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutSuccess(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testput.db")
	kv := NewKV(dbPath)

	buf := strings.NewReader("testdata")
	req := httptest.NewRequest("PUT", "/test/put/data", buf)
	w := httptest.NewRecorder()
	kv.put(w, req)
	resp := w.Result()
	assert.Equal(t, 204, resp.StatusCode)
}

func TestGetSuccess(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testget.db")
	kv := NewKV(dbPath)

	data := "testdata"
	buf := strings.NewReader(data)
	req := httptest.NewRequest("PUT", "/test/get/data", buf)
	w := httptest.NewRecorder()
	kv.put(w, req)
	resp := w.Result()
	require.Equal(t, 204, resp.StatusCode)

	req = httptest.NewRequest("GET", "/test/get/data", nil)
	w = httptest.NewRecorder()
	kv.get(w, req)
	resp = w.Result()
	assert.Equal(t, 200, resp.StatusCode)
	body := make([]byte, len(data))
	_, err := resp.Body.Read(body)
	require.NoError(t, err)
	assert.Equal(t, data, string(body))
}

func TestGetNotExisting(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testget.db")
	kv := NewKV(dbPath)

	req := httptest.NewRequest("GET", "/test/get/data", nil)
	w := httptest.NewRecorder()
	kv.get(w, req)
	resp := w.Result()
	assert.Equal(t, 404, resp.StatusCode)
}

func TestDeleteSuccess(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testdelete.db")
	kv := NewKV(dbPath)

	buf := strings.NewReader("testdata")
	req := httptest.NewRequest("PUT", "/test/delete/data", buf)
	w := httptest.NewRecorder()
	kv.put(w, req)
	resp := w.Result()
	require.Equal(t, 204, resp.StatusCode)

	req = httptest.NewRequest("DELETE", "/test/delete/data", nil)
	w = httptest.NewRecorder()
	kv.delete(w, req)
	resp = w.Result()
	assert.Equal(t, 204, resp.StatusCode)

	req = httptest.NewRequest("GET", "/test/delete/data", nil)
	w = httptest.NewRecorder()
	kv.get(w, req)
	resp = w.Result()
	assert.Equal(t, 404, resp.StatusCode)
}

func TestDeleteNotExisting(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testdelete.db")
	kv := NewKV(dbPath)

	req := httptest.NewRequest("DELETE", "/test/delete/data", nil)
	w := httptest.NewRecorder()
	kv.get(w, req)
	resp := w.Result()
	assert.Equal(t, 404, resp.StatusCode)
}

func TestListSuccess(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "testlist.db")
	kv := NewKV(dbPath)
	count := 10
	for i := 0; i < count; i++ {
		data := fmt.Sprintf("testdata%d", i)
		buf := strings.NewReader(data)
		path := fmt.Sprintf("/test/list/data/%d", i)
		req := httptest.NewRequest("PUT", path, buf)
		w := httptest.NewRecorder()
		kv.put(w, req)
		resp := w.Result()
		require.Equal(t, 204, resp.StatusCode)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	kv.list(w, req)
	resp := w.Result()
	require.Equal(t, 200, resp.StatusCode)
	defer resp.Body.Close()
	s := bufio.NewScanner(resp.Body)
	var cnt int
	for s.Scan() {
		expected := fmt.Sprintf("list/data/%d=testdata%d", cnt, cnt)
		actual := s.Text()
		assert.Equal(t, expected, actual)
		cnt++
	}
	assert.Equal(t, count, cnt)
}
