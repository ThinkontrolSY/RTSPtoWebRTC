package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	defaultPgHost = "postgres"
	defaultPgPort = 5432
	defaultPgUser = "postgres"
	defaultPgDb   = "smt"
	defaultStun   = "localhost:3478"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	go serveHTTP()
	go serveStreams()
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Println(sig)
		done <- true
	}()
	log.Println("Server Start Awaiting Signal")
	<-done
	log.Println("Exiting")
}

func getPgParam() (string, string, string, string, uint16) {
	var pgHost, user, password, dbname string
	var pgPort uint64
	if pgHost = os.Getenv("PGHOST"); pgHost == "" {
		pgHost = defaultPgHost
	}

	if pgPort, _ = strconv.ParseUint(os.Getenv("PGPORT"), 10, 64); pgPort == 0 {
		pgPort = defaultPgPort
	}

	if user = os.Getenv("PGUSER"); user == "" {
		user = defaultPgUser
	}

	password = os.Getenv("PGPASSWD")

	if dbname = os.Getenv("PGDATABASE"); dbname == "" {
		dbname = defaultPgDb
	}
	log.Printf("%v, %v, %v, %v, %v", user, password, dbname, pgHost, pgPort)
	return user, password, dbname, pgHost, uint16(pgPort)
}

func connectPg(user, password, dbname, pgHost string, pgPort uint16) (*sqlx.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, user, password, dbname)
	return sqlx.Connect("postgres", psqlInfo)
}
