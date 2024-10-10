package datamodel

import (
	"strings"
	"sync"
)

type State struct {
	Bootstrapped bool
	Changes      map[string]Parameter
	Deleted      map[string]struct{}
	defaults     map[string]Parameter
	lock         sync.RWMutex
}

func newState() *State {
	return &State{
		Changes:  make(map[string]Parameter),
		Deleted:  make(map[string]struct{}),
		defaults: make(map[string]Parameter),
	}
}

func (s *State) WithDefaults(dm map[string]Parameter) *State {
	s.defaults = dm
	return s
}

func (s *State) get(name string) (p Parameter, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if _, ok := s.Deleted[name]; ok {
		return p, false
	}
	if p, ok = s.Changes[name]; ok {
		return
	}
	p, ok = s.defaults[name]
	return
}

func (s *State) forEach(fn func(Parameter) (cont bool)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, p := range s.Changes {
		if cont := fn(p); !cont {
			return
		}
	}
	for _, p := range s.defaults {
		// Skip if deleted
		if _, ok := s.Deleted[p.Path]; ok {
			continue
		}
		// Skip if present in the state
		if _, ok := s.Changes[p.Path]; ok {
			continue
		}
		if cont := fn(p); !cont {
			return
		}
	}
}

func (s *State) save(p Parameter) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Changes[p.Path] = p
	delete(s.Deleted, p.Path)
}

func (s *State) delete(name string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.Changes[name]; ok {
		delete(s.Changes, name)
		s.Deleted[name] = struct{}{}
	} else if _, ok := s.defaults[name]; ok {
		s.Deleted[name] = struct{}{}
	}
}

func (s *State) deletePrefix(prefix string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, p := range s.Changes {
		if strings.HasPrefix(p.Path, prefix) {
			delete(s.Changes, p.Path)
			s.Deleted[p.Path] = struct{}{}
		}
	}
	for _, p := range s.defaults {
		if strings.HasPrefix(p.Path, prefix) {
			s.Deleted[p.Path] = struct{}{}
		}
	}
}

func (s *State) reset() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Bootstrapped = false
	s.Changes = make(map[string]Parameter)
	s.Deleted = make(map[string]struct{})
}
