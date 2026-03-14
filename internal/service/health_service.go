package service

import (
	"context"
	"time"

	"github.com/Boyeep/flippy-backend/internal/config"
	"github.com/Boyeep/flippy-backend/internal/domain"
)

type pinger interface {
	Ping(context.Context) error
}

type HealthService struct {
	cfg config.Config
	db  pinger
}

func NewHealthService(cfg config.Config, db pinger) HealthService {
	return HealthService{cfg: cfg, db: db}
}

func (s HealthService) Status() domain.HealthStatus {
	database := "unavailable"
	if s.db != nil && s.db.Ping(context.Background()) == nil {
		database = "ok"
	}

	return domain.HealthStatus{
		Name:      "flippy-backend",
		Env:       s.cfg.App.Env,
		Version:   s.cfg.App.Version,
		Timestamp: time.Now().UTC(),
		Database:  database,
	}
}
