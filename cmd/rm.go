package cmd

import (
	"ftp-client/ftp"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm [URL]",
	Short: "Delete a file on the FTP server",
	Long:  `Delete the file on the FTP server at the given URL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  rmCommand,
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

func rmCommand(cmd *cobra.Command, args []string) error {
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
	return client.Delete(parsed.Path)
}
