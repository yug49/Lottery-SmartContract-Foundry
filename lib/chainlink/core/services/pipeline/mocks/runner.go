// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	pipeline "github.com/smartcontractkit/chainlink/v2/core/services/pipeline"
	mock "github.com/stretchr/testify/mock"

	sqlutil "github.com/smartcontractkit/chainlink-common/pkg/sqlutil"

	uuid "github.com/google/uuid"
)

// Runner is an autogenerated mock type for the Runner type
type Runner struct {
	mock.Mock
}

type Runner_Expecter struct {
	mock *mock.Mock
}

func (_m *Runner) EXPECT() *Runner_Expecter {
	return &Runner_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with no fields
func (_m *Runner) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Runner_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Runner_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *Runner_Expecter) Close() *Runner_Close_Call {
	return &Runner_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *Runner_Close_Call) Run(run func()) *Runner_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Runner_Close_Call) Return(_a0 error) *Runner_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Runner_Close_Call) RunAndReturn(run func() error) *Runner_Close_Call {
	_c.Call.Return(run)
	return _c
}

// ExecuteAndInsertFinishedRun provides a mock function with given fields: ctx, spec, vars, saveSuccessfulTaskRuns
func (_m *Runner) ExecuteAndInsertFinishedRun(ctx context.Context, spec pipeline.Spec, vars pipeline.Vars, saveSuccessfulTaskRuns bool) (int64, pipeline.TaskRunResults, error) {
	ret := _m.Called(ctx, spec, vars, saveSuccessfulTaskRuns)

	if len(ret) == 0 {
		panic("no return value specified for ExecuteAndInsertFinishedRun")
	}

	var r0 int64
	var r1 pipeline.TaskRunResults
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, pipeline.Spec, pipeline.Vars, bool) (int64, pipeline.TaskRunResults, error)); ok {
		return rf(ctx, spec, vars, saveSuccessfulTaskRuns)
	}
	if rf, ok := ret.Get(0).(func(context.Context, pipeline.Spec, pipeline.Vars, bool) int64); ok {
		r0 = rf(ctx, spec, vars, saveSuccessfulTaskRuns)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, pipeline.Spec, pipeline.Vars, bool) pipeline.TaskRunResults); ok {
		r1 = rf(ctx, spec, vars, saveSuccessfulTaskRuns)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(pipeline.TaskRunResults)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, pipeline.Spec, pipeline.Vars, bool) error); ok {
		r2 = rf(ctx, spec, vars, saveSuccessfulTaskRuns)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Runner_ExecuteAndInsertFinishedRun_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExecuteAndInsertFinishedRun'
type Runner_ExecuteAndInsertFinishedRun_Call struct {
	*mock.Call
}

// ExecuteAndInsertFinishedRun is a helper method to define mock.On call
//   - ctx context.Context
//   - spec pipeline.Spec
//   - vars pipeline.Vars
//   - saveSuccessfulTaskRuns bool
func (_e *Runner_Expecter) ExecuteAndInsertFinishedRun(ctx interface{}, spec interface{}, vars interface{}, saveSuccessfulTaskRuns interface{}) *Runner_ExecuteAndInsertFinishedRun_Call {
	return &Runner_ExecuteAndInsertFinishedRun_Call{Call: _e.mock.On("ExecuteAndInsertFinishedRun", ctx, spec, vars, saveSuccessfulTaskRuns)}
}

func (_c *Runner_ExecuteAndInsertFinishedRun_Call) Run(run func(ctx context.Context, spec pipeline.Spec, vars pipeline.Vars, saveSuccessfulTaskRuns bool)) *Runner_ExecuteAndInsertFinishedRun_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(pipeline.Spec), args[2].(pipeline.Vars), args[3].(bool))
	})
	return _c
}

func (_c *Runner_ExecuteAndInsertFinishedRun_Call) Return(runID int64, results pipeline.TaskRunResults, err error) *Runner_ExecuteAndInsertFinishedRun_Call {
	_c.Call.Return(runID, results, err)
	return _c
}

