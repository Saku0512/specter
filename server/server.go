package server

import (
	"net/http"
	"sync/atomic"

	"github.com/Saku0512/specter/config"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine  atomic.Pointer[gin.Engine]
	verbose bool
	history *RequestHistory
	state   *StateStore
	vars    *VarStore
	cfg     atomic.Pointer[config.Config]
	dynamic *DynamicRouteStore
	store   *DataStore
}

func New(cfg *config.Config, verbose bool) *Server {
	s := &Server{verbose: verbose, history: &RequestHistory{}, state: &StateStore{}, vars: newVarStore(), dynamic: &DynamicRouteStore{}, store: newDataStore()}
	s.cfg.Store(cfg)
	s.engine.Store(newEngine(cfg, verbose, s.history, s.state, s.vars, s.dynamic, s.store, s.rebuild))
	return s
}

func (s *Server) rebuild() {
	cfg := s.cfg.Load()
	s.engine.Store(newEngine(cfg, s.verbose, s.history, s.state, s.vars, s.dynamic, s.store, s.rebuild))
}

func (s *Server) Reload(cfg *config.Config) {
	s.cfg.Store(cfg)
	s.rebuild()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.Load().ServeHTTP(w, r)
}
