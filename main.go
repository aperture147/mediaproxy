package main

import (
	"context"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"mediaproxy/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		okJson := "{\"status\": \"ok\"}"
		fmt.Fprint(w, okJson)
	})
	r.PathPrefix("/image").Handler(server.NewImageRouter(ctx, 10))

	allowedHeaders := handlers.AllowedHeaders([]string{"Authorization"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT"})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handlers.CORS(allowedHeaders, allowedMethods)(r),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln("listen:", err)
		}
	}()

	log.Println("server started, listening to 8080 port")

	<-done
	log.Print("stopping server")
	stopCtx, cancel2 := context.WithTimeout(ctx, 30*time.Second)
	defer cancel2()

	if err := srv.Shutdown(stopCtx); err != nil {
		log.Println("server was not gracefully shutdown, terminated")
	}

	log.Println("server was gracefully stopped")
}
