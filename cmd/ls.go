package cmd

import (
	"fmt"

	"ftp-client/ftp"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls <URL>",
	Short: "List current directory on the FTP server",
	Long:  `Print the directory listing from the FTP server at the given URL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLs,
}

func init() {
	rootCmd.AddCommand(lsCmd)
}

func runLs(cmd *cobra.Command, args []string) error {
	parsed, err := ftp.ParseURL(args[0])
	if err != nil {
		return err
	}

	client, err := ftp.NewClient(parsed.Host, parsed.Port, verbose)
	if err != nil {
		return err
	}

	defer client.Quit()
	if err := client.Login(parsed.User, parsed.Password); err != nil {
		return err
	}

	listing, err := client.List(parsed.Path)
	if err != nil {
		return err
	}
	fmt.Print(string(listing))
	return nil
}
