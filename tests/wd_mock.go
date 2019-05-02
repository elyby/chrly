package tests

import (
	"time"

	"github.com/mono83/slf"
	"github.com/mono83/slf/wd"
	"github.com/stretchr/testify/mock"
)

type WdMock struct {
	mock.Mock
}

func (m *WdMock) Trace(message string, p ...slf.Param) {
	m.Called(message)
}

func (m *WdMock) Debug(message string, p ...slf.Param) {
	m.Called(message)
}

func (m *WdMock) Info(message string, p ...slf.Param) {
	m.Called(message)
}

func (m *WdMock) Warning(message string, p ...slf.Param) {
	m.Called(message)
}

func (m *WdMock) Error(message string, p ...slf.Param) {
	m.Called(message)
}

func (m *WdMock) Alert(message string, p ...slf.Param) {
	m.Called(message)
}

func (m *WdMock) Emergency(message string, p ...slf.Param) {
	m.Called(message)
}

func (m *WdMock) IncCounter(name string, value int64, p ...slf.Param) {
	m.Called(name, value)
}

func (m *WdMock) UpdateGauge(name string, value int64, p ...slf.Param) {
	m.Called(name, value)
}

func (m *WdMock) RecordTimer(name string, d time.Duration, p ...slf.Param) {
	m.Called(name, d)
}

func (m *WdMock) Timer(name string, p ...slf.Param) slf.Timer {
	return slf.NewTimer(name, p, m)
}

func (m *WdMock) WithParams(p ...slf.Param) wd.Watchdog {
	panic("this method shouldn't be used")
}
