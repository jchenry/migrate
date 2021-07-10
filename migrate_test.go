package migrate

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestHelperFuncs(t *testing.T) {
	path, db, err := createTestDB()
	if err != nil {
		t.Fail()
	}
	if err = teardownTestDB(path, db); err != nil {
		t.Fail()
	}
}

func TestCreateVersionTable(t *testing.T) {
	path, db, err := createTestDB()
	if err != nil {
		t.Fail()
	}

	err = createVersionTable(db, Sqlite3())
	if err != nil {
		t.Fatal(err)
	}

	if err = teardownTestDB(path, db); err != nil {
		t.Fail()
	}
}

func TestIncrementVersion(t *testing.T) {
	path, db, err := createTestDB()
	if err != nil {
		t.Fail()
	}

	sl3 := Sqlite3()

	err = createVersionTable(db, sl3)
	if err != nil {
		t.Fatal(err)
	}

	descriptions := []string{
		"this is a test",
		"this is another test",
	}

	for _, d := range descriptions {
		err = incrementVersion(db, sl3, d)
		if err != nil {
			t.Fatal(err)
		}
	}

	rows, err := db.Query("SELECT id, description from dbversion")
	if err != nil {
		t.Fatal(err)
	}
	var id int
	var description string
	for r := 1; rows.Next(); r++ {
		err = rows.Scan(&id, &description)
		if err != nil {
			t.Fatal(err)
		}
		if id != r || !strings.EqualFold(description, descriptions[r-1]) {
			t.Fatalf("first row does not match %d %s: %d %s", id, descriptions[r-1], r, description)
		}

	}

	if err = teardownTestDB(path, db); err != nil {
		t.Fail()
	}
}

func TestDbVersion(t *testing.T) {
	path, db, err := createTestDB()
	if err != nil {
		t.Fail()
	}

	sl3 := Sqlite3()

	err = createVersionTable(db, sl3)
	if err != nil {
		t.Fatal(err)
	}

	ver, err := dbVersion(db, sl3)
	if ver != 0 || err != nil {
		t.Fatalf("version not 0 as expected (actual %d) or err: %#v", ver, err)
	}

	err = incrementVersion(db, sl3, "Test 1")
	ver, err = dbVersion(db, sl3)
	if ver != 1 {
		t.Fatalf("version not 1 as expected (actual %d)", ver)
	}
	if err != nil {
		t.Fatalf("err on dbversion of first increment: %#v", err)
	}

	// err = incrementVersion(db, d)

	if err = teardownTestDB(path, db); err != nil {
		t.Fail()
	}
}

func TestApply(t *testing.T) {
	path, db, err := createTestDB()
	if err != nil {
		t.Fail()
	}

	sl3 := Sqlite3()

	records :=
		[]Record{
			{
				Description: "create people table",
				F: func(ctx Context) (err error) {
					_, err = ctx.Exec(`
				CREATE TABLE people (
					given_name VARCHAR(20),
					surname VARCHAR(30),
					sex CHAR(1),
					age SMALLINT);
				`)
					return
				},
			},
			{
				Description: "Insert a person into people",
				F: func(ctx Context) (err error) {
					_, err = ctx.Exec(`INSERT INTO people VALUES('Henry','Colin','M', 42)`)
					return
				},
			},
		}

	err = Apply(db, sl3, records)

	if err != nil {
		t.Fatal(err)
	}

	r := db.QueryRow("SELECT given_name FROM people")

	var given_name string
	r.Scan(&given_name)

	if given_name != "Henry" {
		t.Fatalf("second migration did not complete: %s != %s", given_name, "Henry")
	}

	// reapply and make sure we dont re-run anything
	err = Apply(db, sl3, records)
	ver, err := dbVersion(db, sl3)
	if ver != 2 {
		t.Fatalf("version not 2 as expected (actual %d)", ver)
	}
	if err != nil {
		t.Fatalf("err on dbversion of re-apply: %#v", err)
	}

	// add bad (causes migrate.Error) case here.

	ishouldntHideUserErrors := errors.New("I should fail")

	records = append(records, Record{
		Description: "Insert a person into people",
		F: func(ctx Context) (err error) {
			return ishouldntHideUserErrors
		},
	})

	err = Apply(db, sl3, records)

	if errors.Unwrap(err) != ishouldntHideUserErrors {
		t.Fatalf("unexpected error returned that should have been record function error: %#v", err)
	}
	ver, err = dbVersion(db, sl3)
	if ver != 2 {
		t.Fatalf("version not 2 as expected (actual %d) after bad record apply", ver)
	}
	if err != nil {
		t.Fatalf("err on dbversion of re-apply with bad record: %#v", err)
	}

	if err = teardownTestDB(path, db); err != nil {
		t.Fail()
	}

}

func createTestDB() (path string, db *sql.DB, err error) {
	if f, err := ioutil.TempFile(os.TempDir(), "migrate-test-db"); err == nil {
		f.Close()
		if db, err := sql.Open("sqlite3", f.Name()); err == nil {
			return f.Name(), db, err
		}
	}
	return
}
func teardownTestDB(path string, db *sql.DB) (err error) {
	if err = db.Close(); err == nil {
		err = os.Remove(path)
	}
	return
}
