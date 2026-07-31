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

func streakTestDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

// ── CalculateStreak (pure) ────────────────────────────────────────────────────

func TestCalculateStreak_DailyConsecutive(t *testing.T) {
	cur, longest, last := service.CalculateStreak("daily", []time.Time{
		streakTestDate("2026-07-01"), streakTestDate("2026-07-02"), streakTestDate("2026-07-03"),
	})
	assert.Equal(t, 3, cur)
	assert.Equal(t, 3, longest)
	assert.Equal(t, streakTestDate("2026-07-03"), *last)
}

func TestCalculateStreak_DailyGapResetsCurrent(t *testing.T) {
	// runs: [07-01,07-02]=2, [07-04]=1 → longest 2, current 1 (ends at latest)
	cur, longest, _ := service.CalculateStreak("daily", []time.Time{
		streakTestDate("2026-07-01"), streakTestDate("2026-07-02"), streakTestDate("2026-07-04"),
	})
	assert.Equal(t, 1, cur)
	assert.Equal(t, 2, longest)
}

func TestCalculateStreak_UnsortedAndDuplicates(t *testing.T) {
	cur, longest, last := service.CalculateStreak("daily", []time.Time{
		streakTestDate("2026-07-03"), streakTestDate("2026-07-01"), streakTestDate("2026-07-02"), streakTestDate("2026-07-03"),
	})
	assert.Equal(t, 3, cur)
	assert.Equal(t, 3, longest)
	assert.Equal(t, streakTestDate("2026-07-03"), *last)
}

func TestCalculateStreak_Empty(t *testing.T) {
	cur, longest, last := service.CalculateStreak("daily", nil)
	assert.Equal(t, 0, cur)
	assert.Equal(t, 0, longest)
	assert.Nil(t, last)
}

func TestCalculateStreak_SingleLog(t *testing.T) {
	cur, longest, last := service.CalculateStreak("daily", []time.Time{streakTestDate("2026-07-10")})
	assert.Equal(t, 1, cur)
	assert.Equal(t, 1, longest)
	assert.Equal(t, streakTestDate("2026-07-10"), *last)
}

func TestCalculateStreak_WeeklyConsecutive(t *testing.T) {
	// 07-01 (wk of 06-29), 07-08 (wk of 07-06), 07-15 (wk of 07-13) → 3 consecutive weeks
	cur, longest, _ := service.CalculateStreak("weekly", []time.Time{
		streakTestDate("2026-07-01"), streakTestDate("2026-07-08"), streakTestDate("2026-07-15"),
	})
	assert.Equal(t, 3, cur)
	assert.Equal(t, 3, longest)
}

func TestCalculateStreak_WeeklySameWeekCountsOnce(t *testing.T) {
	// 07-06 (Mon) and 07-08 (Wed) are the same ISO week → 1 period
	cur, longest, _ := service.CalculateStreak("weekly", []time.Time{
		streakTestDate("2026-07-06"), streakTestDate("2026-07-08"),
	})
	assert.Equal(t, 1, cur)
	assert.Equal(t, 1, longest)
}

func TestCalculateStreak_WeeklyGap(t *testing.T) {
	// wk of 06-29 then wk of 07-13 (a week skipped) → not consecutive
	cur, longest, _ := service.CalculateStreak("weekly", []time.Time{
		streakTestDate("2026-07-01"), streakTestDate("2026-07-15"),
	})
	assert.Equal(t, 1, cur)
	assert.Equal(t, 1, longest)
}

func TestCalculateStreak_MonthlyConsecutive(t *testing.T) {
	cur, longest, _ := service.CalculateStreak("monthly", []time.Time{
		streakTestDate("2026-01-15"), streakTestDate("2026-02-20"), streakTestDate("2026-03-01"),
	})
	assert.Equal(t, 3, cur)
	assert.Equal(t, 3, longest)
}

func TestCalculateStreak_MonthlyGap(t *testing.T) {
	cur, longest, _ := service.CalculateStreak("monthly", []time.Time{
		streakTestDate("2026-01-15"), streakTestDate("2026-03-20"),
	})
	assert.Equal(t, 1, cur)
	assert.Equal(t, 1, longest)
}

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockStreakRepo struct{ mock.Mock }

func (m *mockStreakRepo) Upsert(s *model.Streak) error {
	args := m.Called(s)
	return args.Error(0)
}
func (m *mockStreakRepo) FindByRoutineID(routineID uuid.UUID) (*model.Streak, error) {
	args := m.Called(routineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Streak), args.Error(1)
}
func (m *mockStreakRepo) FindByUserID(userID uuid.UUID) ([]model.Streak, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Streak), args.Error(1)
}

