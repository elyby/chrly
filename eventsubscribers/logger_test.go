package eventsubscribers

import (
	"net"
	"net/url"
	"syscall"
	"testing"

	"github.com/mono83/slf/params"
	"github.com/stretchr/testify/mock"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/dispatcher"
	"github.com/elyby/chrly/tests"
)

type LoggerTestCase struct {
	Events        [][]interface{}
	ExpectedCalls [][]interface{}
}

var loggerTestCases = map[string]*LoggerTestCase{}

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
			wdMock := &tests.WdMock{}
			if c.ExpectedCalls != nil {
				for _, c := range c.ExpectedCalls {
					topicName, _ := c[0].(string)
					wdMock.On(topicName, c[1:]...)
				}
			}

			reporter := &Logger{
				Logger: wdMock,
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
					wdMock.AssertCalled(t, topicName, c[1:]...)
				}
			}
		})
	}
}
