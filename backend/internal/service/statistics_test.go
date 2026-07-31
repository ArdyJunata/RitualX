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

type mockStatisticsLogRepo struct{ mock.Mock }

func (m *mockStatisticsLogRepo) SumCountByUser(userID uuid.UUID) (int, error) {
	a := m.Called(userID)
	return a.Int(0), a.Error(1)
}
func (m *mockStatisticsLogRepo) SumCountByUserAndDateRange(userID uuid.UUID, from, to time.Time) (int, error) {
	a := m.Called(userID, from, to)
	return a.Int(0), a.Error(1)
}
func (m *mockStatisticsLogRepo) SumCountByRoutine(routineID uuid.UUID) (int, error) {
	a := m.Called(routineID)
	return a.Int(0), a.Error(1)
}
func (m *mockStatisticsLogRepo) CountByRoutine(routineID uuid.UUID) (int, error) {
	a := m.Called(routineID)
	return a.Int(0), a.Error(1)
}

type mockStatisticsStreakRepo struct{ mock.Mock }

func (m *mockStatisticsStreakRepo) FindByUserID(userID uuid.UUID) ([]model.Streak, error) {
	a := m.Called(userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]model.Streak), a.Error(1)
}
func (m *mockStatisticsStreakRepo) FindByRoutineID(routineID uuid.UUID) (*model.Streak, error) {
	a := m.Called(routineID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*model.Streak), a.Error(1)
}

type mockStatisticsRoutineRepo struct{ mock.Mock }

func (m *mockStatisticsRoutineRepo) FindByIDAndUserID(id, userID uuid.UUID) (*model.Routine, error) {
	a := m.Called(id, userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*model.Routine), a.Error(1)
}
func (m *mockStatisticsRoutineRepo) CountActiveByUserID(userID uuid.UUID) (int, error) {
	a := m.Called(userID)
	return a.Int(0), a.Error(1)
}

type mockStatisticsUserRepo struct{ mock.Mock }

func (m *mockStatisticsUserRepo) FindByID(id uuid.UUID) (*model.User, error) {
	a := m.Called(id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*model.User), a.Error(1)
}

func newStatisticsSvc() (*service.StatisticsService, *mockStatisticsLogRepo, *mockStatisticsStreakRepo, *mockStatisticsRoutineRepo, *mockStatisticsUserRepo) {
	lr := new(mockStatisticsLogRepo)
	sr := new(mockStatisticsStreakRepo)
	rr := new(mockStatisticsRoutineRepo)
	ur := new(mockStatisticsUserRepo)
	return service.NewStatisticsServiceIface(lr, sr, rr, ur), lr, sr, rr, ur
}

// ── Overview ──────────────────────────────────────────────────────────────────

func TestStatistics_Overview(t *testing.T) {
	userID := uuid.New()
	svc, lr, sr, rr, ur := newStatisticsSvc()

	ur.On("FindByID", userID).Return(&model.User{Level: 3, XP: 150}, nil)
	lr.On("SumCountByUser", userID).Return(42, nil)
	rr.On("CountActiveByUserID", userID).Return(5, nil)
	sr.On("FindByUserID", userID).Return([]model.Streak{
		{CurrentStreak: 2, LongestStreak: 4},
		{CurrentStreak: 0, LongestStreak: 7},
	}, nil)

	out, err := svc.GetOverview(userID)

	assert.Nil(t, err)
	assert.Equal(t, 42, out.TotalCompletions)
	assert.Equal(t, 5, out.ActiveRoutines)
	assert.Equal(t, 1, out.ActiveStreaks) // only current>0
	assert.Equal(t, 7, out.LongestStreak) // max longest
	assert.Equal(t, 3, out.Level)
	assert.Equal(t, 150, out.XP)
}

