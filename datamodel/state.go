package datamodel

import (
	"strings"
	"sync"
)

type state struct {
	Bootstrapped bool
	Changes      map[string]Parameter
	Deleted      map[string]struct{}
	original     map[string]Parameter
	lock         sync.RWMutex
}

func newState(original map[string]Parameter) *state {
	return &state{
		Changes:  make(map[string]Parameter),
		Deleted:  make(map[string]struct{}),
		original: original,
	}
}

func (s *state) get(name string) (p Parameter, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if _, ok := s.Deleted[name]; ok {
		return p, false
	}
	if p, ok = s.Changes[name]; ok {
		return
	}
	p, ok = s.original[name]
	return
}

func (s *state) save(p Parameter) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Changes[p.Path] = p
	delete(s.Deleted, p.Path)
}

func (s *state) delete(name string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.Changes[name]; ok {
		delete(s.Changes, name)
		s.Deleted[name] = struct{}{}
	} else if _, ok := s.original[name]; ok {
		s.Deleted[name] = struct{}{}
	}
}

func (s *state) deletePrefix(prefix string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, p := range s.Changes {
		if strings.HasPrefix(p.Path, prefix) {
			delete(s.Changes, p.Path)
			s.Deleted[p.Path] = struct{}{}
		}
	}
	for _, p := range s.original {
		if strings.HasPrefix(p.Path, prefix) {
			delete(s.Changes, p.Path)
			s.Deleted[p.Path] = struct{}{}
		}
	}
}

func (s *state) forEach(fn func(Parameter) (cont bool)) {
	for _, p := range s.Changes {
		if cont := fn(p); !cont {
			return
		}
	}
	for _, p := range s.original {
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
