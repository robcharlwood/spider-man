package cmd

import (
	"github.com/spf13/cobra"
)

var spidermanCmd = &cobra.Command{
	Use:   "spider-man",
	Short: "Your friendly neighbourhood web crawler",
	Long:  `spider-man is a web crawling CLI tool for Go.`,
}

func Execute() error {
	return spidermanCmd.Execute()
}

func init() {
	spidermanCmd.AddCommand(crawlCmd)
}
