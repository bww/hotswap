package main

import (
  "os"
  "os/exec"
  "os/signal"
  "io"
  "fmt"
  "time"
  "flag"
  "sync"
  "path"
  "strings"
)

import (
  "github.com/fsnotify/fsnotify"
)

var pname string
var lock sync.Mutex
var proc *os.Process
var group *grouper

var   psignal time.Time
const threshold = time.Second * 3

var conf struct {
  Debug   bool
  Verbose bool
}

/**
 * Flag string list
 */
type flagList []string

/**
 * Set a flag
 */
func (s *flagList) Set(v string) error {
  *s = append(*s, v)
  return nil
}

/**
 * Describe
 */
func (s *flagList) String() string {
  return fmt.Sprintf("%+v", *s)
}

/**
 * You know what it does.
 */
func main() {
  var watchDirs, watchFilters flagList
  
  pname = os.Args[0]
  if x := strings.LastIndex(pname, "/"); x > 0 {
    pname = pname[x+1:]
  }
  
  cmdline       := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  fDebug        := cmdline.Bool     ("debug",         false,    "Enable debugging mode.")
  fVerbose      := cmdline.Bool     ("verbose",       false,    "Enable verbose debugging mode.")
  cmdline.Var    (&watchDirs,        "watch",                   "Watch a directory tree for changes. Provide this flag repeatedly to watch multiple directories.")
  cmdline.Var    (&watchFilters,     "filter",                  "Watch only files with specific name patterns for changes. Specify a glob pattern, e.g. '*.go'.")
  cmdline.Parse(os.Args[1:])
  
  conf.Debug = *fDebug
  conf.Verbose = *fVerbose
  
  group = newGrouper(time.Millisecond * 100, 100, func(v interface{}) error {
    return term(v.(*os.Process))
  })
  
  if len(watchDirs) > 0 {
    fmt.Printf("%v: Watching roots:\n", pname)
    for _, e := range watchDirs {
      fmt.Printf("  -> %s\n", e)
    }
  }
  
  args := cmdline.Args()
  if len(args) < 2 {
    fmt.Printf("%v: Usage hotswap <command> [args]\n", pname)
    return
  }
  
  c := args[0]
  a := args[1:]
  
  c, err := resolve(c)
  if err != nil {
    panic(err)
  }
  
  go monitor(watchDirs, watchFilters)
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
  fmt.Printf("%v: %v %v\n", pname, c, strings.Join(a, " "))
  
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
    fmt.Printf("%v: Process exited with error: %v\n", pname, err)
  }
  
}

/**
 * Kill the currently running process, allowing it to restart
 */
func term(p *os.Process) error {
  fmt.Printf("%v: Reloading process...\n", pname)
  if p != nil {
    err := p.Signal(os.Interrupt)
    if err != nil {
      return fmt.Errorf("Could not signal process %v: %v", p.Pid, err)
    }
  }
  return nil
}

/**
 * Monitor for restart
 */
func monitor(d, f []string) {
  if len(d) < 1 {
    return
  }
  
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    panic(err)
  }
  
  for _, e := range d {
    err = monitorPath(watcher, e, f)
    if err != nil {
      panic(err)
    }
  }
  
  for {
    select {
      case err, ok := <- watcher.Errors:
        if !ok { break }
        panic(err)
      case _, ok := <- watcher.Events:
        if !ok { break }
        err := group.Event(process())
        if err != nil {
          panic(err)
        }
    }
  }
  
}

/**
 * Recursively monitor
 */
func monitorPath(w *fsnotify.Watcher, d string, f []string) error {
  
  finfo, err := os.Stat(d)
  if err != nil {
    return err
  }
  
  if finfo.IsDir() {
    
    file, err := os.Open(d)
    if err != nil {
      return err
    }
    
    defer file.Close()
    
    edir, err := file.Readdir(0)
    if err != nil {
      return err
    }
    
    for _, e := range edir {
      err = monitorPath(w, path.Join(d, e.Name()), f)
      if err != nil {
        return err
      }
    }
    
  }else{
    
    if len(f) > 0 {
      match := false
      fname := finfo.Name()
      for _, x := range f {
        m, err := path.Match(x, fname)
        if err != nil {
          return err
        }
        if m {
          match = true
          break
        }
      }
      if !match {
        return nil
      }
    }
    
    if conf.Verbose {
      fmt.Println("  +", d)
    }
    w.Add(d)
  }
  
  return nil
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
      if e == os.Kill || time.Since(psignal) < threshold {
        os.Exit(0)
      }else if err := term(process()); err != nil {
        panic(err)
      }
      psignal = time.Now()
    }
  }()
}
