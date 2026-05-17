package main

import (
	"fmt"
	"net/http"
	"time"
	"unitly/bootstrap"
	"unitly/internal/db"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	app := bootstrap.CreateAppliaction()
	dbAddr := app.GetDatabaseURL()
	db, err := db.NewPostgresDB(
		dbAddr,
		25,
		10,
		1*time.Hour,
		30*time.Minute,
		1*time.Minute,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	mux := http.NewServeMux()
	app.Run(mux)
}
