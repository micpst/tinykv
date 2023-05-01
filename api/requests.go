package api

type rebalanceRequest struct {
	key  []byte
	from string
	to   string
}

func (s *Server) rebalance(req rebalanceRequest) bool {
	return true
}
