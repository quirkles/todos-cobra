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

delete command:

```go
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Printf("Delete command needs an integer argument for the id to delete\n")
			return
		}
		idStr := args[0]
		id, parseInputErr := strconv.ParseInt(idStr, 10, 64)
		if parseInputErr != nil {
			fmt.Printf("Delete command needs an integer argument for the id, got: %s\n", idStr)
			return
		}
		deleteErr := todosRepository.Delete(id)
		if deleteErr != nil {
			fmt.Println("Error deleting todo:", deleteErr)
		}
	},
```
