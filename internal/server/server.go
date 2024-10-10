package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/c0rlyy/hermis/internal/config"
	"github.com/c0rlyy/hermis/internal/database"
	"github.com/c0rlyy/hermis/internal/utils"
)

type Server struct {
	port int

	hashParams utils.Params

	db *database.MongoDb

	cfg *config.EnvContents
}

// creates goalng type http server using echo
func NewServer() *http.Server {
	cfg := config.GetEnv()
	port, _ := strconv.Atoi(cfg.Port)

	NewServer := &Server{
		port:       port,
		hashParams: utils.NewParams(),
		db:         database.NewDb(cfg),
		cfg:        &cfg,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf("localhost:%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
