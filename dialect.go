package migrate

type Dialect interface {
	CreateTable(table string) string
	TableExists(table string) string
	CheckVersion(table string) string
	InsertVersion(table string) string
}

func Sqlite3() Dialect {
	return sqlite3{}
}

type sqlite3 struct{}

func (s sqlite3) CreateTable(table string) string {
	return "CREATE TABLE " + table + ` ( 
	id INTEGER PRIMARY KEY AUTOINCREMENT,  
	description VARCHAR, 
	applied TIMESTAMP);`
}

func (s sqlite3) TableExists(table string) string {
	return "SELECT * FROM " + table + ";"
}

func (s sqlite3) CheckVersion(table string) string {
	return "SELECT id FROM " + table + " ORDER BY id DESC LIMIT 0, 1;"
}

func (s sqlite3) InsertVersion(table string) string {
	return "INSERT INTO " + table + "(description, applied) VALUES (?,?);"
}
