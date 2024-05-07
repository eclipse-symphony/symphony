package internal

import "github.com/stretchr/testify/mock"

type MockT struct {
	mock.Mock
}

func NewMockT() *MockT {
	return &MockT{}
}

// Errorf provides a mock function with given fields: format, args
func (_m *MockT) Errorf(format string, args ...interface{}) {
	_m.Called(format, args)
}

func (_m *MockT) Fatalf(format string, args ...interface{}) {
	_m.Called(format, args)
}

// Helper provides a mock function
func (_m *MockT) Helper() {
	_m.Called()
}
