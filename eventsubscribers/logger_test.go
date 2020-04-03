package eventsubscribers

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"syscall"
	"testing"

	"github.com/mono83/slf"
	"github.com/mono83/slf/params"
	"github.com/stretchr/testify/mock"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/dispatcher"
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

type timeoutError struct{}

func (*timeoutError) Error() string   { return "timeout error" }
func (*timeoutError) Timeout() bool   { return true }
func (*timeoutError) Temporary() bool { return false }

func init() {
	// mojang_textures providers errors
	for _, providerName := range []string{"usernames", "textures"} {
		pn := providerName // Store pointer to iteration value
		loggerTestCases["should not log when no error occurred for "+pn+" provider"] = &LoggerTestCase{
			Events: [][]interface{}{
				{"mojang_textures:" + pn + ":after_call", pn, &mojang.ProfileInfo{}, nil},
			},
			ExpectedCalls: nil,
		}

		loggerTestCases["should not log when some network errors occured for "+pn+" provider"] = &LoggerTestCase{
			Events: [][]interface{}{
				{"mojang_textures:" + pn + ":after_call", pn, nil, &timeoutError{}},
				{"mojang_textures:" + pn + ":after_call", pn, nil, &url.Error{Op: "GET", URL: "http://localhost"}},
				{"mojang_textures:" + pn + ":after_call", pn, nil, &net.OpError{Op: "read"}},
				{"mojang_textures:" + pn + ":after_call", pn, nil, &net.OpError{Op: "dial"}},
				{"mojang_textures:" + pn + ":after_call", pn, nil, syscall.ECONNREFUSED},
			},
			ExpectedCalls: [][]interface{}{
				{"Debug", "Mojang " + pn + " provider resulted an error :err", mock.AnythingOfType("params.Error")},
			},
		}

		loggerTestCases["should log expected mojang errors for "+pn+" provider"] = &LoggerTestCase{
			Events: [][]interface{}{
				{"mojang_textures:" + pn + ":after_call", pn, nil, &mojang.BadRequestError{
					ErrorType: "IllegalArgumentException",
					Message:   "profileName can not be null or empty.",
				}},
				{"mojang_textures:" + pn + ":after_call", pn, nil, &mojang.ForbiddenError{}},
				{"mojang_textures:" + pn + ":after_call", pn, nil, &mojang.TooManyRequestsError{}},
			},
			ExpectedCalls: [][]interface{}{
				{"Debug", "Mojang " + pn + " provider resulted an error :err", mock.AnythingOfType("params.Error")},
				{"Warning",
					":name: :err",
					mock.MatchedBy(func(strParam params.String) bool {
						return strParam.Key == "name" && strParam.Value == pn
					}),
					mock.MatchedBy(func(errParam params.Error) bool {
						if errParam.Key != "err" {
							return false
						}

						if _, ok := errParam.Value.(*mojang.BadRequestError); ok {
							return true
						}

						if _, ok := errParam.Value.(*mojang.ForbiddenError); ok {
							return true
						}

						if _, ok := errParam.Value.(*mojang.TooManyRequestsError); ok {
							return true
						}

						return false
					}),
				},
			},
		}

		loggerTestCases["should call error when unexpected error occurred for "+pn+" provider"] = &LoggerTestCase{
			Events: [][]interface{}{
				{"mojang_textures:" + pn + ":after_call", pn, nil, &mojang.ServerError{Status: 500}},
			},
			ExpectedCalls: [][]interface{}{
				{"Debug", "Mojang " + pn + " provider resulted an error :err", mock.AnythingOfType("params.Error")},
				{"Error",
					":name: Unexpected Mojang response error: :err",
					mock.MatchedBy(func(strParam params.String) bool {
						return strParam.Key == "name" && strParam.Value == pn
					}),
					mock.MatchedBy(func(errParam params.Error) bool {
						if errParam.Key != "err" {
							return false
						}

						if _, ok := errParam.Value.(*mojang.ServerError); !ok {
							return false
						}

						return true
					}),
				},
			},
		}
	}
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
