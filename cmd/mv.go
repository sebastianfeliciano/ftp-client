package cmd

import (
	"fmt"
	"ftp-client/ftp"
	"os"

	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:   "mv <ARG1> <ARG2>",
	Short: "Move a file between local and FTP server",
	Long: `Move the file given by ARG1 to the file given by ARG2.
If ARG1 is a local file, then ARG2 must be a URL, and vice-versa.`,
	Args: cobra.ExactArgs(2),
	RunE: mvCommand,
}

func init() {
	rootCmd.AddCommand(mvCmd)
}

func mvCommand(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]
	srcURL, dstURL := isFTPURL(src), isFTPURL(dst)
	if srcURL == dstURL {
		return fmt.Errorf("one argument must be a local path and the other an ftp:// URL")
	}
	if srcURL {
		// RETR from FTP to local, then DELE on server
		parsed, err := ftp.ParseURL(src)
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
		if err := client.SetTransferMode(); err != nil {
			return err
		}
		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := client.Retr(parsed.Path, f); err != nil {
			return err
		}
		return client.Delete(parsed.Path)
	}
	// STOR from local to FTP, then remove local file
	parsed, err := ftp.ParseURL(dst)
	if err != nil {
		return err
	}
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	client, err := ftp.NewClient(parsed.Host, parsed.Port, verbose)
	if err != nil {
		return err
	}
	defer client.Quit()
	if err := client.Login(parsed.User, parsed.Password); err != nil {
		return err
	}
	if err := client.SetTransferMode(); err != nil {
		return err
	}
	if err := client.Stor(parsed.Path, f); err != nil {
		return err
	}
	return os.Remove(src)
}
