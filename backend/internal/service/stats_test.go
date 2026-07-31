package service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockStatsLogRepo struct{ mock.Mock }

func (m *mockStatsLogRepo) ListByRoutineAndDateRange(routineID uuid.UUID, from, to time.Time) ([]model.RoutineLog, error) {
	args := m.Called(routineID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.RoutineLog), args.Error(1)
}
func (m *mockStatsLogRepo) SumCountByRoutineAndDateRange(routineID uuid.UUID, from, to time.Time) (int, error) {
	args := m.Called(routineID, from, to)
	return args.Int(0), args.Error(1)
}

type mockStatsRoutineRepo struct{ mock.Mock }

func (m *mockStatsRoutineRepo) FindByIDAndUserID(id, userID uuid.UUID) (*model.Routine, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Routine), args.Error(1)
}

// ── CurrentPeriodRange (pure) ─────────────────────────────────────────────────

func TestCurrentPeriodRange_Daily(t *testing.T) {
	from, to := service.CurrentPeriodRange("daily", streakTestDate("2026-07-15"))
	assert.Equal(t, streakTestDate("2026-07-15"), from)
	assert.Equal(t, streakTestDate("2026-07-15"), to)
}

func TestCurrentPeriodRange_Weekly(t *testing.T) {
	// 2026-07-15 is a Wednesday → week Mon 07-13 .. Sun 07-19
	from, to := service.CurrentPeriodRange("weekly", streakTestDate("2026-07-15"))
	assert.Equal(t, streakTestDate("2026-07-13"), from)
	assert.Equal(t, streakTestDate("2026-07-19"), to)
}

func TestCurrentPeriodRange_Monthly(t *testing.T) {
	from, to := service.CurrentPeriodRange("monthly", streakTestDate("2026-07-15"))
	assert.Equal(t, streakTestDate("2026-07-01"), from)
	assert.Equal(t, streakTestDate("2026-07-31"), to)
}

// ── GetHeatmap ────────────────────────────────────────────────────────────────

func TestGetHeatmap_BuildsMap(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()
	lr := new(mockStatsLogRepo)
	rr := new(mockStatsRoutineRepo)

	rr.On("FindByIDAndUserID", routineID, userID).Return(&model.Routine{ID: routineID, UserID: userID}, nil)
	lr.On("ListByRoutineAndDateRange", routineID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return([]model.RoutineLog{
			{LoggedAt: streakTestDate("2026-03-05"), Count: 2},
			{LoggedAt: streakTestDate("2026-03-06"), Count: 1},
		}, nil)

	svc := service.NewStatsServiceIface(lr, rr)
	m, err := svc.GetHeatmap(userID, routineID, 2026)

	assert.Nil(t, err)
	assert.Equal(t, 2, m["2026-03-05"])
	assert.Equal(t, 1, m["2026-03-06"])
	assert.Len(t, m, 2)
}

func TestGetHeatmap_RoutineNotFound(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()
	lr := new(mockStatsLogRepo)
	rr := new(mockStatsRoutineRepo)

	rr.On("FindByIDAndUserID", routineID, userID).Return(nil, nil)

	svc := service.NewStatsServiceIface(lr, rr)
	m, err := svc.GetHeatmap(userID, routineID, 2026)

	assert.Nil(t, m)
	assert.NotNil(t, err)
	assert.Equal(t, "NOT_FOUND", err.Code)
}

// ── GetProgress ───────────────────────────────────────────────────────────────

func TestGetProgress_Weekly(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()
	lr := new(mockStatsLogRepo)
	rr := new(mockStatsRoutineRepo)

	rr.On("FindByIDAndUserID", routineID, userID).
		Return(&model.Routine{ID: routineID, UserID: userID, PeriodType: "weekly", TargetCount: 5}, nil)
	lr.On("SumCountByRoutineAndDateRange", routineID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(4, nil)

	svc := service.NewStatsServiceIface(lr, rr)
	p, err := svc.GetProgress(userID, routineID)

	assert.Nil(t, err)
	assert.Equal(t, 4, p.Completed)
	assert.Equal(t, 5, p.Target)
	assert.Equal(t, "weekly", p.Period)
}

func TestGetProgress_RoutineNotFound(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()
	lr := new(mockStatsLogRepo)
	rr := new(mockStatsRoutineRepo)

	rr.On("FindByIDAndUserID", routineID, userID).Return(nil, nil)

	svc := service.NewStatsServiceIface(lr, rr)
	p, err := svc.GetProgress(userID, routineID)

	assert.Nil(t, p)
	assert.NotNil(t, err)
	assert.Equal(t, "NOT_FOUND", err.Code)
}
