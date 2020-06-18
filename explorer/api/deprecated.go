package api

import (
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-indexer/core/server"
	"net/http"
)

// Deprecated
func (s *httpServer) epochsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochsOld", r.RequestURI)
	defer s.pm.Complete(id)

	startIndex, count, err := server.ReadOldPaginatorParams(mux.Vars(r))
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochsOld(startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) epochBlocksOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochBlocksOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochBlocksOld(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) epochFlipsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochFlipsOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochFlipsOld(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) epochIdentitiesOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentitiesOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentitiesOld(epoch, convertStates(r.Form["prevstates[]"]), convertStates(r.Form["states[]"]),
		startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) epochInvitesOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochInvitesOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochInvitesOld(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) epochTxsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochTxsOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochTxsOld(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) epochBadAuthorsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochBadAuthorsOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochBadAuthorsOld(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) epochIdentitiesRewardsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochIdentitiesRewardsOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochIdentitiesRewardsOld(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) blockTxsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("blockTxsOld", r.RequestURI)
	defer s.pm.Complete(id)

	var resp interface{}
	vars := mux.Vars(r)
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	height, err := server.ReadUint(vars, "id")
	if err != nil {
		resp, err = s.db.BlockTxsByHashOld(vars["id"], startIndex, count)
	} else {
		resp, err = s.db.BlockTxsByHeightOld(height, startIndex, count)
	}
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) identityEpochsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityEpochsOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityEpochsOld(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) identityFlipsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityFlipsOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityFlipsOld(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) identityInvitesOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityInvitesOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityInvitesOld(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) identityTxsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityTxsOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityTxsOld(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) identityRewardsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityRewardsOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityRewardsOld(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) identityEpochRewardsOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("identityEpochRewardsOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.IdentityEpochRewardsOld(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) addressPenaltiesOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("addressPenaltiesOld", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.AddressPenaltiesOld(vars["address"], startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) balancesOld(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("balancesOld", r.RequestURI)
	defer s.pm.Complete(id)

	startIndex, count, err := server.ReadOldPaginatorParams(mux.Vars(r))
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.BalancesOld(startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}

// Deprecated
func (s *httpServer) epochRewards(w http.ResponseWriter, r *http.Request) {
	id := s.pm.Start("epochRewards", r.RequestURI)
	defer s.pm.Complete(id)

	vars := mux.Vars(r)
	epoch, err := server.ReadUint(vars, "epoch")
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	startIndex, count, err := server.ReadOldPaginatorParams(vars)
	if err != nil {
		server.WriteErrorResponse(w, err, s.log)
		return
	}
	resp, err := s.db.EpochRewards(epoch, startIndex, count)
	server.WriteResponse(w, resp, err, s.log)
}
