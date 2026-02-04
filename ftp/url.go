package ftp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// FTPURL holds parsed FTP connection and path information.
// Format: ftp://[USER[:PASSWORD]@]HOST[:PORT]/PATH
type FTPURL struct {
	Host     string
	Port     int
	User     string
	Password string
	Path     string
}

// DefaultFTPPort is the default FTP control  port.
const DefaultFTPPort = 21

// ParseURL parses an FTP URL string into FTPURL.
// Format: ftp://[user:password@]host[:port]/path
// Default user is "anonymous" with no password; default port is 21.
func ParseURL(raw string) (*FTPURL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme != "ftp" {
		return nil, fmt.Errorf("unsupported: %s (expected ftp)", u.Scheme)
	}
	// We have to get the host.
	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing host in URL")
	}

	port := DefaultFTPPort
	if p := u.Port(); p != "" {
		port, err = strconv.Atoi(p)
		if err != nil || port <= 0 || port > 65535 {
			return nil, fmt.Errorf("invalid port: %s", p)
		}
	}
	// Default user is "anonymous" with no password.
	user := "anonymous"
	password := ""
	if u.User != nil {
		user = u.User.Username()
		password, _ = u.User.Password()
	}
	path := u.Path
	if path == "" {
		path = "/"
	}
	// Ensure path is absolute for FTP (no leading slash required by some servers, but we keep it)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return &FTPURL{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Path:     path,
	}, nil
}

// String returns a short description of the URL (without password).
func (f *FTPURL) String() string {
	return fmt.Sprintf("ftp://%s@%s:%d%s", f.User, f.Host, f.Port, f.Path)
}
