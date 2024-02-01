package eventsubscribers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mono83/slf"
	"github.com/mono83/slf/params"
	"github.com/stretchr/testify/mock"

	"ely.by/chrly/internal/dispatcher"
)

type LoggerMock struct {
	mock.Mock
}

func prepareLoggerArgs(message string, params []slf.Param) []interface{} {
	args := []interface{}{message}
	for _, v := range params {
		args = append(args, v.(interface{}))
	}

	return args
}

func (l *LoggerMock) Trace(message string, params ...slf.Param) {
	l.Called(prepareLoggerArgs(message, params)...)
}

func (l *LoggerMock) Debug(message string, params ...slf.Param) {
	l.Called(prepareLoggerArgs(message, params)...)
}

func (l *LoggerMock) Info(message string, params ...slf.Param) {
	l.Called(prepareLoggerArgs(message, params)...)
}

func (l *LoggerMock) Warning(message string, params ...slf.Param) {
	l.Called(prepareLoggerArgs(message, params)...)
}

func (l *LoggerMock) Error(message string, params ...slf.Param) {
	l.Called(prepareLoggerArgs(message, params)...)
}

func (l *LoggerMock) Alert(message string, params ...slf.Param) {
	l.Called(prepareLoggerArgs(message, params)...)
}

func (l *LoggerMock) Emergency(message string, params ...slf.Param) {
	l.Called(prepareLoggerArgs(message, params)...)
}

type LoggerTestCase struct {
	Events        [][]interface{}
	ExpectedCalls [][]interface{}
}

var loggerTestCases = map[string]*LoggerTestCase{
	"should log each request to the skinsystem": {
		Events: [][]interface{}{
			{"skinsystem:after_request",
				(func() *http.Request {
					req := httptest.NewRequest("GET", "http://localhost/skins/username.png", nil)
					req.Header.Add("User-Agent", "Test user agent")

					return req
				})(),
				201,
			},
		},
		ExpectedCalls: [][]interface{}{
			{"Info",
				":ip - - \":method :path\" :statusCode - \":userAgent\" \":forwardedIp\"",
				mock.MatchedBy(func(strParam params.String) bool {
					return strParam.Key == "ip" && strParam.Value == "192.0.2.1"
				}),
				mock.MatchedBy(func(strParam params.String) bool {
					return strParam.Key == "method" && strParam.Value == "GET"
				}),
				mock.MatchedBy(func(strParam params.String) bool {
					return strParam.Key == "path" && strParam.Value == "/skins/username.png"
				}),
				mock.MatchedBy(func(strParam params.Int) bool {
					return strParam.Key == "statusCode" && strParam.Value == 201
				}),
				mock.MatchedBy(func(strParam params.String) bool {
					return strParam.Key == "userAgent" && strParam.Value == "Test user agent"
				}),
				mock.MatchedBy(func(strParam params.String) bool {
					return strParam.Key == "forwardedIp" && strParam.Value == ""
				}),
			},
		},
	},
	"should log each request to the skinsystem 2": {
		Events: [][]interface{}{
			{"skinsystem:after_request",
				(func() *http.Request {
					req := httptest.NewRequest("GET", "http://localhost/skins/username.png?authlib=1.5.2", nil)
					req.Header.Add("User-Agent", "Test user agent")
					req.Header.Add("X-Forwarded-For", "1.2.3.4")

					return req
				})(),
				201,
			},
		},
		ExpectedCalls: [][]interface{}{
			{"Info",
				":ip - - \":method :path\" :statusCode - \":userAgent\" \":forwardedIp\"",
				mock.Anything, // Already tested
				mock.Anything, // Already tested
				mock.MatchedBy(func(strParam params.String) bool {
					return strParam.Key == "path" && strParam.Value == "/skins/username.png?authlib=1.5.2"
				}),
				mock.Anything, // Already tested
				mock.Anything, // Already tested
				mock.MatchedBy(func(strParam params.String) bool {
					return strParam.Key == "forwardedIp" && strParam.Value == "1.2.3.4"
				}),
			},
		},
	},
}

func TestLogger(t *testing.T) {
	for name, c := range loggerTestCases {
		t.Run(name, func(t *testing.T) {
			loggerMock := &LoggerMock{}
			if c.ExpectedCalls != nil {
				for _, c := range c.ExpectedCalls {
					topicName, _ := c[0].(string)
					loggerMock.On(topicName, c[1:]...)
				}
			}

			reporter := &Logger{
				Logger: loggerMock,
			}

			d := dispatcher.New()
			reporter.ConfigureWithDispatcher(d)
			for _, args := range c.Events {
				eventName, _ := args[0].(string)
				d.Emit(eventName, args[1:]...)
			}

			if c.ExpectedCalls != nil {
				for _, c := range c.ExpectedCalls {
					topicName, _ := c[0].(string)
					loggerMock.AssertCalled(t, topicName, c[1:]...)
				}
			}
		})
	}
}
