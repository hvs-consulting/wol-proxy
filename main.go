package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ilyakaznacheev/cleanenv"
)

var config Config

//go:embed static
var embeddedContent embed.FS

//go:embed templates/index.html
var embeddedHTMLTemplate string

func main() {
	err := cleanenv.ReadConfig("config.yml", &config)
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/wake", wakeHandler).Methods("POST")
	router.HandleFunc(config.ServiceAliveRoute, apiHandler).Methods("GET")

	router.PathPrefix("/static").Handler(http.FileServer(http.FS(embeddedContent)))

	// go templating machinery is way to big so we just replace the two variables manually
	indexHtml := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(embeddedHTMLTemplate,
		"{{ .ServiceName }}", config.ServiceName),
		"{{ .WakeupTime }}", fmt.Sprint(config.ServerWakeupTime)),
		"{{ .AliveRoute }}", config.ServiceAliveRoute,
	)
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			w.Write([]byte(indexHtml))
		} else {
			// if the path of the URL is unknown we just redirect the browser to / so it loads the start button
			w.Header().Set("Location", "/")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	})

	log.Println("Starting server on " + config.Address)
	srv := http.Server{
		Addr:    config.Address,
		Handler: router,
	}
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// unexpected error. port in use? ...
		log.Fatalf("ListenAndServe() failed: %v", err)
	}
}

// this is called when the user clicks the button
func wakeHandler(w http.ResponseWriter, r *http.Request) {
	err := SendMagicPacket(config.MacAddr, config.BroadcastAddress+":9")
	if err != nil {
		log.Printf("failed to send magic packet: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// we send a TooEarly http status code so the client knows when to reload
func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusTooEarly)
}

type Config struct {
	Address           string `yaml:"listenAddress" env:"LISTEN_ADDRESS" env-default:"0.0.0.0:8080"`
	BroadcastAddress  string `yaml:"broadcastAddress" env-required:"BCAST_ADDRESS"`
	MacAddr           string `yaml:"macAddress" env-required:"MAC_ADDRESS"`
	ServiceName       string `yaml:"serviceName" env-required:"SERVICE_NAME"`
	ServerWakeupTime  uint   `yaml:"serverWakeupTime" env-required:"SERVER_WAKEUP_TIME"`
	ServiceAliveRoute string `yaml:"serviceAliveRoute" env-required:"SERVICE_ALIVE_ROUTE"`
}
