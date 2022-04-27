Set up folders
```
mkdir todos
cd todos
mkdir terraform
mkdir cli
```
Install cobra
`go install github.com/spf13/cobra-cli@latest`
init go module
```
cd cli
go mod init todos
```
Init cobra
```
cobra-cli init --viper
```
Add commands
```
cobra-cli add create
cobra-cli add config
cobra-cli add set -p 'configCmd'
```

update initConfig in root.go to initialize the viper config, remove the optional config file

```go
func initConfig() {
    // Find home directory.
    home, err := os.UserHomeDir()
    cobra.CheckErr(err)

    // Search config in home directory with name ".todos" (without extension).
    viper.AddConfigPath(home)
    viper.SetConfigType("yaml")
    viper.SetConfigName(".todos")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if err := viper.SafeWriteConfig(); err != nil {
			fmt.Fprintln(os.Stdout, "Error creating config file", err)
		}
	}
}
```

config set:

```go
	Run: func(cmd *cobra.Command, args []string) {
		viper.Set(args[0], args[1])
		viper.WriteConfig()
	},
```

config list:

```go
	Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Config")
        fmt.Println(viper.AllSettings())
    },
```

Install the squilte driver:

`go get github.com/mattn/go-sqlite3\`

create LocalConnection: cli/repository/sqliteRepository/LocalConnection.go

```go
package sqliteRepository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"todos/todo"
)

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
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
		return ErrDeleteFailed
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
```

add migration to root cmd `initConfig`:

```go
	dbPath := filepath.Join(home, ".todos.db")
	_, openDbFileErr := os.Open(dbPath)
	var doMigrate = false
	if errors.Is(openDbFileErr, os.ErrNotExist) {
		os.Create(dbPath)
		doMigrate = true
	}
	db, dbErr := sql.Open("sqlite3", dbPath)
	if dbErr != nil {
		log.Fatal(dbErr)
	}
	todosRepository = sqliteRepository.NewSQLiteRepository(db)
	if doMigrate {
		if migrateErr := todosRepository.Migrate(); migrateErr != nil {
			log.Fatal(migrateErr)
		}
	}
```

create Todo model

cli/todo/Todo.go

```go
package todo

type Todo struct {
	Id         string
	Title      string
	Body       string
	IsComplete bool
}
```

create add command: `cobra-cli add add`

add flag and demo:

```go
func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&todoDescription, "description", "d", "", "Optional description")
}

Run: func(cmd *cobra.Command, args []string) {
    desc, _ := cmd.Flags().GetString("description")
    var newTodo = todo.Todo{
        Title:      args[0],
        Body:       desc,
        IsComplete: false,
    }
    saved, err := todosRepository.Create(newTodo)
    if err != nil {
        fmt.Println("Error saving todo", err)
    } else {
        fmt.Println("Saved: ", saved)
    }
},
```

add list command: `cobra-cli add list`

```go
	Run: func(cmd *cobra.Command, args []string) {
		todos, _ := todosRepository.FetchAll()
		fmt.Println(todos)
	},
```

install color

go get github.com/fatih/color

add colors:

```go
	Run: func(cmd *cobra.Command, args []string) {
		todos, _ := todosRepository.FetchAll()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		blue := color.New(color.FgBlue).SprintFunc()
		magenta := color.New(color.FgMagenta).SprintFunc()
		for _, todo := range todos {
			var doneSymbol string
			var doneColor func(a ...interface{}) string
			if todo.IsComplete {
				doneSymbol = "✓"
				doneColor = green
			} else {
				doneSymbol = "×"
				doneColor = red
			}
			if len(todo.Body) == 0 {
				fmt.Printf(
					"%s  %s: %s.\n",
					doneColor(doneSymbol),
					yellow(todo.Id),
					magenta(todo.Title),
				)
			} else {
				fmt.Printf(
					"%s  %s: %s - %s.\n",
					doneColor(doneSymbol),
					yellow(todo.Id),
					magenta(todo.Title),
					blue(todo.Body),
				)
			}
		}
	},
```
build:

`go build -o todos main.go`


add delete command:

`cobra-cli add delete`
