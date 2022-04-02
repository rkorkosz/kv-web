package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	chi "github.com/go-chi/chi/v5"
	bolt "go.etcd.io/bbolt"
)

var ErrNotFound = errors.New("not found")

type KV struct {
	opts   *bolt.Options
	log    *log.Logger
	dbPath string
}

func NewKV(dbPath string) *KV {
	db, err := bolt.Open(dbPath, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	return &KV{dbPath: dbPath, log: log.New(os.Stdout, "[KV] ", log.LstdFlags)}
}

func (kv *KV) Router(r chi.Router) {
	r.Get("/{bucket}", kv.list)
	r.Put("/{bucket}/{key}", kv.put)
	r.Get("/{bucket}/{key}", kv.get)
}

func getBucketAndKey(s string) (string, string, error) {
	chunks := strings.SplitN(s, "/", 3)
	if len(chunks) < 3 {
		return "", "", errors.New("invalid")
	}
	return chunks[1], chunks[2], nil
}

func getBucket(s string) (string, error) {
	bucket := s[1:]
	if bucket == "" {
		return "", errors.New("invalid")
	}
	return bucket, nil
}

func (kv *KV) list(w http.ResponseWriter, r *http.Request) {
	db, err := bolt.Open(kv.dbPath, 0666, &bolt.Options{ReadOnly: true})
	if err != nil {
		kv.log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer db.Close()
	bucket, err := getBucket(r.URL.Path)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return errors.New("not found")
		}
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			w.Write(k)
			w.Write([]byte("="))
			w.Write(v)
			_, err = w.Write([]byte("\n"))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		kv.log.Println(err)
		http.Error(w, http.StatusText(404), 404)
		return
	}
}

func (kv *KV) get(w http.ResponseWriter, r *http.Request) {
	db, err := bolt.Open(kv.dbPath, 0666, &bolt.Options{ReadOnly: true})
	if err != nil {
		kv.log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer db.Close()
	bucket, key, err := getBucketAndKey(r.URL.Path)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return ErrNotFound
		}
		val := bucket.Get([]byte(key))
		if val == nil {
			return ErrNotFound
		}
		_, err = w.Write(val)
		return err
	})
	if err != nil {
		kv.log.Println(err)
		http.Error(w, http.StatusText(404), 404)
		return
	}
}

func (kv *KV) put(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	db, err := bolt.Open(kv.dbPath, 0666, kv.opts)
	if err != nil {
		kv.log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer db.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		kv.log.Println(err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	bucket, key, err := getBucketAndKey(r.URL.Path)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(key), body)
	})
	if err != nil {
		kv.log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	w.WriteHeader(204)
}

func (kv *KV) delete(w http.ResponseWriter, r *http.Request) {
	db, err := bolt.Open(kv.dbPath, 0666, kv.opts)
	if err != nil {
		kv.log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer db.Close()

	bucket, key, err := getBucketAndKey(r.URL.Path)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return ErrNotFound
		}
		return bucket.Delete([]byte(key))
	})
	if errors.Is(err, ErrNotFound) {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	if err != nil {
		kv.log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	w.WriteHeader(204)
}
