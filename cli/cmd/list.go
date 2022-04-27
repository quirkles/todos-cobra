/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
}

func init() {
	rootCmd.AddCommand(listCmd)
}
