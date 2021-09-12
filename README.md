# migrate

`migrate` is a package for SQL datbase migrations in the spirit of dbstore(rsc.io/dbstore) it is intended to keep its footprint small, requiring only an additional table in the database there is no rollback support as you should only ever roll forward. Sqlite3 support is provided, support for other datbases can be added by implementing the `Dialect` interface

## Installation

```bash
go get github.com/jchenry/migrate
```

## Usage

```go
...
records :=
		[]Record{
			{
				Description: "create people table",
				F: func(ctx Context) (err error) {
					_, err = ctx.Exec(`
				CREATE TABLE people (
					given_name VARCHAR(20),
					surname VARCHAR(30),
					gender CHAR(1),
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

	err = migrate.Apply(db, migrate.Sqlite3(), records)
    ...

```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)

courtesey of https://www.makeareadme.com
