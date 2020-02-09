package core

import (
	"database/sql"
	"fmt"
)

func Init(db *sql.DB) (err error) {
	ddls := []string{managersDDL, clientsDDL}
	err = execQueries(ddls, db)
	if err != nil {
		return err
	}

	initialData := []string{managersInitData}
	err = execQueries(initialData, db)
	if err != nil {
		return err
	}

	return nil
}

func execQueries(queries []string, db *sql.DB) (err error) {
	for _, query := range queries {
		_, err = db.Exec(query)
		if err != nil {
			return fmt.Errorf("can't execute db query '%v': %w", query, err)
		}
	}
	return nil
}
