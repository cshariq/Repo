package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hackclub/hackatime/helpers"
	routeutils "github.com/hackclub/hackatime/routes/utils"

	conf "github.com/hackclub/hackatime/config"
	"github.com/hackclub/hackatime/middlewares"
	"github.com/hackclub/hackatime/services"
)

type SummaryApiHandler struct {
	config      *conf.Config
	userSrvc    services.IUserService
	summarySrvc services.ISummaryService
}

func NewSummaryApiHandler(userService services.IUserService, summaryService services.ISummaryService) *SummaryApiHandler {
	return &SummaryApiHandler{
		summarySrvc: summaryService,
		userSrvc:    userService,
		config:      conf.Get(),
	}
}

func (h *SummaryApiHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
	r.Get("/", h.Get)

	router.Mount("/summary", r)
}

// @Summary Retrieve a summary
// @ID get-summary
// @Tags summary
// @Produce json
// @Param interval query string false "Interval identifier" Enums(today, yesterday, week, month, year, 7_days, last_7_days, 30_days, last_30_days, 6_months, last_6_months, 12_months, last_12_months, last_year, any, all_time, low_skies, high_seas)
// @Param from query string false "Start date (e.g. '2021-02-07')"
// @Param to query string false "End date (e.g. '2021-02-08')"
// @Param recompute query bool false "Whether to recompute the summary from raw heartbeat or use cache"
// @Param project query string false "Project to filter by"
// @Param language query string false "Language to filter by"
// @Param editor query string false "Editor to filter by"
// @Param operating_system query string false "OS to filter by"
// @Param machine query string false "Machine to filter by"
// @Param label query string false "Project label to filter by"
// @Param user query string false "The user to filter by if using Bearer authentication and the admin token"
// @Security ApiKeyAuth
// @Success 200 {object} models.Summary
// @Router /summary [get]
func (h *SummaryApiHandler) Get(w http.ResponseWriter, r *http.Request) {
	summary, err, status := routeutils.LoadUserSummary(h.summarySrvc, r)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	helpers.RespondJSON(w, r, http.StatusOK, summary)
}
