package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

type RootOptions struct {
    Region  string
    Profile string
}

var rootCmd = &cobra.Command{
    Use:   "awsh",
    Short: "Awsh is aws resource check helper.",
    Long:  "We want to make operational work easier. For example, check log, run build project, copy file from s3...etc.",
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        _, err = fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    rootCmd.InitDefaultVersionFlag()

    rootCmd.PersistentFlags().String("profile", "default", "Input profile")
    rootCmd.PersistentFlags().String("region", "ap-northeast-1", "Input region")
}
