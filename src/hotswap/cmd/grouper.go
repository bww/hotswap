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
 * Invoke the action immediately and clear it. If the action has already
 * been fired, this method does nothing.
 */
func (g *grouper) Invoke() error {
  g.Lock()
  defer g.Unlock()
  if g.action == nil {
    return nil
  }
  r := g.action()
  g.action = nil
  return r
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
  
  g.Lock()
  
  if g.cancel != nil {
    close(g.cancel)
  }
  g.cancel = make(chan struct{})
  cancel = g.cancel
  
  if g.timer != nil {
    g.timer.Stop()
  }
  g.timer = time.NewTimer(d)
  tick = g.timer.C
  
  g.Unlock()
  
  select {
    case <- cancel:
      fmt.Println("CANCEL")
    case <- tick:
      fmt.Println("DO IT")
      return g.Invoke()
      
  }
  
  return nil
}
