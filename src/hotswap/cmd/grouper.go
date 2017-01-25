package main

import (
  "time"
  "sync"
)

/**
 * A grouper action
 */
type grouperAction func(interface{})(error)

/**
 * Groups invocations
 */
type grouper struct {
  sync.Mutex
  pevent  time.Time
  ecount  int
  tlimit  time.Duration
  elimit  int
  action  grouperAction
  first   bool
}

/**
 * Create a grouper
 */
func newGrouper(t time.Duration, e int, a grouperAction) *grouper {
  return &grouper{sync.Mutex{}, time.Time{}, 0, t, e, a, true}
}

/**
 * Note an event.
 */
func (g *grouper) Event(v interface{}) error {
  g.Lock()
  defer g.Unlock()
  g.ecount++
  
  var err error
  if g.first || g.ecount > g.elimit || time.Since(g.pevent) > g.tlimit {
    err = g.action(v)
    g.ecount = 0
  }
  
  g.pevent = time.Now()
  return err
}