func (_c *Runner_ExecuteAndInsertFinishedRun_Call) RunAndReturn(run func(context.Context, pipeline.Spec, pipeline.Vars, bool) (int64, pipeline.TaskRunResults, error)) *Runner_ExecuteAndInsertFinishedRun_Call {
	_c.Call.Return(run)
	return _c
}

// ExecuteRun provides a mock function with given fields: ctx, spec, vars
func (_m *Runner) ExecuteRun(ctx context.Context, spec pipeline.Spec, vars pipeline.Vars) (*pipeline.Run, pipeline.TaskRunResults, error) {
	ret := _m.Called(ctx, spec, vars)

	if len(ret) == 0 {
		panic("no return value specified for ExecuteRun")
	}

	var r0 *pipeline.Run
	var r1 pipeline.TaskRunResults
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, pipeline.Spec, pipeline.Vars) (*pipeline.Run, pipeline.TaskRunResults, error)); ok {
		return rf(ctx, spec, vars)
	}
	if rf, ok := ret.Get(0).(func(context.Context, pipeline.Spec, pipeline.Vars) *pipeline.Run); ok {
		r0 = rf(ctx, spec, vars)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pipeline.Run)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, pipeline.Spec, pipeline.Vars) pipeline.TaskRunResults); ok {
		r1 = rf(ctx, spec, vars)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(pipeline.TaskRunResults)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, pipeline.Spec, pipeline.Vars) error); ok {
		r2 = rf(ctx, spec, vars)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Runner_ExecuteRun_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExecuteRun'
type Runner_ExecuteRun_Call struct {
	*mock.Call
}

// ExecuteRun is a helper method to define mock.On call
//   - ctx context.Context
//   - spec pipeline.Spec
//   - vars pipeline.Vars
func (_e *Runner_Expecter) ExecuteRun(ctx interface{}, spec interface{}, vars interface{}) *Runner_ExecuteRun_Call {
	return &Runner_ExecuteRun_Call{Call: _e.mock.On("ExecuteRun", ctx, spec, vars)}
}

func (_c *Runner_ExecuteRun_Call) Run(run func(ctx context.Context, spec pipeline.Spec, vars pipeline.Vars)) *Runner_ExecuteRun_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(pipeline.Spec), args[2].(pipeline.Vars))
	})
	return _c
}

func (_c *Runner_ExecuteRun_Call) Return(run *pipeline.Run, trrs pipeline.TaskRunResults, err error) *Runner_ExecuteRun_Call {
	_c.Call.Return(run, trrs, err)
	return _c
}

func (_c *Runner_ExecuteRun_Call) RunAndReturn(run func(context.Context, pipeline.Spec, pipeline.Vars) (*pipeline.Run, pipeline.TaskRunResults, error)) *Runner_ExecuteRun_Call {
	_c.Call.Return(run)
	return _c
}

// HealthReport provides a mock function with no fields
func (_m *Runner) HealthReport() map[string]error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for HealthReport")
	}

	var r0 map[string]error
	if rf, ok := ret.Get(0).(func() map[string]error); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]error)
		}
	}

	return r0
}

// Runner_HealthReport_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HealthReport'
type Runner_HealthReport_Call struct {
	*mock.Call
}

// HealthReport is a helper method to define mock.On call
func (_e *Runner_Expecter) HealthReport() *Runner_HealthReport_Call {
	return &Runner_HealthReport_Call{Call: _e.mock.On("HealthReport")}
}

func (_c *Runner_HealthReport_Call) Run(run func()) *Runner_HealthReport_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Runner_HealthReport_Call) Return(_a0 map[string]error) *Runner_HealthReport_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Runner_HealthReport_Call) RunAndReturn(run func() map[string]error) *Runner_HealthReport_Call {
	_c.Call.Return(run)
	return _c
}

