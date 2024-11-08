package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/c0rlyy/hermis/internal/broker"
	"github.com/c0rlyy/hermis/internal/config"
	"github.com/c0rlyy/hermis/internal/database"
	"github.com/c0rlyy/hermis/internal/utils"
)

type Server struct {
	port       int
	hashParams utils.Params
	db         *database.MongoDb
	cfg        *config.EnvContents
	mb         *broker.MessageBroker
}

// creates goalng type http server using echo
func NewServer() *http.Server {
	cfg := config.GetEnv()
	port, _ := strconv.Atoi(cfg.Port)

	db := database.NewDb(cfg)
	if err := db.Health(); err != nil {
		log.Fatal("error with database")
	}

	NewServer := &Server{
		port:       port,
		hashParams: utils.NewParams(),
		db:         db,
		cfg:        &cfg,
		mb:         broker.NewMessageBroker(),
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
