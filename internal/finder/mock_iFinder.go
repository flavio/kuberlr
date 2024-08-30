// Code generated by mockery v2.43.2. DO NOT EDIT.

package finder

import mock "github.com/stretchr/testify/mock"

// MockiFinder is an autogenerated mock type for the iFinder type
type MockiFinder struct {
	mock.Mock
}

type MockiFinder_Expecter struct {
	mock *mock.Mock
}

func (_m *MockiFinder) EXPECT() *MockiFinder_Expecter {
	return &MockiFinder_Expecter{mock: &_m.Mock}
}

// AllKubectlBinaries provides a mock function with given fields: reverseSort
func (_m *MockiFinder) AllKubectlBinaries(reverseSort bool) KubectlBinaries {
	ret := _m.Called(reverseSort)

	if len(ret) == 0 {
		panic("no return value specified for AllKubectlBinaries")
	}

	var r0 KubectlBinaries
	if rf, ok := ret.Get(0).(func(bool) KubectlBinaries); ok {
		r0 = rf(reverseSort)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(KubectlBinaries)
		}
	}

	return r0
}

// MockiFinder_AllKubectlBinaries_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AllKubectlBinaries'
type MockiFinder_AllKubectlBinaries_Call struct {
	*mock.Call
}

// AllKubectlBinaries is a helper method to define mock.On call
//   - reverseSort bool
func (_e *MockiFinder_Expecter) AllKubectlBinaries(reverseSort interface{}) *MockiFinder_AllKubectlBinaries_Call {
	return &MockiFinder_AllKubectlBinaries_Call{Call: _e.mock.On("AllKubectlBinaries", reverseSort)}
}

func (_c *MockiFinder_AllKubectlBinaries_Call) Run(run func(reverseSort bool)) *MockiFinder_AllKubectlBinaries_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(bool))
	})
	return _c
}

func (_c *MockiFinder_AllKubectlBinaries_Call) Return(_a0 KubectlBinaries) *MockiFinder_AllKubectlBinaries_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockiFinder_AllKubectlBinaries_Call) RunAndReturn(run func(bool) KubectlBinaries) *MockiFinder_AllKubectlBinaries_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockiFinder creates a new instance of MockiFinder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockiFinder(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockiFinder {
	mock := &MockiFinder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}