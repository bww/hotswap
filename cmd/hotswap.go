package main

import (
  "os"
  "os/exec"
  "os/signal"
  "io"
  "fmt"
  "time"
  "sync"
  "strings"
)

var lock sync.Mutex
var proc *os.Process

var previous time.Time
const threshold = time.Second * 3

/**
 * You know what it does.
 */
func main() {
  if len(os.Args) < 2 {
    fmt.Printf("%v: Usage hotswap <command> [args]\n", os.Args[0])
    return
  }
  
  c := os.Args[1]
  a := os.Args[2:]
  
  c, err := resolve(c)
  if err != nil {
    panic(err)
  }
  
  go monitor(c)
  go signals()
  
  for {
    run(c, a)
  }
}

/**
 * Resove an executable path
 */
func resolve(c string) (string, error) {
  _, err := os.Stat(c)
  if err != nil {
    if !os.IsNotExist(err) {
      return "", err
    }
    c, err = exec.LookPath(c)
    if err != nil {
      return "", err
    }
  }
  return c, nil
}

/**
 * Get the currently-running process
 */
func process() *os.Process {
  lock.Lock()
  defer lock.Unlock()
  return proc
}

/**
 * Set the currently-running process
 */
func setProcess(p *os.Process) {
  lock.Lock()
  proc = p
  lock.Unlock()
}

/**
 * Run a process.
 */
func run(c string, a []string) {
  fmt.Printf("%v: %v %v", os.Args[0], c, strings.Join(a, " "))
  
  cmd := exec.Command(c, a...)
  cmd.Env = append(os.Environ(), fmt.Sprintf("GO_HOTSWAP_MANAGER_PID=%d", os.Getpid()))
  
  pout, err := cmd.StdoutPipe()
  if err != nil {
    panic(err)
  }
  
  perr, err := cmd.StderrPipe()
  if err != nil {
    panic(err)
  }
  
  err = cmd.Start()
  if err != nil {
    panic(err)
  }
  
  setProcess(cmd.Process)
  defer setProcess(nil)
  
  fmt.Println()
  go io.Copy(os.Stdout, pout)
  go io.Copy(os.Stdout, perr)
  
  err = cmd.Wait()
  if err != nil {
    fmt.Printf("%v: Process exited with error: %v\n", os.Args[0], err)
  }
  
}

/**
 * Kill the currently running process, allowing it to restart
 */
func term(p *os.Process) {
  fmt.Printf("%v: Reloading process", os.Args[0])
  if p != nil {
    err := p.Signal(os.Interrupt)
    if err != nil {
      fmt.Printf("%v: Could not signal process %v: %v", os.Args[0], p.Pid, err)
    }
  }
}

/**
 * Monitor for restart
 */
func monitor(r string) {
  for range time.Tick(time.Second) {
    // ...
  }
}

/**
 * Handle signals
 */
func signals() {
  sig := make(chan os.Signal, 1)
  signal.Notify(sig, os.Interrupt)
  signal.Notify(sig, os.Kill)
  go func() {
    for e := range sig {
      if e == os.Kill || time.Since(previous) < threshold {
        os.Exit(0)
      }else{
        term(process())
      }
      previous = time.Now()
    }
  }()
}
