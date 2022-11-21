package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dnflash/demo-p1-go-user-management-service/internal/database"
	"github.com/dnflash/demo-p1-go-user-management-service/internal/server"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type config struct {
	userDBURI            string
	serverAddress        string
	rawAccessTokenSecret string
}

func main() {
	log.Printf("Go Demo - User Management Service")
	appContext := context.Background()

	c, err := getConfig()
	if err != nil {
		log.Printf("Get config error: %v", err)
		return
	}

	accessTokenSecret, err := jwk.FromRaw([]byte(c.rawAccessTokenSecret))
	if err != nil {
		log.Printf("Failed to create access token secret key")
		return
	}

	userDBConn, err := database.ConnectUserDB(appContext, c.userDBURI)
	if err != nil {
		log.Printf("Error connecting to UserDB at %s", c.userDBURI)
		return
	}
	defer func() {
		if err := userDBConn.Disconnect(appContext); err != nil {
			log.Printf("Error disconnecting from UserDB: %v", err)
		}
	}()

	srv := server.Server{
		UserDB:            database.UserDatabase{Database: userDBConn.Database(database.UserDB)},
		AccessTokenSecret: accessTokenSecret,
	}

	httpSrv := &http.Server{
		Addr:           c.serverAddress,
		Handler:        srv.Router(),
		WriteTimeout:   15 * time.Second,
		ReadTimeout:    15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1024,
	}

	errChan := make(chan error, 1)
	go func() {
		log.Printf("Serving on %s", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Listen and serve error: %v", err)
			errChan <- err
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	select {
	case <-sigChan:
		if err := httpSrv.Shutdown(appContext); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		log.Printf("Server shutdown")
	case <-errChan:
	}
}

func getConfig() (config, error) {
	c := config{}
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			log.Printf("config.yaml file not found")
			log.Printf("Reading config from environment variables, prefix: APP_")
			viper.SetEnvPrefix("APP")
			viper.AutomaticEnv()
		} else {
			return c, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		log.Printf("Reading config from config.yaml")
	}
	var missingConfig []string
	c.userDBURI = viper.GetString("userDb")
	if c.userDBURI == "" {
		missingConfig = append(missingConfig, "userDb")
	}
	c.serverAddress = viper.GetString("serverAddress")
	if c.serverAddress == "" {
		missingConfig = append(missingConfig, "serverAddress")
	}
	c.rawAccessTokenSecret = viper.GetString("accessTokenSecret")
	if c.rawAccessTokenSecret == "" {
		missingConfig = append(missingConfig, "accessTokenSecret")
	}
	if len(missingConfig) > 0 {
		return c, fmt.Errorf("missing config: %v", missingConfig)
	}
	return c, nil
}
