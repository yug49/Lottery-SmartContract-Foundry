// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	common "github.com/ethereum/go-ethereum/common"

	mock "github.com/stretchr/testify/mock"

	s4 "github.com/smartcontractkit/chainlink/v2/core/services/s4"
)

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

type Storage_Expecter struct {
	mock *mock.Mock
}

func (_m *Storage) EXPECT() *Storage_Expecter {
	return &Storage_Expecter{mock: &_m.Mock}
}

// Constraints provides a mock function with no fields
func (_m *Storage) Constraints() s4.Constraints {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Constraints")
	}

	var r0 s4.Constraints
	if rf, ok := ret.Get(0).(func() s4.Constraints); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(s4.Constraints)
	}

	return r0
}

// Storage_Constraints_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Constraints'
type Storage_Constraints_Call struct {
	*mock.Call
}

// Constraints is a helper method to define mock.On call
func (_e *Storage_Expecter) Constraints() *Storage_Constraints_Call {
	return &Storage_Constraints_Call{Call: _e.mock.On("Constraints")}
}

func (_c *Storage_Constraints_Call) Run(run func()) *Storage_Constraints_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Storage_Constraints_Call) Return(_a0 s4.Constraints) *Storage_Constraints_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_Constraints_Call) RunAndReturn(run func() s4.Constraints) *Storage_Constraints_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, key
func (_m *Storage) Get(ctx context.Context, key *s4.Key) (*s4.Record, *s4.Metadata, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *s4.Record
	var r1 *s4.Metadata
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, *s4.Key) (*s4.Record, *s4.Metadata, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *s4.Key) *s4.Record); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*s4.Record)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *s4.Key) *s4.Metadata); ok {
		r1 = rf(ctx, key)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*s4.Metadata)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, *s4.Key) error); ok {
		r2 = rf(ctx, key)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Storage_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Storage_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - key *s4.Key
func (_e *Storage_Expecter) Get(ctx interface{}, key interface{}) *Storage_Get_Call {
	return &Storage_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *Storage_Get_Call) Run(run func(ctx context.Context, key *s4.Key)) *Storage_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*s4.Key))
	})
	return _c
}

func (_c *Storage_Get_Call) Return(_a0 *s4.Record, _a1 *s4.Metadata, _a2 error) *Storage_Get_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *Storage_Get_Call) RunAndReturn(run func(context.Context, *s4.Key) (*s4.Record, *s4.Metadata, error)) *Storage_Get_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx, address
func (_m *Storage) List(ctx context.Context, address common.Address) ([]*s4.SnapshotRow, error) {
	ret := _m.Called(ctx, address)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []*s4.SnapshotRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) ([]*s4.SnapshotRow, error)); ok {
		return rf(ctx, address)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) []*s4.SnapshotRow); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*s4.SnapshotRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address) error); ok {
		r1 = rf(ctx, address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type Storage_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - address common.Address
func (_e *Storage_Expecter) List(ctx interface{}, address interface{}) *Storage_List_Call {
	return &Storage_List_Call{Call: _e.mock.On("List", ctx, address)}
}

func (_c *Storage_List_Call) Run(run func(ctx context.Context, address common.Address)) *Storage_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address))
	})
	return _c
}

func (_c *Storage_List_Call) Return(_a0 []*s4.SnapshotRow, _a1 error) *Storage_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_List_Call) RunAndReturn(run func(context.Context, common.Address) ([]*s4.SnapshotRow, error)) *Storage_List_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: ctx, key, record, signature
func (_m *Storage) Put(ctx context.Context, key *s4.Key, record *s4.Record, signature []byte) error {
	ret := _m.Called(ctx, key, record, signature)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *s4.Key, *s4.Record, []byte) error); ok {
		r0 = rf(ctx, key, record, signature)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type Storage_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - ctx context.Context
//   - key *s4.Key
//   - record *s4.Record
//   - signature []byte
func (_e *Storage_Expecter) Put(ctx interface{}, key interface{}, record interface{}, signature interface{}) *Storage_Put_Call {
	return &Storage_Put_Call{Call: _e.mock.On("Put", ctx, key, record, signature)}
}

func (_c *Storage_Put_Call) Run(run func(ctx context.Context, key *s4.Key, record *s4.Record, signature []byte)) *Storage_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*s4.Key), args[2].(*s4.Record), args[3].([]byte))
	})
	return _c
}

func (_c *Storage_Put_Call) Return(_a0 error) *Storage_Put_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_Put_Call) RunAndReturn(run func(context.Context, *s4.Key, *s4.Record, []byte) error) *Storage_Put_Call {
	_c.Call.Return(run)
	return _c
}

// NewStorage creates a new instance of Storage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *Storage {
	mock := &Storage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
