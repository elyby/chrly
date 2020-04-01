package tests

import (
	"time"

	"github.com/mono83/slf"
	"github.com/mono83/slf/wd"
	"github.com/stretchr/testify/mock"
)

func paramsSliceToInterfaceSlice(params []slf.Param) []interface{} {
	result := make([]interface{}, len(params))
	for i, v := range params {
		result[i], _ = v.(interface{})
	}

	return result
}

func prepareLoggerArgs(message string, params []slf.Param) []interface{} {
	return append([]interface{}{message}, paramsSliceToInterfaceSlice(params)...)
}

type WdMock struct {
	mock.Mock
}

func (m *WdMock) Trace(message string, params ...slf.Param) {
	m.Called(prepareLoggerArgs(message, params)...)
}

func (m *WdMock) Debug(message string, params ...slf.Param) {
	m.Called(prepareLoggerArgs(message, params)...)
}

func (m *WdMock) Info(message string, params ...slf.Param) {
	m.Called(prepareLoggerArgs(message, params)...)
}

func (m *WdMock) Warning(message string, params ...slf.Param) {
	m.Called(prepareLoggerArgs(message, params)...)
}

func (m *WdMock) Error(message string, params ...slf.Param) {
	m.Called(prepareLoggerArgs(message, params)...)
}

func (m *WdMock) Alert(message string, params ...slf.Param) {
	m.Called(prepareLoggerArgs(message, params)...)
}

func (m *WdMock) Emergency(message string, params ...slf.Param) {
	m.Called(prepareLoggerArgs(message, params)...)
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
