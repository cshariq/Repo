package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	conf "github.com/hackclub/hackatime/config"
	"github.com/hackclub/hackatime/middlewares"
	"github.com/hackclub/hackatime/models"
	"github.com/hackclub/hackatime/models/view"
	routeutils "github.com/hackclub/hackatime/routes/utils"
	"github.com/hackclub/hackatime/services"
	"github.com/hackclub/hackatime/utils"
)

type ProjectsHandler struct {
	config           *conf.Config
	userService      services.IUserService
	heartbeatService services.IHeartbeatService
}

func NewProjectsHandler(userService services.IUserService, heartbeatService services.IHeartbeatService) *ProjectsHandler {
	return &ProjectsHandler{
		config:           conf.Get(),
		userService:      userService,
		heartbeatService: heartbeatService,
	}
}

func (h *ProjectsHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userService).
			WithRedirectTarget(defaultErrorRedirectTarget()).
			WithRedirectErrorMessage("unauthorized").Handler,
	)
	r.Get("/", h.GetIndex)

	router.Mount("/projects", r)
}

func (h *ProjectsHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
	if err := templates[conf.ProjectsTemplate].Execute(w, h.buildViewModel(r, w)); err != nil {
		conf.Log().Request(r).Error("failed to get projects page", "error", err)
	}
}

func (h *ProjectsHandler) buildViewModel(r *http.Request, w http.ResponseWriter) *view.ProjectsViewModel {
	user := middlewares.GetPrincipal(r)
	if user == nil { // this should actually never occur, because of auth middleware
		w.WriteHeader(http.StatusUnauthorized)
		return h.buildViewModel(r, w).WithError("unauthorized")
	}

	pageParams := utils.ParsePageParamsWithDefault(r, 1, 24)
	// note: pagination is not fully implemented, yet
	// count function to get total item / total pages is missing
	// and according ui (+ optionally search bar) is missing, too

	var err error
	var projects []*models.ProjectStats

	projects, err = h.heartbeatService.GetUserProjectStats(user, time.Time{}, utils.BeginOfToday(time.Local), pageParams, false)
	if err != nil {
		conf.Log().Request(r).Error("error while fetching project stats", "userID", user.ID, "error", err)
		return &view.ProjectsViewModel{
			SharedLoggedInViewModel: view.SharedLoggedInViewModel{
				SharedViewModel: view.NewSharedViewModel(h.config, &view.Messages{Error: criticalError}),
				User:            user,
				ApiKey:          user.ApiKey,
			},
		}
	}

	vm := &view.ProjectsViewModel{
		SharedLoggedInViewModel: view.SharedLoggedInViewModel{
			SharedViewModel: view.NewSharedViewModel(h.config, nil),
			User:            user,
			ApiKey:          user.ApiKey,
		},
		Projects:   projects,
		PageParams: pageParams,
	}
	return routeutils.WithSessionMessages(vm, r, w)
}