package mysql

import "database/sql"

// SELECT
func Read(queryStr string, args ...any) (rows *sql.Rows, err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()

	return db.Query(queryStr, args...)
}

// INSERT, UPDATE, DELETE
func Write(queryStr string, args ...any) (result sql.Result, err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()

	return db.Exec(queryStr, args...)
}

// SELECT
func TxRead(tx *sql.Tx, queryStr string, args ...any) (rows *sql.Rows, err error) {
	return tx.Query(queryStr, args...)
}

// INSERT, UPDATE, DELETE
func TxWrite(tx *sql.Tx, queryStr string, args ...any) (result sql.Result, err error) {
	return tx.Exec(queryStr, args...)
}
