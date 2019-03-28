package sshd

import (
	"io/ioutil"

	gossh "golang.org/x/crypto/ssh"
)

func getPrivatekey(keyPath string) gossh.Signer {
	privateBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}

	private, err := gossh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}
	return private
}