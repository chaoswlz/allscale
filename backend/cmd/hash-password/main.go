package main

import (
  "fmt"
  "log"
  "os"
  "syscall"

  "golang.org/x/crypto/bcrypt"
  "golang.org/x/term"
)

func main() {
  fmt.Print("请输入密码: ")
  passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
  fmt.Println()
  if err != nil {
    log.Fatal(err)
  }
  if len(passwordBytes) == 0 {
    log.Fatal("密码不能为空")
  }

  hash, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
  if err != nil {
    log.Fatal(err)
  }

  if _, err := os.Stdout.Write(append(hash, '\n')); err != nil {
    log.Fatal(err)
  }
}