func TestStatistics_OverviewUserNotFound(t *testing.T) {
	userID := uuid.New()
	svc, _, _, _, ur := newStatisticsSvc()
	ur.On("FindByID", userID).Return(nil, nil)

	out, err := svc.GetOverview(userID)

	assert.Nil(t, out)
	assert.NotNil(t, err)
	assert.Equal(t, "NOT_FOUND", err.Code)
}

// ── Routine stats ─────────────────────────────────────────────────────────────

func TestStatistics_RoutineStats(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()
	svc, lr, sr, rr, _ := newStatisticsSvc()

	created := time.Now().UTC().AddDate(0, 0, -9) // 10 active days inclusive
	rr.On("FindByIDAndUserID", routineID, userID).
		Return(&model.Routine{ID: routineID, UserID: userID, CreatedAt: created}, nil)
	lr.On("SumCountByRoutine", routineID).Return(12, nil)
	lr.On("CountByRoutine", routineID).Return(5, nil)
	last := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	sr.On("FindByRoutineID", routineID).
		Return(&model.Streak{CurrentStreak: 3, LongestStreak: 6, LastCompleted: &last}, nil)

	out, err := svc.GetRoutineStats(userID, routineID)

	assert.Nil(t, err)
	assert.Equal(t, 12, out.TotalCompletions)
	assert.Equal(t, 5, out.DaysLogged)
	assert.Equal(t, 3, out.CurrentStreak)
	assert.Equal(t, 6, out.LongestStreak)
	assert.Equal(t, 0.5, out.CompletionRate) // 5 / 10
	assert.Equal(t, last, *out.LastCompleted)
}

func TestStatistics_RoutineStatsNotFound(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()
	svc, _, _, rr, _ := newStatisticsSvc()
	rr.On("FindByIDAndUserID", routineID, userID).Return(nil, nil)

	out, err := svc.GetRoutineStats(userID, routineID)

	assert.Nil(t, out)
	assert.NotNil(t, err)
	assert.Equal(t, "NOT_FOUND", err.Code)
}

func TestStatistics_RoutineStatsNilStreak(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()
	svc, lr, sr, rr, _ := newStatisticsSvc()

	rr.On("FindByIDAndUserID", routineID, userID).
		Return(&model.Routine{ID: routineID, UserID: userID, CreatedAt: time.Now().UTC()}, nil)
	lr.On("SumCountByRoutine", routineID).Return(0, nil)
	lr.On("CountByRoutine", routineID).Return(0, nil)
	sr.On("FindByRoutineID", routineID).Return(nil, nil)

	out, err := svc.GetRoutineStats(userID, routineID)

	assert.Nil(t, err)
	assert.Equal(t, 0, out.CurrentStreak)
	assert.Equal(t, 0, out.LongestStreak)
	assert.Nil(t, out.LastCompleted)
	assert.Equal(t, float64(0), out.CompletionRate)
}

// ── Summaries ─────────────────────────────────────────────────────────────────

func TestStatistics_WeeklySummary(t *testing.T) {
	userID := uuid.New()
	svc, lr, _, _, _ := newStatisticsSvc()

	// current period queried first, then previous
	lr.On("SumCountByUserAndDateRange", userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(10, nil).Once()
	lr.On("SumCountByUserAndDateRange", userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(6, nil).Once()

	out, err := svc.GetWeeklySummary(userID)

	assert.Nil(t, err)
	assert.Equal(t, 10, out.This.Completions)
	assert.Equal(t, 6, out.Last.Completions)
	assert.Equal(t, 4, out.Delta)
}

func TestStatistics_MonthlySummary(t *testing.T) {
	userID := uuid.New()
	svc, lr, _, _, _ := newStatisticsSvc()

	lr.On("SumCountByUserAndDateRange", userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(8, nil).Once()
	lr.On("SumCountByUserAndDateRange", userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(5, nil).Once()

	out, err := svc.GetMonthlySummary(userID)

	assert.Nil(t, err)
	assert.Equal(t, 8, out.This.Completions)
	assert.Equal(t, 5, out.Last.Completions)
	assert.Equal(t, 3, out.Delta)
}
