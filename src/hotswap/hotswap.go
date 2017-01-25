package main

import (
  "os"
  "os/exec"
  "fmt"
  "time"
  // "path"
  // "sort"
  "strings"
  // "strconv"
  // "io/ioutil"
)

// import (
//   "github.com/kardianos/osext"
// )

func main() {
  fmt.Println(strings.Join(os.Args, ", "))
  <- time.After(time.Second * 2)
  
  if os.Getenv("GO_HOTSWAP_CREATOR_PID") != "" {
    fmt.Println("THIS IS A SPAWN PROCESS")
    fmt.Println(os.Environ())
    <- time.After(time.Second * 10)
    fmt.Println("Ok, bye.")
  }else{
    fmt.Println("THIS IS A ROOT PROCESS")
    cmd := exec.Command(os.Args[0], os.Args[1:]...)
    cmd.Env = append(os.Environ(), fmt.Sprintf("GO_HOTSWAP_CREATOR_PID=%d", os.Getpid()))
    err := cmd.Start()
    if err != nil {
      panic(err)
    }
  }
  
}
