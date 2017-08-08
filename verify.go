package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
)

func sshLogin(taskItem map[string]string) {
	host := taskItem["host"]
	username := taskItem["username"]
	password := taskItem["password"]

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout: time.Duration(timeout) * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	Client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		if strings.Contains(err.Error(), "unable to authenticate") {
			outputStr = fmt.Sprintf(" %-15s|%s => %s", host, username, password)
			return
		} else if sshClosed(err.Error()) {
			outputStr = fmt.Sprintf(" %-15s|%s => %s", host, username, password)
			deadHost.Add(host)
		} else {
			FlushPrint(PrepareText(fmt.Sprintf(" %-15s|%s => %s %v%s", host, username, password, err, LF)))
			deadHost.Add(host)
		}
	} else {
		defer Client.Close()
		session, err := Client.NewSession()
		if err != nil {
			return
		}
		defer session.Close()

		var b bytes.Buffer
		session.Stdout = &b
		err = session.Run("ifconfig")
		if err == nil {
			if strings.Contains(b.String(), "127.0.0.1") {
				vulHost.Add(host)
				text := fmt.Sprintf("[success]%-15s|%s => %s", host, username, password)
				FlushPrint(green(PrepareText(text) + LF))
				writeLock.Lock()
				Write2File(vulFile, "["+NowTime("")+"]"+host+" "+username+"=>"+password+LF)
				writeLock.Unlock()
			}
		}
	}
}

func sshClosed(text string) bool {
	return strings.Contains(text, "i/o timeout") ||
		strings.Contains(text, "connection timed out") ||
		strings.Contains(text, "closed by the remote host") ||
		strings.Contains(text, "handshake failed") ||
		strings.Contains(text, "refused it") ||
		strings.Contains(text, "no route to host") ||
		strings.Contains(text, "connection refused")
}

func mysqlClosed(text string) bool {
	return strings.Contains(text, "i/o timeout") ||
		strings.Contains(text, "not allowed to connect to")
}

func mysqlLogin(taskItem map[string]string) {
	host := taskItem["host"]
	username := taskItem["username"]
	password := taskItem["password"]
	db, _ := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/?charset=utf8&timeout=%ds", username, password, host, timeout))
	defer db.Close()
	err := db.Ping()
	if err != nil {
		if strings.Contains(err.Error(), "Access denied for user") {
			outputStr = fmt.Sprintf(" %-15s|%s => %s", host, username, password)
			return
		} else {
			if mysqlClosed(err.Error()) {
				deadHost.Add(host)
			} else {
				FlushPrint(red(PrepareText(host+" "+err.Error()) + LF))
				deadHost.Add(host)
			}
		}
	} else {
		vulHost.Add(host)
		text := fmt.Sprintf("[success]%-15s|%s => %s", host, username, password)
		FlushPrint(green(PrepareText(text) + LF))
		writeLock.Lock()
		Write2File(vulFile, "["+NowTime("")+"]"+host+" "+username+"=>"+password+LF)
		writeLock.Unlock()
	}
}

func ftpLogin(taskItem map[string]string) {
	//	host := taskItem["host"]
	//	username := taskItem["username"]
	//	password := taskItem["password"]
}
