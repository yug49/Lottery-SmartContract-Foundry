// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/smartcontractkit/chainlink-common/pkg/types"
)

// CCIPTransactionStatusChecker is an autogenerated mock type for the CCIPTransactionStatusChecker type
type CCIPTransactionStatusChecker struct {
	mock.Mock
}

type CCIPTransactionStatusChecker_Expecter struct {
	mock *mock.Mock
}

func (_m *CCIPTransactionStatusChecker) EXPECT() *CCIPTransactionStatusChecker_Expecter {
	return &CCIPTransactionStatusChecker_Expecter{mock: &_m.Mock}
}

// CheckMessageStatus provides a mock function with given fields: ctx, msgID
func (_m *CCIPTransactionStatusChecker) CheckMessageStatus(ctx context.Context, msgID string) ([]types.TransactionStatus, int, error) {
	ret := _m.Called(ctx, msgID)

	if len(ret) == 0 {
		panic("no return value specified for CheckMessageStatus")
	}

	var r0 []types.TransactionStatus
	var r1 int
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]types.TransactionStatus, int, error)); ok {
		return rf(ctx, msgID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []types.TransactionStatus); ok {
		r0 = rf(ctx, msgID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.TransactionStatus)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) int); ok {
		r1 = rf(ctx, msgID)
	} else {
		r1 = ret.Get(1).(int)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string) error); ok {
		r2 = rf(ctx, msgID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CCIPTransactionStatusChecker_CheckMessageStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckMessageStatus'
type CCIPTransactionStatusChecker_CheckMessageStatus_Call struct {
	*mock.Call
}

// CheckMessageStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - msgID string
func (_e *CCIPTransactionStatusChecker_Expecter) CheckMessageStatus(ctx interface{}, msgID interface{}) *CCIPTransactionStatusChecker_CheckMessageStatus_Call {
	return &CCIPTransactionStatusChecker_CheckMessageStatus_Call{Call: _e.mock.On("CheckMessageStatus", ctx, msgID)}
}

func (_c *CCIPTransactionStatusChecker_CheckMessageStatus_Call) Run(run func(ctx context.Context, msgID string)) *CCIPTransactionStatusChecker_CheckMessageStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CCIPTransactionStatusChecker_CheckMessageStatus_Call) Return(transactionStatuses []types.TransactionStatus, retryCounter int, err error) *CCIPTransactionStatusChecker_CheckMessageStatus_Call {
	_c.Call.Return(transactionStatuses, retryCounter, err)
	return _c
}

func (_c *CCIPTransactionStatusChecker_CheckMessageStatus_Call) RunAndReturn(run func(context.Context, string) ([]types.TransactionStatus, int, error)) *CCIPTransactionStatusChecker_CheckMessageStatus_Call {
	_c.Call.Return(run)
	return _c
}

// NewCCIPTransactionStatusChecker creates a new instance of CCIPTransactionStatusChecker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCCIPTransactionStatusChecker(t interface {
	mock.TestingT
	Cleanup(func())
}) *CCIPTransactionStatusChecker {
	mock := &CCIPTransactionStatusChecker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
