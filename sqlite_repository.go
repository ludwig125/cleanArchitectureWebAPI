package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteRepository struct {
	db *sql.DB
}

// interfaceを実装しているか保証する
// See: http://golang.org/doc/faq#guarantee_satisfies_interface
var _ ActorRepository = (*sqliteRepository)(nil)

func NewSQLiteActorRepository(dbName string) (ActorRepository, error) {
	// 対象のDBがなくても新規に作ってしまうようなので、DBファイルの存在確認する
	if !exists(dbName) {
		return nil, fmt.Errorf("no such db file: %s", dbName)
	}

	db, err := connSQLite(dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connection db: %v", err)
	}
	log.Printf("connected %s successfully", dbName)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %v", err)
	}
	log.Printf("ping %s successfully", dbName)
	return &sqliteRepository{db: db}, nil
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func connSQLite(dbName string) (*sql.DB, error) {
	// DNS: root:password@tcp(ipaddress:port)/dbname
	// https://github.com/go-sql-driver/mysql#examples
	// パスワードなしで、localhostに対して、デフォルトの3306 portに接続する場合は以下でいい
	return sql.Open("sqlite3", dbName)
}

func (r *sqliteRepository) GetAll() ([]Actor, error) {
	rows, err := r.db.Query("SELECT * FROM actor")
	if err != nil {
		return nil, fmt.Errorf("failed to select all actors, err: %v", err)
	}
	defer rows.Close()

	var actors []Actor
	for rows.Next() {
		var a Actor
		err := rows.Scan(&a.ID, &a.Name, &a.Age)
		if err != nil {
			return nil, fmt.Errorf("failed to scan: %v", err)
		}
		actors = append(actors, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %v", err)
	}
	return actors, nil
}

func (r *sqliteRepository) FindByID(id int) ([]Actor, error) {
	rows, err := r.db.Query("SELECT * FROM actor WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %v", err)
	}
	return scanActors(rows)
}

func (r *sqliteRepository) FindByName(name string) ([]Actor, error) {
	rows, err := r.db.Query("SELECT * FROM actor WHERE name = ?", name)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %v", err)
	}
	return scanActors(rows)
}

func (r *sqliteRepository) FindByAge(age int) ([]Actor, error) {
	rows, err := r.db.Query("SELECT * FROM actor WHERE age = ?", age)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %v", err)
	}
	return scanActors(rows)
}

func scanActors(rows *sql.Rows) ([]Actor, error) {
	var actors []Actor
	defer rows.Close()
	for rows.Next() {
		var a Actor
		err := rows.Scan(&a.ID, &a.Name, &a.Age)
		if err != nil {
			return nil, fmt.Errorf("failed to scan: %v", err)
		}
		actors = append(actors, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %v", err)
	}
	return actors, nil
}

func (r *sqliteRepository) Update(a Actor) error {

	q := "INSERT OR REPLACE INTO actor(name, age) VALUES($1, $2);"
	res, err := r.db.Exec(q, a.Name, a.Age)
	if err != nil {
		return fmt.Errorf("failed to update db: %v", err)
	}
	// コマンドで影響を受けた件数が０ならエラーとする
	row, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get RowsAffected: %v", err)
	}
	if row == 0 {
		return errors.New("no row got affected")
	}
	return nil
}

func (r *sqliteRepository) DeleteByID(id int) error {
	q := "DELETE FROM actor WHERE id = $1;"
	res, err := r.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("failed to delete from db: %v", err)
	}

	// コマンドで影響を受けた件数が０ならエラーとする
	row, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get RowsAffected: %v", err)
	}
	if row == 0 {
		return errors.New("no row got affected")
	}
	return nil
}
