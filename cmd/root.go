/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "flex-insights-cli",
	Short: "Helper program for interacting with the Flex Insights API.",
	Long: `This is a CLI tool for interacting with the Flex Insights API. It provides a wrapper around the Flex Insights API with the aim to simplify its use.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.flex-insights-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}