// InitializePipeline provides a mock function with given fields: spec
func (_m *Runner) InitializePipeline(spec pipeline.Spec) (*pipeline.Pipeline, error) {
	ret := _m.Called(spec)

	if len(ret) == 0 {
		panic("no return value specified for InitializePipeline")
	}

	var r0 *pipeline.Pipeline
	var r1 error
	if rf, ok := ret.Get(0).(func(pipeline.Spec) (*pipeline.Pipeline, error)); ok {
		return rf(spec)
	}
	if rf, ok := ret.Get(0).(func(pipeline.Spec) *pipeline.Pipeline); ok {
		r0 = rf(spec)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pipeline.Pipeline)
		}
	}

	if rf, ok := ret.Get(1).(func(pipeline.Spec) error); ok {
		r1 = rf(spec)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Runner_InitializePipeline_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InitializePipeline'
type Runner_InitializePipeline_Call struct {
	*mock.Call
}

// InitializePipeline is a helper method to define mock.On call
//   - spec pipeline.Spec
func (_e *Runner_Expecter) InitializePipeline(spec interface{}) *Runner_InitializePipeline_Call {
	return &Runner_InitializePipeline_Call{Call: _e.mock.On("InitializePipeline", spec)}
}

func (_c *Runner_InitializePipeline_Call) Run(run func(spec pipeline.Spec)) *Runner_InitializePipeline_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(pipeline.Spec))
	})
	return _c
}

func (_c *Runner_InitializePipeline_Call) Return(_a0 *pipeline.Pipeline, _a1 error) *Runner_InitializePipeline_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Runner_InitializePipeline_Call) RunAndReturn(run func(pipeline.Spec) (*pipeline.Pipeline, error)) *Runner_InitializePipeline_Call {
	_c.Call.Return(run)
	return _c
}

// InsertFinishedRun provides a mock function with given fields: ctx, ds, run, saveSuccessfulTaskRuns
func (_m *Runner) InsertFinishedRun(ctx context.Context, ds sqlutil.DataSource, run *pipeline.Run, saveSuccessfulTaskRuns bool) error {
	ret := _m.Called(ctx, ds, run, saveSuccessfulTaskRuns)

	if len(ret) == 0 {
		panic("no return value specified for InsertFinishedRun")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlutil.DataSource, *pipeline.Run, bool) error); ok {
		r0 = rf(ctx, ds, run, saveSuccessfulTaskRuns)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Runner_InsertFinishedRun_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InsertFinishedRun'
type Runner_InsertFinishedRun_Call struct {
	*mock.Call
}

// InsertFinishedRun is a helper method to define mock.On call
//   - ctx context.Context
//   - ds sqlutil.DataSource
//   - run *pipeline.Run
//   - saveSuccessfulTaskRuns bool
func (_e *Runner_Expecter) InsertFinishedRun(ctx interface{}, ds interface{}, run interface{}, saveSuccessfulTaskRuns interface{}) *Runner_InsertFinishedRun_Call {
	return &Runner_InsertFinishedRun_Call{Call: _e.mock.On("InsertFinishedRun", ctx, ds, run, saveSuccessfulTaskRuns)}
}

func (_c *Runner_InsertFinishedRun_Call) Run(run func(ctx context.Context, ds sqlutil.DataSource, run *pipeline.Run, saveSuccessfulTaskRuns bool)) *Runner_InsertFinishedRun_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlutil.DataSource), args[2].(*pipeline.Run), args[3].(bool))
	})
	return _c
}

func (_c *Runner_InsertFinishedRun_Call) Return(_a0 error) *Runner_InsertFinishedRun_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Runner_InsertFinishedRun_Call) RunAndReturn(run func(context.Context, sqlutil.DataSource, *pipeline.Run, bool) error) *Runner_InsertFinishedRun_Call {
	_c.Call.Return(run)
	return _c
}

