package cmd

import (
	"ftp-client/ftp"

	"github.com/spf13/cobra"
)

var rmdirCmd = &cobra.Command{
	Use:   "rmdir [URL]",
	Short: "Delete a directory on the FTP server",
	Long:  `Delete the directory on the FTP server at the given URL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  rmdirCommand,
}

func init() {
	rootCmd.AddCommand(rmdirCmd)
}

func rmdirCommand(cmd *cobra.Command, args []string) error {
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
	return client.RemoveDir(parsed.Path)
}
