/*
 * Repository Secrets Scanner
 *
 * This is a simple backend API to allow a user to configure repositories for scanning, trigger a scan of those repositories, and retrieve the results.
 *
 * API version: 0.0.1
 * Contact: sean.critica@gmail.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	sw "github.com/UserProblem/reposcanner/go"
	"github.com/joho/godotenv"
)

func main() {
	log.Printf("Server started")
	godotenv.Load()

	var app sw.App

	loadDBParameters(&app)

	app.Initialize(loadNoop())
	app.Run()

	err := http.ListenAndServe(":8080", app.Router)

	app.CleanUp()
	log.Fatal(err.Error())
}

func loadDBParameters(app *sw.App) {
	app.DBType = os.Getenv("DATABASE_TYPE")
	log.Printf("Using database type '%v'", app.DBType)

	if app.DBType == "postgresql" {
		app.DB = sw.GetPsqlDBInstance()
		app.DB.Host = os.Getenv("DATABASE_HOST")
		if pnum, err := strconv.Atoi(os.Getenv("DATABASE_PORT")); err != nil {
			log.Fatal(err.Error())
		} else {
			app.DB.Port = pnum
		}
		app.DB.User = os.Getenv("DATABASE_USER")
		app.DB.Password = os.Getenv("DATABASE_PASSWORD")
		app.DB.DBname = os.Getenv("DATABASE_NAME")
		app.DB.Initialize()
	}
}

func loadNoop() bool {
	if noop := os.Getenv("ENGINE_NOOP"); noop == "1" {
		log.Printf("Running with no-op scanner.")
		return true
	}
	return false
}
