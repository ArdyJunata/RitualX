package service_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockRoutineLogRepo struct{ mock.Mock }

func (m *mockRoutineLogRepo) Upsert(l *model.RoutineLog) error {
	args := m.Called(l)
	return args.Error(0)
}
func (m *mockRoutineLogRepo) FindByID(id uuid.UUID) (*model.RoutineLog, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RoutineLog), args.Error(1)
}
func (m *mockRoutineLogRepo) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type mockRoutineRepo2 struct{ mock.Mock }

func (m *mockRoutineRepo2) FindByIDAndUserID(id, userID uuid.UUID) (*model.Routine, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Routine), args.Error(1)
}

// ── Log ───────────────────────────────────────────────────────────────────────

func TestLog_Success(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()
	logID := uuid.New()

	logRepo := &mockRoutineLogRepo{}
	routineRepo := &mockRoutineRepo2{}

	routine := &model.Routine{ID: routineID, UserID: userID}
	routineRepo.On("FindByIDAndUserID", routineID, userID).Return(routine, nil)

	logRepo.On("Upsert", mock.AnythingOfType("*model.RoutineLog")).
		Run(func(args mock.Arguments) {
			l := args.Get(0).(*model.RoutineLog)
			l.ID = logID
		}).
		Return(nil)

	svc := service.NewRoutineLogServiceIface(logRepo, routineRepo)
	req := service.LogRoutineRequest{LoggedAt: "2026-07-15"}
	result, svcErr := svc.Log(userID, routineID, req)

	assert.Nil(t, svcErr)
	assert.Equal(t, logID, result.ID)
	assert.Equal(t, routineID, result.RoutineID)
}

func TestLog_RoutineNotFound(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()

	logRepo := &mockRoutineLogRepo{}
	routineRepo := &mockRoutineRepo2{}
	routineRepo.On("FindByIDAndUserID", routineID, userID).Return(nil, nil)

	svc := service.NewRoutineLogServiceIface(logRepo, routineRepo)
	req := service.LogRoutineRequest{LoggedAt: "2026-07-15"}
	_, svcErr := svc.Log(userID, routineID, req)

	assert.NotNil(t, svcErr)
	assert.Equal(t, "NOT_FOUND", svcErr.Code)
}

func TestLog_InvalidDate(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()

	logRepo := &mockRoutineLogRepo{}
	routineRepo := &mockRoutineRepo2{}

	routine := &model.Routine{ID: routineID, UserID: userID}
	routineRepo.On("FindByIDAndUserID", routineID, userID).Return(routine, nil)

	svc := service.NewRoutineLogServiceIface(logRepo, routineRepo)
	req := service.LogRoutineRequest{LoggedAt: "not-a-date"}
	_, svcErr := svc.Log(userID, routineID, req)

	assert.NotNil(t, svcErr)
	assert.Equal(t, "VALIDATION_ERROR", svcErr.Code)
}

func TestLog_CountDefault1WhenNil(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()

	logRepo := &mockRoutineLogRepo{}
	routineRepo := &mockRoutineRepo2{}

	routine := &model.Routine{ID: routineID, UserID: userID}
	routineRepo.On("FindByIDAndUserID", routineID, userID).Return(routine, nil)
	logRepo.On("Upsert", mock.MatchedBy(func(l *model.RoutineLog) bool {
		return l.Count == 1
	})).Return(nil)

	svc := service.NewRoutineLogServiceIface(logRepo, routineRepo)
	req := service.LogRoutineRequest{LoggedAt: "2026-07-15"} // Count nil → default 1
	result, svcErr := svc.Log(userID, routineID, req)

	assert.Nil(t, svcErr)
	assert.Equal(t, 1, result.Count)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDeleteLog_Success(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()
	logID := uuid.New()

	logRepo := &mockRoutineLogRepo{}
	routineRepo := &mockRoutineRepo2{}

	existingLog := &model.RoutineLog{ID: logID, RoutineID: routineID, UserID: userID}
	logRepo.On("FindByID", logID).Return(existingLog, nil)
	logRepo.On("Delete", logID).Return(nil)

	svc := service.NewRoutineLogServiceIface(logRepo, routineRepo)
	svcErr := svc.Delete(userID, routineID, logID)

	assert.Nil(t, svcErr)
}

func TestDeleteLog_NotFound(t *testing.T) {
	userID := uuid.New()
	routineID := uuid.New()
	logID := uuid.New()

	logRepo := &mockRoutineLogRepo{}
	routineRepo := &mockRoutineRepo2{}
	logRepo.On("FindByID", logID).Return(nil, nil)

	svc := service.NewRoutineLogServiceIface(logRepo, routineRepo)
	svcErr := svc.Delete(userID, routineID, logID)

	assert.NotNil(t, svcErr)
	assert.Equal(t, "NOT_FOUND", svcErr.Code)
}

func TestDeleteLog_WrongUser(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	routineID := uuid.New()
	logID := uuid.New()

	logRepo := &mockRoutineLogRepo{}
	routineRepo := &mockRoutineRepo2{}

	existingLog := &model.RoutineLog{ID: logID, RoutineID: routineID, UserID: otherUserID}
	logRepo.On("FindByID", logID).Return(existingLog, nil)

	svc := service.NewRoutineLogServiceIface(logRepo, routineRepo)
	svcErr := svc.Delete(userID, routineID, logID)

	assert.NotNil(t, svcErr)
	assert.Equal(t, "FORBIDDEN", svcErr.Code)
}
