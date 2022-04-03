package main

import (
	"context"
	"crypto/tls"
	"flag"

	"github.com/rkorkosz/web"
)

func main() {
	dbPath := flag.String("dbPath", "kv.db", "database path")
	email := flag.String("email", "", "acme email")
	host := flag.String("host", "", "acme host")
	cert := flag.String("cert", "cert.pem", "certificate path")
	key := flag.String("key", "key.pem", "certificate key path")
	addr := flag.String("bind", ":8000", "bind address")
	flag.Parse()
	var tlsConfig *tls.Config
	if *email != "" && *host != "" {
		tlsConfig = web.AutoCertWhitelist(*email, *host)
	} else {
		tlsConfig = web.LocalTLSConfig(*cert, *key)
	}
	srv := web.Server(
		web.WithAddr(*addr),
		web.WithTLSConfig(tlsConfig),
		web.WithHandler(NewHandler(NewBoltKV(*dbPath))),
	)
	web.RunServer(context.Background(), srv)
}
