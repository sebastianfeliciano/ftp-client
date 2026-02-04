package cmd

import (
	"fmt"
	"os"
	"strings"

	"ftp-client/ftp"

	"github.com/spf13/cobra"
)

var cpCmd = &cobra.Command{
	Use:   "cp <ARG1> <ARG2>",
	Short: "Copy a file between local and FTP server",
	Long: `Copy the file given by ARG1 to the file given by ARG2.
If ARG1 is a local file, then ARG2 must be a URL, and vice-versa.`,
	Args: cobra.ExactArgs(2),
	RunE: runCp,
}

func init() {
	rootCmd.AddCommand(cpCmd)
}

func isFTPURL(s string) bool { return strings.HasPrefix(s, "ftp://") }

func runCp(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]
	srcURL, dstURL := isFTPURL(src), isFTPURL(dst)
	if srcURL == dstURL {
		return fmt.Errorf("one argument must be a local path and the other an ftp:// URL")
	}
	if srcURL {
		// RETR from FTP to local
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
		return client.Retr(parsed.Path, f)
	}
	// STOR from local to FTP
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
	return client.Stor(parsed.Path, f)
}
