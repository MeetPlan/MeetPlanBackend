package genetic

import (
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"go.uber.org/zap"
)

type protonImpl struct {
	db     sql.SQL
	config ProtonConfig
	logger *zap.SugaredLogger
}

type Proton interface {
	AssembleTimetable() ([]Meeting, error)
}

func NewGeneticProton(db sql.SQL, logger *zap.SugaredLogger) (Proton, error) {
	protonConfig, err := LoadConfig()
	return &protonImpl{db: db, config: protonConfig, logger: logger}, err
}
