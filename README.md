# gluassh
A simple ssh lib for gopher-lua based on golang.org/x/crypto/ssh

## Usage

Import the lib `import "github.com/cyian-1756/gluassh"`

Load it with `L.PreloadModule("ssh", gluassh.Loader)`

Require the lib in lua

`ssh = require("ssh")`

Get a shell

`local shell, err = ssh.login(USERNAME, PASSWORD, HOST, PORT, InsecureIgnoreHostKey)`

If InsecureIgnoreHostKey is true then the host key is ignored

Send a command 

`local r, err = ssh.sendCommand(shell, "whoami")`

If you just want to connect and run a one liner you can use the helper func connectAndCommand

`ssh.connectAndCommand(USERNAME, PASSWORD, HOST, PORT, InsecureIgnoreHostKey, COMMAND)`