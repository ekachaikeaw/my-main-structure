package main

import (
	"isekai-shop/internal/config"
	"isekai-shop/internal/databases/postgres"
	"isekai-shop/internal/http/rest"
)

func main() {
	cfg := config.New()
	db := postgres.NewPosgres(cfg.Database)
	server := rest.NewServer(cfg, db)

	server.Start()
}