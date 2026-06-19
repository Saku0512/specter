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

func (v *VarStore) Replace(vals map[string]string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.vars = map[string]string{}
	for k, val := range vals {
		v.vars[k] = val
	}
}

type ScenarioStore struct {
	mu     sync.Mutex
	active string
}

func (s *ScenarioStore) Get() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

func (s *ScenarioStore) Set(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = name
}

func (s *ScenarioStore) Clear() {
	s.Set("")
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

func (d *DynamicRouteStore) Update(id string, route config.Route) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, r := range d.routes {
		if r.ID == id {
			d.routes[i] = DynamicRoute{ID: id, Route: route}
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

type TimelineProgress struct {
	Key         string `json:"key"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Source      string `json:"source"`
	Step        int    `json:"step"`
	Total       int    `json:"total"`
	Calls       uint64 `json:"calls"`
	Complete    bool   `json:"complete"`
	Description string `json:"description,omitempty"`
}

type TimelineStore struct {
	mu       sync.Mutex
	progress map[string]TimelineProgress
}

func newTimelineStore() *TimelineStore {
	return &TimelineStore{progress: map[string]TimelineProgress{}}
}

func (t *TimelineStore) Advance(key, method, path, source string, total int) TimelineProgress {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.progress == nil {
		t.progress = map[string]TimelineProgress{}
	}
	p := t.progress[key]
	p.Key = key
	p.Method = method
	p.Path = path
	p.Source = source
	p.Total = total
	p.Calls++
	step := int(p.Calls)
	if step > total {
		step = total
	}
	if step < 1 {
		step = 1
	}
	p.Step = step
	p.Complete = total > 0 && step >= total
	t.progress[key] = p
	return p
}

func (t *TimelineStore) Snapshot(defs []TimelineProgress) []TimelineProgress {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]TimelineProgress, 0, len(defs))
	for _, def := range defs {
		p := def
		if current, ok := t.progress[def.Key]; ok {
			p.Step = current.Step
			p.Calls = current.Calls
			p.Complete = current.Complete
		}
		out = append(out, p)
	}
	return out
}

func (t *TimelineStore) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.progress, key)
}

func (t *TimelineStore) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.progress = map[string]TimelineProgress{}
}

func newID() string { return gofakeit.UUID() }
