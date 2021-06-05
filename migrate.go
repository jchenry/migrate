package migrate

import (
	"database/sql"
	"fmt"
	"time"
)

const table = "dbversion"

var tableCreateSql = "CREATE TABLE " + table + ` ( 
	id INTEGER PRIMARY KEY AUTOINCREMENT,  
	description VARCHAR, 
	applied TIMESTAMP
);`
var tableCheckSql = "SELECT * FROM " + table + ";"
var versionCheckSql = "SELECT id FROM " + table + " ORDER BY id DESC LIMIT 0, 1;"
var versionInsertSql = "INSERT INTO " + table + "(description, applied) VALUES (?,?);"

type Error struct {
	description string
	wrapped     error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %v", e.description, e.wrapped)
}

func (e Error) Unwrap() error {
	return e.wrapped
}

type Record struct {
	Description string
	F           func(ctx Context) error
}

type Context interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func Apply(ctx Context, migrations []Record) (err error) {
	if err = initialize(ctx); err == nil {
		var currentVersion int64
		if currentVersion, err = dbVersion(ctx); err == nil {
			migrations = migrations[currentVersion:] // only apply what hasnt been been applied already
			for i, m := range migrations {
				if err = apply(ctx, m); err != nil {
					err = Error{
						description: fmt.Sprintf("error performing migration \"%s\"", migrations[i].Description),
						wrapped:     err,
					}
					break
				}
			}
		}
	}
	return
}

func initialize(ctx Context) (err error) {
	if noVersionTable(ctx) {
		return createVersionTable(ctx)
	}
	return nil
}

func noVersionTable(ctx Context) bool {
	rows, table_check := ctx.Query(tableCheckSql)
	if rows != nil {
		defer rows.Close()
	}
	return table_check != nil
}

func apply(ctx Context, r Record) (err error) {
	if err = r.F(ctx); err == nil {
		err = incrementVersion(ctx, r.Description)
	}
	return
}

func createVersionTable(ctx Context) (err error) {
	_, err = ctx.Exec(tableCreateSql)
	return
}

func incrementVersion(ctx Context, description string) (err error) {
	_, err = ctx.Exec(versionInsertSql, description, time.Now())
	return
}

func dbVersion(ctx Context) (id int64, err error) {
	row, err := ctx.Query(versionCheckSql)
	if row.Next() {
		err = row.Scan(&id)
	}
	return
}
