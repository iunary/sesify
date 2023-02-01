package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sesify",
	Short: "Simple email sender",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	PrintBanner()
	if err := godotenv.Load(); err != nil {
		log.Fatal(err.Error())
	}
	cobra.CheckErr(rootCmd.Execute())
}

func PrintBanner() {
	_, err := fmt.Fprintf(os.Stdout, "\033[1;33m%s\033[0m\n\n", Banner())
	if err != nil {
		log.Println(err)
	}
}

func Banner() string {
	banner := strings.Join([]string{
		` ___  ____  ___  ____  ____  _  _ `,
		`/ __)( ___)/ __)(_  _)( ___)( \/ )`,
		`\__ \ )__) \__ \ _)(_  )__)  \  / `,
		`(___/(____)(___/(____)(__)   (__) `,
	}, "\n")

	return banner
}
