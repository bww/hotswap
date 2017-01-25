package main

import (
  "fmt"
  "time"
  "sync"
)

/**
 * A grouper action
 */
type grouperAction func()(error)

/**
 * Groups invocations
 */
type grouper struct {
  sync.Mutex
  delay   time.Duration
  action  grouperAction
  cancel  chan struct{}
  timer   *time.Timer
}

/**
 * Create a grouper
 */
func newGrouper(d time.Duration, a grouperAction) *grouper {
  return &grouper{sync.Mutex{}, d, a, nil, nil}
}

/**
 * Note an event.
 */
func (g *grouper) Event() {
  g.After(g.delay)
}

/**
 * Wait asynchronously and fire the action after the specified duration
 */
func (g *grouper) After(d time.Duration) {
  go func() {
    err := g.after(d)
    if err != nil {
      panic(err)
    }
  }()
}

/**
 * Wait and fire asynchronously after the specified duration
 */
func (g *grouper) after(d time.Duration) error {
  var tick <-chan time.Time
  var cancel <-chan struct{}
  
  fmt.Println("HERE UH OK")
  g.Lock()
  
  fmt.Println("CHECK-1", g.cancel)
  if g.cancel != nil {
    close(g.cancel)
  }
  g.cancel = make(chan struct{})
  cancel = g.cancel
  
  fmt.Println("CHECK-O", g.timer)
  if g.timer != nil {
    fmt.Println("STOP/CANCEL/SWAP TIMER")
    g.timer.Stop()
  }
  g.timer = time.NewTimer(d)
  tick = g.timer.C
  
  g.Unlock()
  
  fmt.Println("WAIT", g.delay)
  
  select {
    case <- cancel:
      fmt.Println("CANCEL")
    case <- tick:
      fmt.Println("DO IT")
      return g.action()
  }
  
  fmt.Println("MMKAY")
  return nil
}
