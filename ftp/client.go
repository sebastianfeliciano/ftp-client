package ftp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// Client implements a minimal FTP client over the control and data channels.
type Client struct {
	control net.Conn
	reader  *bufio.Reader
	verbose bool
}

// Response holds an FTP server response.
type Response struct {
	Code    int
	Message string
}

// NewClient connects to the FTP server and reads the hello message.
// Caller must call Quit() when done.
func NewClient(host string, port int, verbose bool) (*Client, error) {

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	// Dial the server and create a new client.
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	reader := bufio.NewReader(conn)
	c := &Client{control: conn, reader: reader, verbose: verbose}
	// Must read server hello before sending any command.
	_, err = c.readResponse()
	// If there is an error, close the connection.
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("read hello: %w", err)
	}
	return c, nil
}

// send sends a single command (appends \r\n).
func (c *Client) send(format string, args ...interface{}) error {
	cmd := fmt.Sprintf(format, args...) + "\r\n"
	if c.verbose {
		fmt.Print("> " + strings.TrimSuffix(cmd, "\r\n") + "\n")
	}
	_, err := c.control.Write([]byte(cmd))
	return err
}

// readResponse reads one FTP response line (code + message).
func (c *Client) readResponse() (*Response, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	if c.verbose {
		fmt.Println("< " + line)
	}
	if len(line) < 4 {
		return nil, fmt.Errorf("invalid response: %q", line)
	}
	// Convert the first 3 characters to an integer.
	code, err := strconv.Atoi(line[:3])
	if err != nil {
		return nil, fmt.Errorf("invalid response code: %q", line[:3])
	}
	msg := ""
	// If there is more than 4 characters, trim the space and assign the rest to the message.
	if len(line) > 4 {
		msg = strings.TrimSpace(line[4:])
	}
	// Track the response code and message.
	return &Response{Code: code, Message: msg}, nil
}

// readResponseMultiline reads a response that may have multiple lines.
func (c *Client) readResponseMultiline() (*Response, error) {
	first, err := c.readResponse()
	if err != nil {
		return nil, err
	}
	// If the response code is between 100 and 399 and the message is not empty and the first character is a hyphen, read multiple lines.
	if first.Code >= 100 && first.Code < 400 && len(first.Message) > 0 && first.Message[0] == '-' {
		for {
			line, err := c.reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")
			if c.verbose {
				fmt.Println("< " + line)
			}
			// If the line is at least 4 characters long, the third character is a space and the first 3 characters are the response code, return the response.
			if len(line) >= 4 && line[3] == ' ' && line[:3] == strconv.Itoa(first.Code) {
				msg := ""
				if len(line) > 4 {
					msg = strings.TrimSpace(line[4:])
				}
				return &Response{Code: first.Code, Message: msg}, nil
			}
		}
	}
	return first, nil
}

// Login sends USER and optionally PASS.
func (c *Client) Login(user, password string) error {
	if err := c.send("USER %s", user); err != nil {
		return err
	}
	resp, err := c.readResponse()
	if err != nil {
		return err
	}
	// 230 = already logged in.
	if resp.Code == 230 {
		return nil
	}
	// If the response code is 331, send the password.
	if resp.Code == 331 {
		if err := c.send("PASS %s", password); err != nil {
			return err
		}
		resp, err = c.readResponse()
		if err != nil {
			return err
		}
	}
	if resp.Code != 230 {
		return fmt.Errorf("login failed: %d %s", resp.Code, resp.Message)
	}
	return nil
}

// SetTransferMode sends TYPE I, MODE S, STRU F for binary transfer.
func (c *Client) SetTransferMode() error {
	for _, cmd := range []string{"TYPE I", "MODE S", "STRU F"} {
		if err := c.send(cmd); err != nil {
			return err
		}
		resp, err := c.readResponse()
		if err != nil {
			return err
		}
		// If the response code is not 200 or 250, return an error.
		if resp.Code != 200 && resp.Code != 250 {
			return fmt.Errorf("%s: %d %s", cmd, resp.Code, resp.Message)
		}
	}
	return nil
}

// pasvResponse is expected to follow the FTP PASV reply format:
// "227 Entering Passive Mode (h1,h2,h3,h4,p1,p2)."
var pasvRE = regexp.MustCompile(`\((\d+),(\d+),(\d+),(\d+),(\d+),(\d+)\)`)

