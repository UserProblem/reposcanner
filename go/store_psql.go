package swagger

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

type PsqlDB struct {
	Host     string
	Port     int
	User     string
	Password string
	DBname   string
	DB       *sql.DB
}

var once sync.Once
var instance *PsqlDB

func GetPsqlDBInstance() *PsqlDB {
	once.Do(func() {
		instance = &PsqlDB{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "",
			DBname:   "reposcanner",
			DB:       nil,
		}
	})

	return instance
}

func (db *PsqlDB) Initialize() {
	if db.DB != nil {
		log.Fatal("Database already initialized.")
	}

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Password, db.DBname)

	if tmpDB, err := sql.Open("postgres", connectionString); err != nil {
		log.Fatal(err)
	} else {
		db.DB = tmpDB
	}
}
