package main

import (
	"flag"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/davidoram/ufacility/database"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
)

/*
 * Http request handler methods
 */
func pingHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong\n")
}
func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic(nil)
}

func main() {

	var (
		dbHost     = flag.String("host", "localhost", "Postgress db host")
		dbName     = flag.String("db", "rewards_development", "Postgress db name")
		dbUser     = flag.String("user", os.Getenv("USER"), "Postgress db user")
		dbPassword = flag.String("password", "", "Postgress db password")
		httpPort   = flag.Int("port", 8080, "HTTP port to listen on")
	)
	flag.Parse()

	log.Println("ufacility started")

	log.Printf("Connecting to database: %v, host: %v, user: %v ...", *dbName, *dbHost, *dbUser)
	connectionString := fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", *dbUser, *dbPassword, *dbHost, *dbName)
	db := sqlx.MustConnect("postgres", connectionString)

	/*
	 * Put your database migrations here, will be processed in order
	 */
	migrations := []database.Migration{
		database.Migration{
			`CREATE TABLE rewards(
				id 						integer,
				permalink 		varchar(255) not null,
				name					varchar(255) not null,

				PRIMARY KEY(id)
			)
			`,
		},
	}
	database.MigrateDatabase(db, &migrations)

	/*
	 * Setup the web server routes
	 */
	router := mux.NewRouter()
	router.HandleFunc("/ping", pingHandler)
	router.HandleFunc("/panic", panicHandler)

	n := negroni.Classic()
	// router goes last
	n.UseHandler(router)

	n.Run(fmt.Sprintf(":%d", *httpPort))
}
