package server

import (
	"net/http"
	"sync/atomic"

	"github.com/Saku0512/specter/config"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine   atomic.Pointer[gin.Engine]
	verbose  bool
	random   bool
	history  *RequestHistory
	state    *StateStore
	vars     *VarStore
	scenario *ScenarioStore
	cfg      atomic.Pointer[config.Config]
	dynamic  *DynamicRouteStore
	store    *DataStore
	timeline *TimelineStore
}

func New(cfg *config.Config, verbose bool, random bool) *Server {
	s, err := NewWithStoreFile(cfg, verbose, random, "")
	if err != nil {
		panic(err)
	}
	return s
}

func seedStoresFromConfig(cfg *config.Config) map[string][]map[string]any {
	seeds := map[string][]map[string]any{}
	for name, storeCfg := range cfg.Stores {
		seeds[name] = storeCfg.Seed
	}
	return seeds
}

func NewWithStoreFile(cfg *config.Config, verbose bool, random bool, storeFile string) (*Server, error) {
	store, err := newDataStoreWithFile(storeFile, seedStoresFromConfig(cfg))
	if err != nil {
		return nil, err
	}
	s := &Server{verbose: verbose, random: random, history: &RequestHistory{}, state: &StateStore{}, vars: newVarStore(), scenario: &ScenarioStore{}, dynamic: &DynamicRouteStore{}, store: store, timeline: newTimelineStore()}
	s.cfg.Store(cfg)
	s.engine.Store(newEngine(cfg, verbose, random, s.history, s.state, s.vars, s.scenario, s.dynamic, s.store, s.timeline, s.rebuild))
	return s, nil
}

func (s *Server) rebuild() {
	cfg := s.cfg.Load()
	s.store.SetSeed(seedStoresFromConfig(cfg))
	s.engine.Store(newEngine(cfg, s.verbose, s.random, s.history, s.state, s.vars, s.scenario, s.dynamic, s.store, s.timeline, s.rebuild))
}

func (s *Server) Reload(cfg *config.Config) {
	s.cfg.Store(cfg)
	s.rebuild()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.Load().ServeHTTP(w, r)
}
