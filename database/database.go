package database

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
)

type Migration struct {
	Sql string // sql to run
}

type MigrationCounter struct {
	Idx int
}

// Run migrations against the database
// panics on error
func MigrateDatabase(db *sqlx.DB, migrations *[]Migration) {
	lastIndex := lastMigrationCounter(db)
	if (lastIndex + 1) >= len(*migrations) {
		log.Println("No db migrations outstanding.")
	} else {
		log.Println("Running db migrations...")
		for index, migration := range *migrations {
			if index > lastIndex {
				runMigration(db, &migration, index)
				lastIndex = index
			}
		}
		log.Println("Db migrations completed ok")
	}
}

// Runs a migration, updates the migration table to reflect that it has been run
// panics on error
func runMigration(db *sqlx.DB, migration *Migration, idx int) {
	log.Println("Executing migration idx:", idx, "sql:", strings.Join(strings.Split(migration.Sql, "\n"), " "))
	db.MustExec(migration.Sql)
	log.Println("Updating migration_counter table, set idx:", idx)
	_, err := db.NamedExec("UPDATE migration_counter SET idx = :idx", &MigrationCounter{idx})
	if err != nil {
		panic(err)
	}
}

// Return the index of the last migration loaded, -1 if none ever loaded
func lastMigrationCounter(db *sqlx.DB) int {
	db.MustExec("CREATE TABLE IF NOT EXISTS migration_counter( idx integer )")

	var migration_index = MigrationCounter{-1}
	tx := db.MustBegin()
	err := tx.Get(&migration_index, "SELECT idx FROM migration_counter")
	switch {
	case err == sql.ErrNoRows:
		log.Println("Creating db migration tables")
		_, err = tx.NamedExec("INSERT INTO migration_counter ( idx ) VALUES ( :idx )", &migration_index)
		if err != nil {
			panic(err)
		}
	case err != nil:
		panic(err)
	default:
		// nil
	}
	tx.Commit()
	return migration_index.Idx
}
