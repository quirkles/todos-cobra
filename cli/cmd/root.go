/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"todos/repository/sqliteRepository"
)

var todosRepository *sqliteRepository.SQLiteRepository

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "todos",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, homeDirErr := os.UserHomeDir()
	cobra.CheckErr(homeDirErr)

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

	// initialize the db if required
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
}
