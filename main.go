package main

import (
	"flag"
	"fmt"
	"github.com/davidoram/ufacility/database"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong\n")
}

func main() {

	var (
		dbHost = flag.String("host",
			"localhost",
			"Postgress db host")
		dbName = flag.String("db",
			"rewards_development",
			"Postgress db name")
		dbUser = flag.String("user",
			os.Getenv("USER"),
			"Postgress db user")
		dbPassword = flag.String("password",
			"",
			"Postgress db password")
	)
	flag.Parse()

	log.Println("ufacility started")

	log.Printf("Connecting to database: %v, host: %v, user: %v ...", *dbName, *dbHost, *dbUser)
	connectionString := fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", *dbUser, *dbPassword, *dbHost, *dbName)
	db := sqlx.MustConnect("postgres", connectionString)

	migrations := []database.Migration{
		database.Migration{
			`create table rewards(
				id int
			)
			`,
		},
	}
	database.MigrateDatabase(db, &migrations)

	// // Incoming requests to your API go under /api/{version}
	// http.HandleFunc("/api/v1/rewards", version1.RewardsHandler)
	// // Requests for outgoing events go under /atom
	// //http.HandleFunc("/atom/v1/rewards", AtomV1Handler)
	// // Requests for monitoring the service itself go under /monitor
	// http.HandleFunc("/ping", PingHandler)
	// //http.HandleFunc("/performance", PerformanceHandler)

	// err := http.ListenAndServe(":12345", nil)
	// if err != nil {
	//   log.Fatal("ListenAndServe: ", err)
	// }
}
