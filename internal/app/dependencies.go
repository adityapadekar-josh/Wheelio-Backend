package app

import "database/sql"

type Dependencies struct{}

func NewServices(db *sql.DB) Dependencies {
	return Dependencies{}
}
