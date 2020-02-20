package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
)

func main() {
	user := flag.String("u", "root", "user name")
	passwd := flag.String("p", "dawter", "user passwd")
	server := flag.String("s", "192.168.55.2:22", "ssh server")
	cmd1 := flag.String("cmd1", "ls", "cmd to tun")

	flag.Parse()

	if err := run(*user, *passwd, *server, *cmd1); err != nil {
		fmt.Printf("error: %s\n", err)
	}

}

func run(user, passwd, server, cmd1 string) error {
	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				// Just send the password back for all questions
				answers := make([]string, len(questions))
				for i := range answers {
					answers[i] = passwd // replace this
				}
				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	client, err := ssh.Dial("tcp", server, &config)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	outReader, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	errReader, err := session.StderrPipe()
	if err != nil {
		return err
	}

	if err := session.Run(cmd1); err != nil {
		b := &bytes.Buffer{}
		io.Copy(b, errReader)
		fmt.Printf("stderr: %s\n", b.String())
		return err
	}

	i := 0
	scanner := bufio.NewScanner(outReader)
	for scanner.Scan() {
		//fields := strings.Split(scanner.Text(), " ")
		fmt.Println(i, scanner.Text())
		i++
	}
	return scanner.Err()
}
