package sqliteRepository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"todos/todo"
)

var (
	ErrDuplicate       = errors.New("record already exists")
	ErrNotExists       = errors.New("row not exists")
	ErrUpdateFailed    = errors.New("update failed")
	ErrNothingToDelete = errors.New("nothing was found to delete with that id")
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (r *SQLiteRepository) Migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS todos(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        body TEXT,
        is_complete INTEGER NOT NULL check (is_complete in (0, 1)) default 0
    );
    `

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) Create(todo todo.Todo) (*todo.Todo, error) {
	res, err := r.db.Exec("INSERT INTO todos(title, body, is_complete) values(?,?,?)", todo.Title, todo.Body, todo.IsComplete)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return nil, ErrDuplicate
			}
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	todo.Id = string(id)

	return &todo, nil
}

func (r *SQLiteRepository) Update(id int64, updated todo.Todo) (*todo.Todo, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec("UPDATE todos SET title = ?, body = ?, is_complete = ? WHERE id = ?", updated.Title, updated.Body, updated.IsComplete, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUpdateFailed
	}

	return &updated, nil
}

func (r *SQLiteRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNothingToDelete
	}

	return err
}

func (r *SQLiteRepository) FetchAll() ([]todo.Todo, error) {
	rows, queryErr := r.db.Query("SELECT * from todos")
	if queryErr != nil {
		fmt.Println("queryErr", queryErr)
		return nil, queryErr
	}

	defer rows.Close()
	var result []todo.Todo
	for rows.Next() {
		item := todo.Todo{}
		scanErr := rows.Scan(&item.Id, &item.Title, &item.Body, &item.IsComplete)
		if scanErr != nil {
			fmt.Println("scanErr", scanErr)
			return nil, scanErr
		}
		result = append(result, item)
	}

	return result, nil
}
