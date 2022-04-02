package main

import (
	"context"
	"flag"

	chi "github.com/go-chi/chi/v5"
	"github.com/rkorkosz/web"
)

func main() {
	dbPath := flag.String("dbPath", "kv.db", "database path")
	flag.Parse()
	kv := NewKV(*dbPath)
	r := chi.NewRouter()
	kv.Router(r)
	srv := web.Server(
		web.WithAddr(":8000"),
		web.WithTLSConfig(web.LocalTLSConfig("cert.pem", "key.pem")),
		web.WithHandler(r),
	)
	web.RunServer(context.Background(), srv)
}
