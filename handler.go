package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	chi "github.com/go-chi/chi/v5"
)

type Handler struct {
	router chi.Router
	log    *log.Logger
	kv     KV
	dbPath string
}

func NewHandler(kv KV) *Handler {
	h := Handler{
		kv:     kv,
		log:    log.New(os.Stdout, "[KV] ", log.LstdFlags),
		router: chi.NewRouter(),
	}
	h.router.Get("/{bucket}/{key}", h.get)
	h.router.Put("/{bucket}/{key}", h.put)
	h.router.Delete("/{bucket}/{key}", h.delete)
	return &h
}

func (kv *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	kv.router.ServeHTTP(w, r)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	out, err := h.kv.Get([]byte(r.URL.Path))
	if err != nil {
		httpErr(h.log, w, err)
		return
	}
	w.Write(out)
}

func (h *Handler) put(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Println(err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	err = h.kv.Put([]byte(r.URL.Path), body)
	if err != nil {
		h.log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	err := h.kv.Delete([]byte(r.URL.Path))
	if err != nil {
		httpErr(h.log, w, err)
		return
	}
	w.WriteHeader(204)
}

func httpErr(log *log.Logger, w http.ResponseWriter, err error) {
	if err != nil && errors.Is(err, ErrNotFound) {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
}
