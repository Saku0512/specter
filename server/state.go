package server

import (
	"sync"
	"time"

	"github.com/Saku0512/specter/config"
	"github.com/brianvoe/gofakeit/v6"
)

const maxHistory = 200

type RequestEntry struct {
	Time    time.Time         `json:"time"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   map[string]string `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

type RequestHistory struct {
	mu      sync.Mutex
	entries []RequestEntry
}

func (h *RequestHistory) add(e RequestEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, e)
	if len(h.entries) > maxHistory {
		h.entries = h.entries[len(h.entries)-maxHistory:]
	}
}

func (h *RequestHistory) all() []RequestEntry {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]RequestEntry, len(h.entries))
	copy(out, h.entries)
	return out
}

func (h *RequestHistory) clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = nil
}

type StateStore struct {
	mu    sync.Mutex
	value string
}

func (s *StateStore) Get() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value
}

func (s *StateStore) Set(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value = v
}

type VarStore struct {
	mu   sync.Mutex
	vars map[string]string
}

func newVarStore() *VarStore { return &VarStore{vars: map[string]string{}} }

func (v *VarStore) Get(key string) string {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.vars[key]
}

func (v *VarStore) Set(key, val string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.vars[key] = val
}

func (v *VarStore) Delete(key string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.vars, key)
}

func (v *VarStore) All() map[string]string {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := make(map[string]string, len(v.vars))
	for k, val := range v.vars {
		out[k] = val
	}
	return out
}

func (v *VarStore) Clear() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.vars = map[string]string{}
}

// DynamicRoute is a route added at runtime via the introspection API.
type DynamicRoute struct {
	ID    string       `json:"id"`
	Route config.Route `json:"route"`
}

type DynamicRouteStore struct {
	mu     sync.Mutex
	routes []DynamicRoute
}

func (d *DynamicRouteStore) Add(route config.Route) string {
	id := newID()
	d.mu.Lock()
	d.routes = append(d.routes, DynamicRoute{ID: id, Route: route})
	d.mu.Unlock()
	return id
}

func (d *DynamicRouteStore) Remove(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, r := range d.routes {
		if r.ID == id {
			d.routes = append(d.routes[:i], d.routes[i+1:]...)
			return true
		}
	}
	return false
}

func (d *DynamicRouteStore) All() []DynamicRoute {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]DynamicRoute, len(d.routes))
	copy(out, d.routes)
	return out
}

func (d *DynamicRouteStore) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.routes = nil
}

func newID() string { return gofakeit.UUID() }