// openDataChannel sends PASV, parses the response, and opens a TCP connection to the data port.
func (c *Client) openDataChannel() (net.Conn, error) {
	if err := c.send("PASV"); err != nil {
		return nil, err
	}
	resp, err := c.readResponse()
	if err != nil {
		return nil, err
	}
	if resp.Code != 227 {
		return nil, fmt.Errorf("PASV failed: %d %s", resp.Code, resp.Message)
	}
	matches := pasvRE.FindStringSubmatch(resp.Message)
	if matches == nil {
		return nil, fmt.Errorf("could not parse PASV response: %s", resp.Message)
	}
	// Parse the matches into 6 integers.
	var nums [6]int
	for i := 1; i <= 6; i++ {
		nums[i-1], _ = strconv.Atoi(matches[i])
	}
	// Format the host and port.
	host := fmt.Sprintf("%d.%d.%d.%d", nums[0], nums[1], nums[2], nums[3])
	// Convert the port to an integer.
	port := nums[4]<<8 + nums[5]
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	// Dial the data channel.
	dataConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("connect data channel: %w", err)
	}
	return dataConn, nil
}

// List performs LIST path and returns the listing bytes.
func (c *Client) List(path string) ([]byte, error) {
	dataConn, err := c.openDataChannel()
	if err != nil {
		return nil, err
	}
	defer dataConn.Close()
	if err := c.send("LIST %s", path); err != nil {
		return nil, err
	}
	// Server sends 150 (or 125) then listing on data channel, then 226 on control.
	resp, err := c.readResponse()
	if err != nil {
		return nil, err
	}
	if resp.Code != 150 && resp.Code != 125 {
		return nil, fmt.Errorf("LIST: %d %s", resp.Code, resp.Message)
	}
	listing, err := io.ReadAll(dataConn)
	if err != nil {
		return nil, err
	}
	resp, err = c.readResponse()
	if err != nil {
		return nil, err
	}
	if resp.Code != 226 && resp.Code != 250 {
		return nil, fmt.Errorf("LIST completion: %d %s", resp.Code, resp.Message)
	}
	return listing, nil
}

// Delete removes a file (DELE).
func (c *Client) Delete(path string) error {
	if err := c.send("DELE %s", path); err != nil {
		return err
	}
	resp, err := c.readResponse()
	if err != nil {
		return err
	}
	// If the response code is not 250, return an error.
	if resp.Code != 250 {
		return fmt.Errorf("DELE: %d %s", resp.Code, resp.Message)
	}
	return nil
}

// RemoveDir removes a directory (RMD).
func (c *Client) RemoveDir(path string) error {
	if err := c.send("RMD %s", path); err != nil {
		return err
	}
	resp, err := c.readResponse()
	if err != nil {
		return err
	}
	if resp.Code != 250 {
		return fmt.Errorf("RMD: %d %s", resp.Code, resp.Message)
	}
	return nil
}

// MakeDir creates a directory (MKD).
func (c *Client) MakeDir(path string) error {
	if err := c.send("MKD %s", path); err != nil {
		return err
	}
	resp, err := c.readResponse()
	if err != nil {
		return err
	}
	if resp.Code != 257 && resp.Code != 250 {
		return fmt.Errorf("MKD: %d %s", resp.Code, resp.Message)
	}
	return nil
}

// Retr downloads the file at path and writes it to w.
func (c *Client) Retr(path string, w io.Writer) error {
	dataConn, err := c.openDataChannel()
	if err != nil {
		return err
	}
	defer dataConn.Close()
	if err := c.send("RETR %s", path); err != nil {
		return err
	}
	resp, err := c.readResponse()
	if err != nil {
		return err
	}
	if resp.Code != 150 && resp.Code != 125 {
		return fmt.Errorf("RETR: %d %s", resp.Code, resp.Message)
	}
	_, err = io.Copy(w, dataConn)
	if err != nil {
		return err
	}
	resp, err = c.readResponse()
	if err != nil {
		return err
	}
	if resp.Code != 226 && resp.Code != 250 {
		return fmt.Errorf("RETR completion: %d %s", resp.Code, resp.Message)
	}
	return nil
}

// Stor uploads the file from r to the given path on the server.
func (c *Client) Stor(path string, r io.Reader) error {
	dataConn, err := c.openDataChannel()
	if err != nil {
		return err
	}
	defer dataConn.Close()
	if err := c.send("STOR %s", path); err != nil {
		return err
	}
	resp, err := c.readResponse()
	if err != nil {
		return err
	}
	if resp.Code != 150 && resp.Code != 125 {
		return fmt.Errorf("STOR: %d %s", resp.Code, resp.Message)
	}
	_, err = io.Copy(dataConn, r)
	if err != nil {
		return err
	}
	dataConn.Close() // client closes data channel for uploads so server knows transfer is done
	resp, err = c.readResponse()
	if err != nil {
		return err
	}
	if resp.Code != 226 && resp.Code != 250 {
		return fmt.Errorf("STOR completion: %d %s", resp.Code, resp.Message)
	}
	return nil
}

// Quit sends QUIT and closes the  connection.
func (c *Client) Quit() error {
	_ = c.send("QUIT")
	return c.control.Close()
}
