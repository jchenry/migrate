package migrate

import (
	"database/sql"
	"fmt"
	"time"
)

const table = "dbversion"

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

func Apply(ctx Context, d Dialect, migrations []Record) (err error) {
	if err = initialize(ctx, d); err == nil {
		var currentVersion int64
		if currentVersion, err = dbVersion(ctx, d); err == nil {
			migrations = migrations[currentVersion:] // only apply what hasnt been been applied already
			for i, m := range migrations {
				if err = apply(ctx, d, m); err != nil {
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

func initialize(ctx Context, d Dialect) (err error) {
	if noVersionTable(ctx, d) {
		return createVersionTable(ctx, d)
	}
	return
}

func noVersionTable(ctx Context, d Dialect) bool {
	rows, table_check := ctx.Query(d.TableExists(table))
	if rows != nil {
		defer rows.Close()
	}
	return table_check != nil
}

func apply(ctx Context, d Dialect, r Record) (err error) {
	if err = r.F(ctx); err == nil {
		err = incrementVersion(ctx, d, r.Description)
	}
	return
}

func createVersionTable(ctx Context, d Dialect) (err error) {
	_, err = ctx.Exec(d.CreateTable(table))
	return
}

func incrementVersion(ctx Context, d Dialect, description string) (err error) {
	_, err = ctx.Exec(d.InsertVersion(table), description, time.Now())
	return
}

func dbVersion(ctx Context, d Dialect) (id int64, err error) {
	row, err := ctx.Query(d.CheckVersion(table))
	if row.Next() {
		err = row.Scan(&id)
	}
	return
}
