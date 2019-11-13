package throttler

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"time"

	"testing"
)

var (
	testTime, _ = time.Parse(time.RFC3339, "2019-01-01T12:34:56Z")
	period      = 5 * time.Minute
	sinceTime   = testTime.Add(-period)
)

type testClock struct {
	ConstTime time.Time
}

func (c *testClock) Now() time.Time {
	return c.ConstTime
}

type testAction struct {
	mock.Mock
}

func (a *testAction) GroupID() string {
	args := a.Called()
	return args.String(0)
}

func (a *testAction) Run() error {
	args := a.Called()
	return args.Error(0)
}

type testStorage struct {
	mock.Mock
}

func (s *testStorage) CountLastExecuted(action Action, after time.Time) (int, error) {
	args := s.Called(action, after)
	return args.Int(0), args.Error(1)
}

func (s *testStorage) SaveSuccessfulExecution(action Action, at time.Time, expiration time.Duration) error {
	args := s.Called(action, at, expiration)
	return args.Error(0)
}

func TestExecute_StorageCountError(t *testing.T) {
	storage := &testStorage{}
	storage.On("CountLastExecuted", mock.Anything, mock.Anything).Return(0, errors.New("something went wrong"))

	throttler := New(5, time.Minute, storage)

	err := throttler.Execute(&testAction{})

	assert.EqualError(t, err, "error querying the storage: something went wrong")
}

func TestExecute_StorageSaveError(t *testing.T) {
	actionsCountInStorage := 0
	maxAllowedActions := 1

	action := &testAction{}
	action.On("Run").Return(nil).Once()

	storage := &testStorage{}
	storage.On("CountLastExecuted", action, mock.Anything).Return(actionsCountInStorage, nil)
	storage.On("SaveSuccessfulExecution", action, mock.Anything, mock.Anything).Return(errors.New("something went wrong"))

	throttler := New(maxAllowedActions, time.Minute, storage)

	err := throttler.Execute(action)

	assert.EqualError(t, err, "error while storing successful execution: something went wrong")
	action.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestExecute_VerifyLimit_NotExceedingLimit(t *testing.T) {
	clock := &testClock{ConstTime: testTime}

	actionsCountInStorage := 1
	maxAllowedActions := 1

	action := &testAction{}

	storage := &testStorage{}
	storage.On("CountLastExecuted", action, mock.Anything).Return(actionsCountInStorage, nil)

	throttler := New(maxAllowedActions, time.Minute, storage)
	throttler.clock = clock

	err := throttler.Execute(action)

	assert.Error(t, err)
	action.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestExecute_VerifyLimit_ExceedingLimit(t *testing.T) {
	clock := &testClock{ConstTime: testTime}

	actionsCountInStorage := 0
	maxAllowedActions := 1

	action := &testAction{}
	action.On("Run").Return(nil).Once()

	storage := &testStorage{}
	storage.On("CountLastExecuted", action, sinceTime).Return(actionsCountInStorage, nil)
	storage.On("SaveSuccessfulExecution", action, testTime, period).Return(nil)

	throttler := New(maxAllowedActions, period, storage)
	throttler.clock = clock

	err := throttler.Execute(action)

	assert.NoError(t, err)
	action.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestExecute_VerifyLimit_NotExceedingLimit_ActionRunError(t *testing.T) {
	actionsCountInStorage := 0
	maxAllowedActions := 1

	action := &testAction{}
	action.On("Run").Return(errors.New("some error")).Once()

	storage := &testStorage{}
	storage.On("CountLastExecuted", action, mock.Anything).Return(actionsCountInStorage, nil)

	throttler := New(maxAllowedActions, period, storage)

	err := throttler.Execute(action)

	assert.EqualError(t, err, "some error")
	action.AssertExpectations(t)
	storage.AssertExpectations(t)
}