type mockStreakLogRepo struct{ mock.Mock }

func (m *mockStreakLogRepo) ListDatesByRoutine(routineID uuid.UUID) ([]time.Time, error) {
	args := m.Called(routineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]time.Time), args.Error(1)
}

type mockStreakRoutineRepo struct{ mock.Mock }

func (m *mockStreakRoutineRepo) FindByIDAndUserID(id, userID uuid.UUID) (*model.Routine, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Routine), args.Error(1)
}

// ── StreakService ─────────────────────────────────────────────────────────────

func TestRecalculate_Success(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()

	sr := new(mockStreakRepo)
	lr := new(mockStreakLogRepo)
	rr := new(mockStreakRoutineRepo)

	rr.On("FindByIDAndUserID", routineID, userID).Return(&model.Routine{ID: routineID, UserID: userID, PeriodType: "daily"}, nil)
	lr.On("ListDatesByRoutine", routineID).Return([]time.Time{
		streakTestDate("2026-07-01"), streakTestDate("2026-07-02"), streakTestDate("2026-07-03"),
	}, nil)
	sr.On("FindByRoutineID", routineID).Return(nil, nil)
	sr.On("Upsert", mock.AnythingOfType("*model.Streak")).Return(nil)

	svc := service.NewStreakServiceIface(sr, lr, rr)
	streak, err := svc.Recalculate(userID, routineID)

	assert.Nil(t, err)
	assert.Equal(t, 3, streak.CurrentStreak)
	assert.Equal(t, 3, streak.LongestStreak)
	sr.AssertCalled(t, "Upsert", mock.AnythingOfType("*model.Streak"))
}

func TestRecalculate_LongestIsHighWaterMark(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()

	sr := new(mockStreakRepo)
	lr := new(mockStreakLogRepo)
	rr := new(mockStreakRoutineRepo)

	rr.On("FindByIDAndUserID", routineID, userID).Return(&model.Routine{ID: routineID, UserID: userID, PeriodType: "daily"}, nil)
	// current logs give only a run of 1
	lr.On("ListDatesByRoutine", routineID).Return([]time.Time{streakTestDate("2026-07-10")}, nil)
	// but a previous longest of 5 exists → must be preserved
	sr.On("FindByRoutineID", routineID).Return(&model.Streak{LongestStreak: 5}, nil)
	sr.On("Upsert", mock.AnythingOfType("*model.Streak")).Return(nil)

	svc := service.NewStreakServiceIface(sr, lr, rr)
	streak, err := svc.Recalculate(userID, routineID)

	assert.Nil(t, err)
	assert.Equal(t, 1, streak.CurrentStreak)
	assert.Equal(t, 5, streak.LongestStreak)
}

func TestRecalculate_RoutineNotFound(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()

	sr := new(mockStreakRepo)
	lr := new(mockStreakLogRepo)
	rr := new(mockStreakRoutineRepo)

	rr.On("FindByIDAndUserID", routineID, userID).Return(nil, nil)

	svc := service.NewStreakServiceIface(sr, lr, rr)
	streak, err := svc.Recalculate(userID, routineID)

	assert.Nil(t, streak)
	assert.NotNil(t, err)
	assert.Equal(t, "NOT_FOUND", err.Code)
}

func TestGetByRoutine_ZeroWhenNone(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()

	sr := new(mockStreakRepo)
	lr := new(mockStreakLogRepo)
	rr := new(mockStreakRoutineRepo)

	rr.On("FindByIDAndUserID", routineID, userID).Return(&model.Routine{ID: routineID, UserID: userID, PeriodType: "daily"}, nil)
	sr.On("FindByRoutineID", routineID).Return(nil, nil)

	svc := service.NewStreakServiceIface(sr, lr, rr)
	streak, err := svc.GetByRoutine(userID, routineID)

	assert.Nil(t, err)
	assert.Equal(t, 0, streak.CurrentStreak)
	assert.Equal(t, 0, streak.LongestStreak)
	assert.Equal(t, routineID, streak.RoutineID)
}

func TestListByUser_Success(t *testing.T) {
	userID := uuid.New()

	sr := new(mockStreakRepo)
	lr := new(mockStreakLogRepo)
	rr := new(mockStreakRoutineRepo)

	sr.On("FindByUserID", userID).Return([]model.Streak{{CurrentStreak: 2, LongestStreak: 4}}, nil)

	svc := service.NewStreakServiceIface(sr, lr, rr)
	streaks, err := svc.ListByUser(userID)

	assert.Nil(t, err)
	assert.Len(t, streaks, 1)
	assert.Equal(t, 2, streaks[0].CurrentStreak)
}
