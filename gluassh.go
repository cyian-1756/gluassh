package gluassh

import (
	"bytes"
	"fmt"
	"io"
	"log"

	lua "github.com/yuin/gopher-lua"
	"golang.org/x/crypto/ssh"
)

func Loader(L *lua.LState) int {
	tb := L.NewTable()
	L.SetFuncs(tb, map[string]lua.LGFunction{
		"connectAndCommand": connectAndCommand,
		"login":             login,
		"sendCommand":       sendCommand,
	})
	L.Push(tb)

	return 1
}

func handleError(err error, L *lua.LState) int {
	L.Push(lua.LNil)
	L.Push(lua.LString(err.Error()))
	return 2

}

type LuaSshSess struct {
	*ssh.Session
	stdin  io.WriteCloser
	stdout io.Reader
	stderr io.Reader
}

// Starts a new ssh session
func login(L *lua.LState) int {
	username := L.CheckString(1)
	password := L.CheckString(2)
	hostname := L.CheckString(3)
	port := L.CheckString(4)
	secure := L.CheckBool(5)
	config := &ssh.ClientConfig{}
	if secure {
		// SSH client config
		config = &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
			// Non-production only
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else {
		// SSH client config
		config = &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
		}
	}

	// Connect to host
	client, err := ssh.Dial("tcp", hostname+":"+port, config)

	if err != nil {
		return handleError(err, L)
	}

	sess, err := client.NewSession()
	if err != nil {
		return handleError(err, L)
	}

	stdBuf, err := sess.StdoutPipe()
	if err != nil {
		return handleError(err, L)
	}

	stdin, err := sess.StdinPipe()
	if err != nil {
		return handleError(err, L)
	}

	stderr, err := sess.StderrPipe()
	if err != nil {
		return handleError(err, L)
	}

	err = sess.Shell()
	if err != nil {
		return handleError(err, L)
	}

	ud := L.NewUserData()
	ud.Value = &LuaSshSess{Session: sess, stdin: stdin, stdout: stdBuf, stderr: stderr}
	L.SetMetatable(ud, L.GetTypeMetatable("ssh_session_ud"))
	L.Push(ud)
	L.Push(lua.LString(string(readConnection(stdBuf))))
	return 2
}

func readConnection(stdBuf io.Reader) []byte {
	buf := make([]byte, 0, 4096) // big buffer
	tmp := make([]byte, 1024)    // using small tmo buffer for demonstrating
	for {
		n, err := stdBuf.Read(tmp)
		if err != nil {
			if err != io.EOF {
				log.Fatal("read error:", err)
			}
			break
		}
		buf = append(buf, tmp[:n]...)
		if 1024 > n {
			break
		}

	}
	return buf
}

// Sends a command to a active ssh session and returns the output
func sendCommand(L *lua.LState) int {
	ud := L.CheckUserData(1)
	command := L.CheckString(2)
	v, _ := ud.Value.(*LuaSshSess)

	_, err := fmt.Fprintf(v.stdin, "%s\n", command)
	if err != nil {
		return handleError(err, L)
	}
	r := string(readConnection(v.stdout))
	L.Push(lua.LString(r))
	L.Push(lua.LNil)
	return 2

	// L.Push(lua.LNil)
	// L.Push(lua.LString("Expected ssh_session_ud"))
	// return 2
}

// Connects to a host, runs a single command and exits.
// Returns: A string with the output of the command and nil if everything worked. If not it returns
// nil and an error message
func connectAndCommand(L *lua.LState) int {

	username := L.CheckString(1)
	password := L.CheckString(2)
	hostname := L.CheckString(3)
	port := L.CheckString(4)
	secure := L.CheckBool(5)
	command := L.CheckString(6)

	config := &ssh.ClientConfig{}

	if secure {
		// SSH client config
		config = &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
			// Non-production only
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else {
		// SSH client config
		config = &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
		}
	}

	// Connect to host
	client, err := ssh.Dial("tcp", hostname+":"+port, config)

	if err != nil {
		handleError(err, L)
	}

	defer client.Close()

	// Create sesssion
	sess, err := client.NewSession()
	if err != nil {
		handleError(err, L)
	}
	defer sess.Close()

	// StdinPipe for commands
	stdin, err := sess.StdinPipe()
	if err != nil {
		handleError(err, L)
	}

	// Uncomment to store output in variable
	var b bytes.Buffer
	sess.Stdout = &b
	sess.Stderr = &b

	// Start remote shell
	err = sess.Shell()
	if err != nil {
		handleError(err, L)
	}

	_, err = fmt.Fprintf(stdin, "%s\n", command)
	if err != nil {
		handleError(err, L)
	}
	fmt.Fprintf(stdin, "%s\n", "exit")

	// Wait for sess to finish
	err = sess.Wait()
	if err != nil {
		handleError(err, L)
	}

	L.Push(lua.LString(b.String()))
	L.Push(lua.LNil)
	return 2

}