// InsertFinishedRuns provides a mock function with given fields: ctx, ds, runs, saveSuccessfulTaskRuns
func (_m *Runner) InsertFinishedRuns(ctx context.Context, ds sqlutil.DataSource, runs []*pipeline.Run, saveSuccessfulTaskRuns bool) error {
	ret := _m.Called(ctx, ds, runs, saveSuccessfulTaskRuns)

	if len(ret) == 0 {
		panic("no return value specified for InsertFinishedRuns")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlutil.DataSource, []*pipeline.Run, bool) error); ok {
		r0 = rf(ctx, ds, runs, saveSuccessfulTaskRuns)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Runner_InsertFinishedRuns_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InsertFinishedRuns'
type Runner_InsertFinishedRuns_Call struct {
	*mock.Call
}

// InsertFinishedRuns is a helper method to define mock.On call
//   - ctx context.Context
//   - ds sqlutil.DataSource
//   - runs []*pipeline.Run
//   - saveSuccessfulTaskRuns bool
func (_e *Runner_Expecter) InsertFinishedRuns(ctx interface{}, ds interface{}, runs interface{}, saveSuccessfulTaskRuns interface{}) *Runner_InsertFinishedRuns_Call {
	return &Runner_InsertFinishedRuns_Call{Call: _e.mock.On("InsertFinishedRuns", ctx, ds, runs, saveSuccessfulTaskRuns)}
}

func (_c *Runner_InsertFinishedRuns_Call) Run(run func(ctx context.Context, ds sqlutil.DataSource, runs []*pipeline.Run, saveSuccessfulTaskRuns bool)) *Runner_InsertFinishedRuns_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlutil.DataSource), args[2].([]*pipeline.Run), args[3].(bool))
	})
	return _c
}

func (_c *Runner_InsertFinishedRuns_Call) Return(_a0 error) *Runner_InsertFinishedRuns_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Runner_InsertFinishedRuns_Call) RunAndReturn(run func(context.Context, sqlutil.DataSource, []*pipeline.Run, bool) error) *Runner_InsertFinishedRuns_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with no fields
func (_m *Runner) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Runner_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type Runner_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *Runner_Expecter) Name() *Runner_Name_Call {
	return &Runner_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *Runner_Name_Call) Run(run func()) *Runner_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Runner_Name_Call) Return(_a0 string) *Runner_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Runner_Name_Call) RunAndReturn(run func() string) *Runner_Name_Call {
	_c.Call.Return(run)
	return _c
}

// OnRunFinished provides a mock function with given fields: _a0
func (_m *Runner) OnRunFinished(_a0 func(*pipeline.Run)) {
	_m.Called(_a0)
}

// Runner_OnRunFinished_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OnRunFinished'
type Runner_OnRunFinished_Call struct {
	*mock.Call
}

// OnRunFinished is a helper method to define mock.On call
//   - _a0 func(*pipeline.Run)
func (_e *Runner_Expecter) OnRunFinished(_a0 interface{}) *Runner_OnRunFinished_Call {
	return &Runner_OnRunFinished_Call{Call: _e.mock.On("OnRunFinished", _a0)}
}

func (_c *Runner_OnRunFinished_Call) Run(run func(_a0 func(*pipeline.Run))) *Runner_OnRunFinished_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(func(*pipeline.Run)))
	})
	return _c
}

func (_c *Runner_OnRunFinished_Call) Return() *Runner_OnRunFinished_Call {
	_c.Call.Return()
	return _c
}

func (_c *Runner_OnRunFinished_Call) RunAndReturn(run func(func(*pipeline.Run))) *Runner_OnRunFinished_Call {
	_c.Run(run)
	return _c
}

// Ready provides a mock function with no fields
func (_m *Runner) Ready() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Ready")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Runner_Ready_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Ready'
type Runner_Ready_Call struct {
	*mock.Call
}

// Ready is a helper method to define mock.On call
func (_e *Runner_Expecter) Ready() *Runner_Ready_Call {
	return &Runner_Ready_Call{Call: _e.mock.On("Ready")}
}

func (_c *Runner_Ready_Call) Run(run func()) *Runner_Ready_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Runner_Ready_Call) Return(_a0 error) *Runner_Ready_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Runner_Ready_Call) RunAndReturn(run func() error) *Runner_Ready_Call {
	_c.Call.Return(run)
	return _c
}

