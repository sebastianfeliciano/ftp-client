package cmd

import (
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "4700ftp [-h] [--verbose] operation params [params ...]",
	Short: "FTP client for listing, copying, moving, and deleting files and directories on remote FTP servers.",
	Long: `FTP client for listing, copying, moving, and deleting files and directories on remote FTP servers.

positional arguments:
operation      The operation to execute. Valid operations are 'ls', 'rm', 'rmdir',
              'mkdir', 'cp', and 'mv'
params         Parameters for the given operation. Will be one or two paths and/or URLs.

URL format: ftp://[USER[:PASSWORD]@]HOST[:PORT]/PATH
Default USER is 'anonymous' with no PASSWORD. Default PORT is 21.`,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print all messages to and from the FTP server")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
