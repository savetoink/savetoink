package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/auth"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/types"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/utils"
)

// HandleGetSends handles the get sends quota status endpoint.
func (h *Handlers) HandleGetSends(w http.ResponseWriter, r *http.Request) {
	accountID := auth.GetAccountIDFromCtx(r.Context())
	startDate := time.Unix(0, 0)

	totalSends, err := h.service.CountSendsByAccountDateRange(r.Context(), accountID, startDate, time.Now())
	if err != nil {
		utils.HandleServiceError(w, r, err, "get sends count")
		return
	}

	if h.cfg.AuthBackend == consts.AuthBackendSharedAPIKey {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(types.SendsResponseNoLimits{
			TotalSends: totalSends,
		})
		return
	}

	startDate = time.Now().AddDate(0, 0, -consts.FreeTierSendPeriodDays)
	sendsCount, err := h.service.CountSendsByAccountDateRange(r.Context(), accountID, startDate, time.Now())
	if err != nil {
		utils.HandleServiceError(w, r, err, "get sends count")
		return
	}

	remainingSends := max(0, consts.MaxFreeTierSendsPerPeriod-sendsCount)
	maxSends := consts.MaxFreeTierSendsPerPeriod
	periodDays := consts.FreeTierSendPeriodDays

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.SendsResponse{
		TotalSends:        totalSends,
		CurrentSends:      sendsCount,
		MaxSendsPerPeriod: maxSends,
		PeriodDays:        periodDays,
		RemainingSends:    remainingSends,
	})
}
