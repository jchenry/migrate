package migrate

// migrate is a package for SQL datbase migrations in the spirit of dbstore(rsc.io/dbstore)
// it is intended to keep its footprint small, requiring only an addiutional table in the database
// there is no rollback support as you should only ever roll forward.
// uses SQL99 compatible SQL only.
