package main

import (
	"bytes"
	"errors"
	"log"

	bolt "go.etcd.io/bbolt"
)

var ErrNotFound = errors.New("not found")

type KV interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
}

type BoltKV struct {
	dbPath string
}

func NewBoltKV(dbPath string) *BoltKV {
	db, err := bolt.Open(dbPath, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	return &BoltKV{dbPath: dbPath}
}

func (boltkv *BoltKV) Get(key []byte) ([]byte, error) {
	db, err := bolt.Open(boltkv.dbPath, 0666, &bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	bucket, key, err := getBucketAndKey(key)
	if err != nil {
		return nil, err
	}
	var val []byte
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return ErrNotFound
		}
		val = bucket.Get([]byte(key))
		if val == nil {
			return ErrNotFound
		}
		return err
	})
	return val, err
}

func (boltkv *BoltKV) Put(key []byte, value []byte) error {
	db, err := bolt.Open(boltkv.dbPath, 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	bucket, key, err := getBucketAndKey(key)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(key), value)
	})
}

func (boltkv *BoltKV) Delete(key []byte) error {
	db, err := bolt.Open(boltkv.dbPath, 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	bucket, key, err := getBucketAndKey(key)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return ErrNotFound
		}
		return bucket.Delete([]byte(key))
	})
}

func getBucketAndKey(s []byte) ([]byte, []byte, error) {
	chunks := bytes.SplitN(s, []byte("/"), 3)
	if len(chunks) < 3 {
		return nil, nil, errors.New("invalid")
	}
	return chunks[1], chunks[2], nil
}

func getBucket(s []byte) ([]byte, error) {
	if len(s) < 2 {
		return nil, errors.New("invalid")
	}
	bucket := s[1:]
	return bucket, nil
}
