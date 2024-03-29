package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Nunoki/onetimesharer/internal/pkg/config"
	"github.com/Nunoki/onetimesharer/internal/pkg/crypter"
	"github.com/Nunoki/onetimesharer/internal/pkg/filestorage"
	"github.com/Nunoki/onetimesharer/internal/pkg/randomizer"
	"github.com/Nunoki/onetimesharer/internal/pkg/server"
	"github.com/Nunoki/onetimesharer/internal/pkg/sqlite"
	"github.com/Nunoki/onetimesharer/pkg/aescfb"
)

const defaultPortHTTP uint = 8000
const defaultPortHTTPS uint = 443
const defaultPayloadLimitBytes = 5000

var (
	ctx = context.Background()
)

// main gets the configuration and all required services, then starts the server, and registers a
// handler for the kill signals to perform a graceful shutdown
func main() {
	conf := configuration()

	encrypter := aescfb.New(encryptionKey())
	store := store(encrypter, conf)

	server := server.New(conf, store)

	// Perform graceful shutdown when interrupted from shell
	go func() {
		fmt.Fprintf(os.Stdout, "Listening on port %d\n", *conf.Port)
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	if err := server.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed:%+v", err)
	}
	log.Print("Server exited properly")
}

// encryptionKey returns the encryption key for the encrypter, which is read from an environment
// variable, or generated anew if none is present. If an encryption key is provided, but not valid
// (incorrect length), an error is output to stderr and the program exits.
func encryptionKey() (key string) {
	key = os.Getenv("OTS_ENCRYPTION_KEY")
	if len(key) == 0 {
		key = randomizer.String(32)
		fmt.Fprintf(
			os.Stdout,
			"Generated encryption key is: %s (set env variable OTS_ENCRYPTION_KEY to use custom key)\n",
			key,
		)
		return
	}

	if len(key) != 32 {
		fmt.Fprintf(
			os.Stderr,
			"Provided encryption key must be 32 characters long, is %d\n",
			len(key),
		)
		os.Exit(1)
	}

	return
}

// configuration processes passed arguments and sets up variables appropriately. If a conflict
// occurs with flag configuration, an error is being output to stderr, and the program exits.
func configuration() config.Config {
	conf := config.Config{}

	conf.Certfile = flag.String(
		"certfile",
		"",
		"Path to certificate file, required when running on HTTPS",
	)
	conf.HTTPS = flag.Bool(
		"https",
		false,
		"Whether to run on HTTPS (requires --certfile and --keyfile)",
	)
	conf.JSONFile = flag.Bool(
		"json",
		false,
		"Use a JSON file as storage instead of the default SQLite database",
	)
	conf.Keyfile = flag.String(
		"keyfile",
		"",
		"Path to key file, required when running on HTTPS",
	)
	conf.PayloadLimit = flag.Uint(
		"payload",
		defaultPayloadLimitBytes,
		"Limit on number of bytes allowed to be posted in the request body, to prevent large payload attacks. Note this includes the entire JSON message, not just the contents of the secret.",
	)
	conf.Port = flag.Uint(
		"port",
		0,
		fmt.Sprintf(
			"Port to run on (default %d for HTTP, %d for HTTPS)",
			defaultPortHTTP,
			defaultPortHTTPS,
		),
	)
	flag.Parse()

	if *conf.HTTPS && (*conf.Certfile == "" || *conf.Keyfile == "") {
		log.Fatal("running on HTTPS requires the certification file and key file (see --help)")
	}

	if *conf.Port == 0 {
		if *conf.HTTPS {
			*conf.Port = defaultPortHTTPS
		} else {
			*conf.Port = defaultPortHTTP
		}
	}

	return conf
}

// store returns an instance of a store for the secrets. It will use the configuration to determine
// which store will be used. If an error occurs on store initialization, an error is output to
// stderr and the program exits.
func store(encrypter crypter.Crypter, conf config.Config) server.Storer {
	var store server.Storer
	var err error
	if *conf.JSONFile {
		store, err = filestorage.New(encrypter)
	} else {
		store, err = sqlite.New(ctx, encrypter)
	}
	if err != nil {
		log.Fatal(err)
	}
	return store
}
