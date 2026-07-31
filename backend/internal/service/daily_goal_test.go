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

// ── Mock ──────────────────────────────────────────────────────────────────────

type mockDailyGoalRepo struct{ mock.Mock }

func (m *mockDailyGoalRepo) Upsert(g *model.DailyGoal) error {
	args := m.Called(g)
	return args.Error(0)
}
func (m *mockDailyGoalRepo) CountActiveRoutines(userID uuid.UUID) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}
func (m *mockDailyGoalRepo) CountCompletedRoutines(userID uuid.UUID, date time.Time) (int, error) {
	args := m.Called(userID, date)
	return args.Int(0), args.Error(1)
}
func (m *mockDailyGoalRepo) FindByUserAndDateRange(userID uuid.UUID, from, to time.Time) ([]model.DailyGoal, error) {
	args := m.Called(userID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.DailyGoal), args.Error(1)
}

// ── DailyGoalService ──────────────────────────────────────────────────────────

func TestDailyGoal_RecalculateAchieved(t *testing.T) {
	userID := uuid.New()
	r := new(mockDailyGoalRepo)
	r.On("CountActiveRoutines", userID).Return(3, nil)
	r.On("CountCompletedRoutines", userID, mock.AnythingOfType("time.Time")).Return(3, nil)
	r.On("Upsert", mock.AnythingOfType("*model.DailyGoal")).Return(nil)

	svc := service.NewDailyGoalServiceIface(r)
	goal, err := svc.Recalculate(userID, time.Now())

	assert.Nil(t, err)
	assert.Equal(t, 3, goal.TotalRoutines)
	assert.Equal(t, 3, goal.Completed)
	assert.True(t, goal.IsAchieved)
	r.AssertCalled(t, "Upsert", mock.AnythingOfType("*model.DailyGoal"))
}

func TestDailyGoal_RecalculateNotAchieved(t *testing.T) {
	userID := uuid.New()
	r := new(mockDailyGoalRepo)
	r.On("CountActiveRoutines", userID).Return(3, nil)
	r.On("CountCompletedRoutines", userID, mock.AnythingOfType("time.Time")).Return(1, nil)
	r.On("Upsert", mock.AnythingOfType("*model.DailyGoal")).Return(nil)

	svc := service.NewDailyGoalServiceIface(r)
	goal, err := svc.Recalculate(userID, time.Now())

	assert.Nil(t, err)
	assert.Equal(t, 1, goal.Completed)
	assert.False(t, goal.IsAchieved)
}

func TestDailyGoal_RecalculateNoRoutines(t *testing.T) {
	userID := uuid.New()
	r := new(mockDailyGoalRepo)
	r.On("CountActiveRoutines", userID).Return(0, nil)
	r.On("CountCompletedRoutines", userID, mock.AnythingOfType("time.Time")).Return(0, nil)
	r.On("Upsert", mock.AnythingOfType("*model.DailyGoal")).Return(nil)

	svc := service.NewDailyGoalServiceIface(r)
	goal, err := svc.Recalculate(userID, time.Now())

	assert.Nil(t, err)
	assert.Equal(t, 0, goal.TotalRoutines)
	assert.False(t, goal.IsAchieved) // total 0 → never achieved
}

func TestDailyGoal_GetTodayComputesLiveNoWrite(t *testing.T) {
	userID := uuid.New()
	r := new(mockDailyGoalRepo)
	r.On("CountActiveRoutines", userID).Return(2, nil)
	r.On("CountCompletedRoutines", userID, mock.AnythingOfType("time.Time")).Return(1, nil)

	svc := service.NewDailyGoalServiceIface(r)
	goal, err := svc.GetToday(userID)

	assert.Nil(t, err)
	assert.Equal(t, 2, goal.TotalRoutines)
	assert.Equal(t, 1, goal.Completed)
	assert.False(t, goal.IsAchieved)
	r.AssertNotCalled(t, "Upsert", mock.Anything) // reads must not write
}

func TestDailyGoal_GetHistory(t *testing.T) {
	userID := uuid.New()
	r := new(mockDailyGoalRepo)
	r.On("FindByUserAndDateRange", userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return([]model.DailyGoal{{TotalRoutines: 3, Completed: 2, IsAchieved: false}}, nil)

	svc := service.NewDailyGoalServiceIface(r)
	goals, err := svc.GetHistory(userID, time.Now().AddDate(0, 0, -7), time.Now())

	assert.Nil(t, err)
	assert.Len(t, goals, 1)
	assert.Equal(t, 3, goals[0].TotalRoutines)
	assert.Equal(t, 2, goals[0].Completed)
}
