package cmd

import (
	"ftp-client/ftp"

	"github.com/spf13/cobra"
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir <URL>",
	Short: "Create a directory on the FTP server",
	Long:  `Create a new directory on the FTP server at the given URL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  mkdirCommand,
}

func init() {
	rootCmd.AddCommand(mkdirCmd)
}

func mkdirCommand(cmd *cobra.Command, args []string) error {
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
	return client.MakeDir(parsed.Path)
}
