package main

import (
  "fmt"
  "time"
)

func main() {
  t := time.Now()
  for {
    fmt.Println("--->", t)
    <- time.After(time.Second)
  }
}
