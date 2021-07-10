module github.com/jchenry/migrate

go 1.16

require github.com/mattn/go-sqlite3 v1.14.7

retract (
    v0.0.1 // Published accidentally.
    v1.0.2 // Contains retractions only.
)
