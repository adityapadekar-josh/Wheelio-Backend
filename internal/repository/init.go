package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/config"
)

func InitDataStore(cfg config.Config) (*sql.DB, error) {
	dbInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name)

	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return db, err
	}

	return db, nil
}