// ResumeRun provides a mock function with given fields: ctx, taskID, value, err
func (_m *Runner) ResumeRun(ctx context.Context, taskID uuid.UUID, value interface{}, err error) error {
	ret := _m.Called(ctx, taskID, value, err)

	if len(ret) == 0 {
		panic("no return value specified for ResumeRun")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, interface{}, error) error); ok {
		r0 = rf(ctx, taskID, value, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Runner_ResumeRun_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResumeRun'
type Runner_ResumeRun_Call struct {
	*mock.Call
}

// ResumeRun is a helper method to define mock.On call
//   - ctx context.Context
//   - taskID uuid.UUID
//   - value interface{}
//   - err error
func (_e *Runner_Expecter) ResumeRun(ctx interface{}, taskID interface{}, value interface{}, err interface{}) *Runner_ResumeRun_Call {
	return &Runner_ResumeRun_Call{Call: _e.mock.On("ResumeRun", ctx, taskID, value, err)}
}

func (_c *Runner_ResumeRun_Call) Run(run func(ctx context.Context, taskID uuid.UUID, value interface{}, err error)) *Runner_ResumeRun_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uuid.UUID), args[2].(interface{}), args[3].(error))
	})
	return _c
}

func (_c *Runner_ResumeRun_Call) Return(_a0 error) *Runner_ResumeRun_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Runner_ResumeRun_Call) RunAndReturn(run func(context.Context, uuid.UUID, interface{}, error) error) *Runner_ResumeRun_Call {
	_c.Call.Return(run)
	return _c
}

// Run provides a mock function with given fields: ctx, run, saveSuccessfulTaskRuns, fn
func (_m *Runner) Run(ctx context.Context, run *pipeline.Run, saveSuccessfulTaskRuns bool, fn func(sqlutil.DataSource) error) (bool, error) {
	ret := _m.Called(ctx, run, saveSuccessfulTaskRuns, fn)

	if len(ret) == 0 {
		panic("no return value specified for Run")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *pipeline.Run, bool, func(sqlutil.DataSource) error) (bool, error)); ok {
		return rf(ctx, run, saveSuccessfulTaskRuns, fn)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *pipeline.Run, bool, func(sqlutil.DataSource) error) bool); ok {
		r0 = rf(ctx, run, saveSuccessfulTaskRuns, fn)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *pipeline.Run, bool, func(sqlutil.DataSource) error) error); ok {
		r1 = rf(ctx, run, saveSuccessfulTaskRuns, fn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Runner_Run_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Run'
type Runner_Run_Call struct {
	*mock.Call
}

// Run is a helper method to define mock.On call
//   - ctx context.Context
//   - run *pipeline.Run
//   - saveSuccessfulTaskRuns bool
//   - fn func(sqlutil.DataSource) error
func (_e *Runner_Expecter) Run(ctx interface{}, run interface{}, saveSuccessfulTaskRuns interface{}, fn interface{}) *Runner_Run_Call {
	return &Runner_Run_Call{Call: _e.mock.On("Run", ctx, run, saveSuccessfulTaskRuns, fn)}
}

func (_c *Runner_Run_Call) Run(run func(ctx context.Context, run *pipeline.Run, saveSuccessfulTaskRuns bool, fn func(sqlutil.DataSource) error)) *Runner_Run_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*pipeline.Run), args[2].(bool), args[3].(func(sqlutil.DataSource) error))
	})
	return _c
}

func (_c *Runner_Run_Call) Return(incomplete bool, err error) *Runner_Run_Call {
	_c.Call.Return(incomplete, err)
	return _c
}

func (_c *Runner_Run_Call) RunAndReturn(run func(context.Context, *pipeline.Run, bool, func(sqlutil.DataSource) error) (bool, error)) *Runner_Run_Call {
	_c.Call.Return(run)
	return _c
}

// Start provides a mock function with given fields: _a0
func (_m *Runner) Start(_a0 context.Context) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Start")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Runner_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type Runner_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *Runner_Expecter) Start(_a0 interface{}) *Runner_Start_Call {
	return &Runner_Start_Call{Call: _e.mock.On("Start", _a0)}
}

func (_c *Runner_Start_Call) Run(run func(_a0 context.Context)) *Runner_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Runner_Start_Call) Return(_a0 error) *Runner_Start_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Runner_Start_Call) RunAndReturn(run func(context.Context) error) *Runner_Start_Call {
	_c.Call.Return(run)
	return _c
}

// NewRunner creates a new instance of Runner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRunner(t interface {
	mock.TestingT
	Cleanup(func())
}) *Runner {
	mock := &Runner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
