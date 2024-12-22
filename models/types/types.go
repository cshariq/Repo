package types

import (
	"time"

	"github.com/hackclub/hackatime/models"
)

type SummaryRetriever func(f, t time.Time, u *models.User, filters *models.Filters) (*models.Summary, error)
