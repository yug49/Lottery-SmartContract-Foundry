// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// KVStore is an autogenerated mock type for the KVStore type
type KVStore struct {
	mock.Mock
}

type KVStore_Expecter struct {
	mock *mock.Mock
}

func (_m *KVStore) EXPECT() *KVStore_Expecter {
	return &KVStore_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: ctx, key
func (_m *KVStore) Get(ctx context.Context, key string) ([]byte, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]byte, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []byte); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// KVStore_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type KVStore_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *KVStore_Expecter) Get(ctx interface{}, key interface{}) *KVStore_Get_Call {
	return &KVStore_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *KVStore_Get_Call) Run(run func(ctx context.Context, key string)) *KVStore_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *KVStore_Get_Call) Return(_a0 []byte, _a1 error) *KVStore_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *KVStore_Get_Call) RunAndReturn(run func(context.Context, string) ([]byte, error)) *KVStore_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Store provides a mock function with given fields: ctx, key, val
func (_m *KVStore) Store(ctx context.Context, key string, val []byte) error {
	ret := _m.Called(ctx, key, val)

	if len(ret) == 0 {
		panic("no return value specified for Store")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []byte) error); ok {
		r0 = rf(ctx, key, val)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// KVStore_Store_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Store'
type KVStore_Store_Call struct {
	*mock.Call
}

// Store is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - val []byte
func (_e *KVStore_Expecter) Store(ctx interface{}, key interface{}, val interface{}) *KVStore_Store_Call {
	return &KVStore_Store_Call{Call: _e.mock.On("Store", ctx, key, val)}
}

func (_c *KVStore_Store_Call) Run(run func(ctx context.Context, key string, val []byte)) *KVStore_Store_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].([]byte))
	})
	return _c
}

func (_c *KVStore_Store_Call) Return(_a0 error) *KVStore_Store_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *KVStore_Store_Call) RunAndReturn(run func(context.Context, string, []byte) error) *KVStore_Store_Call {
	_c.Call.Return(run)
	return _c
}

// NewKVStore creates a new instance of KVStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewKVStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *KVStore {
	mock := &KVStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
