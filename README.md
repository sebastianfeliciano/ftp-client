# 4700ftp – FTP Client

A command-line FTP client that supports directory listing, creating and removing directories, deleting files, and copying or moving files between the local machine and a remote FTP server.

## High-Level Approach

- **CLI**: Cobra provides the root command and six subcommands (`ls`, `mkdir`, `rm`, `rmdir`, `cp`, `mv`) with a shared `--verbose` flag. Arguments are either FTP URLs or local paths;

- **URL parsing**: FTP URLs are parsed with `net/url` into host, port (default 21), user (default `anonymous`), password, and path.

- **FTP protocol**: A small `ftp` package implements the protocol by hand. It opens a control connection to the server, reads the welcome message before sending any command, then sends USER/PASS for login. For any data transfer (list, download, upload) it sends PASV, parses the 227 reply to get the data port, opens a second TCP connection, performs the transfer, then reads the completion response on the control channel. TYPE I, MODE S, and STRU F are sent before uploads/downloads. The client closes the data connection after uploads; the server closes it after list and download.

- **Operations**: Each Cobra subcommand parses its arguments (URLs and/or paths), calls into the `ftp` package to connect, login, run the right FTP commands, and then quits. The hardest part was implementing and wiring the FTP commands correctly, since they are the backbone of the Cobra commands and of the parsing: every operation depends on sending the right commands in the right order and handling control vs. data channel correctly.

## Challenges

- **FTP command order and semantics**: Getting the FTP command sequence right was the main difficulty. The server must receive the welcome message read first, then USER (and PASS when required), and for data operations PASV must be sent and the 227 reply parsed to open the data channel before LIST, RETR, or STOR. 

TYPE/MODE/STRU are required before binary transfers. Mistakes here caused confusing server errors until the flow matched the spec.

If even one step is out of order, the server just hangs or throws an error. Doing the Math for Ports: When the server is ready to send data, it sends back a weird-looking string of six numbers. I had to use a regex to grab those numbers and do some math to figure out which port to connect to.

- **Source vs. Destination**: For commands like cp, I had to write logic to figure out if the user was trying to upload (local to remote) or download (remote to local) based on the URL format.
## How I Tested the Code

- **Build**: Ran `make` (and `go build -o 4700ftp .`) to ensure the project compiles and produces the `4700ftp` binary.

- **Verbose protocol check**: Ran `./4700ftp -v ls "ftp://user:pass@ftp.4700.network/"` to see the raw FTP exchange (welcome, USER, PASS, then either listing or error). This verified that the control connection, login sequence, and  PASV/LIST work and that messages end with `\r\n`.

- **Live server**: Used the course FTP server at `ftp.4700.network` with my credentials: listed the home directory and ran commands inside it, to see if files and directories could be made and deleted. Also the autograder was very handy.