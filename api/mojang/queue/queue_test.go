package queue

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/elyby/chrly/api/mojang"
	testify "github.com/stretchr/testify/assert"
)

func TestJobsQueue_GetTexturesForUsername(t *testing.T) {
	delay = 50 * time.Millisecond

	t.Run("receive textures for one username", func(t *testing.T) {
		assert := testify.New(t)

		usernamesToUuids = createUsernameToUuidsMock(
			assert,
			[]string{"maksimkurb"},
			[]*mojang.ProfileInfo{
				{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
			},
			nil,
		)
		uuidToTextures = createUuidToTextures([]*createUuidToTexturesResult{
			createTexturesResult("0d252b7218b648bfb86c2ae476954d32", "maksimkurb"),
		})

		queue := &JobsQueue{Storage: &NilStorage{}}
		result := queue.GetTexturesForUsername("maksimkurb")

		if assert.NotNil(result) {
			assert.Equal("0d252b7218b648bfb86c2ae476954d32", result.Id)
			assert.Equal("maksimkurb", result.Name)
		}
	})

	t.Run("receive textures for few usernames", func(t *testing.T) {
		assert := testify.New(t)

		usernamesToUuids = createUsernameToUuidsMock(
			assert,
			[]string{"maksimkurb", "Thinkofdeath"},
			[]*mojang.ProfileInfo{
				{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
				{Id: "4566e69fc90748ee8d71d7ba5aa00d20", Name: "Thinkofdeath"},
			},
			nil,
		)
		uuidToTextures = createUuidToTextures([]*createUuidToTexturesResult{
			createTexturesResult("0d252b7218b648bfb86c2ae476954d32", "maksimkurb"),
			createTexturesResult("4566e69fc90748ee8d71d7ba5aa00d20", "Thinkofdeath"),
		})

		queue := &JobsQueue{Storage: &NilStorage{}}
		resultChan1 := make(chan *mojang.SignedTexturesResponse)
		resultChan2 := make(chan *mojang.SignedTexturesResponse)
		go func() {
			resultChan1 <- queue.GetTexturesForUsername("maksimkurb")
		}()
		go func() {
			resultChan2 <- queue.GetTexturesForUsername("Thinkofdeath")
		}()

		assert.NotNil(<-resultChan1)
		assert.NotNil(<-resultChan2)
	})

	t.Run("query no more than 100 usernames and all left on the next iteration", func(t *testing.T) {
		assert := testify.New(t)

		usernames := make([]string, 120, 120)
		for i := 0; i < 120; i++ {
			usernames[i] = randStr(8)
		}

		usernamesToUuids = createUsernameToUuidsMock(assert, usernames[0:100], []*mojang.ProfileInfo{}, nil)

		queue := &JobsQueue{Storage: &NilStorage{}}

		scheduleUsername := func(username string) {
			queue.GetTexturesForUsername(username)
		}

		for _, username := range usernames {
			go scheduleUsername(username)
			time.Sleep(50 * time.Microsecond) // Add delay to have consistent order
		}

		// Let it begin first iteration
		time.Sleep(delay + delay/2)

		usernamesToUuids = createUsernameToUuidsMock(
			assert,
			usernames[100:120],
			[]*mojang.ProfileInfo{},
			nil,
		)

		time.Sleep(delay)
	})

	t.Run("should do nothing if queue is empty", func(t *testing.T) {
		assert := testify.New(t)

		usernamesToUuids = createUsernameToUuidsMock(assert, []string{"maksimkurb"}, []*mojang.ProfileInfo{}, nil)
		uuidToTextures = func(uuid string, signed bool) (*mojang.SignedTexturesResponse, error) {
			t.Error("this method shouldn't be called")
			return nil, nil
		}

		// Perform first iteration and await it finish
		queue := &JobsQueue{Storage: &NilStorage{}}
		result := queue.GetTexturesForUsername("maksimkurb")
		assert.Nil(result)

		// Override external API call that indicates, that queue is still trying to obtain somethid
		usernamesToUuids = func(usernames []string) ([]*mojang.ProfileInfo, error) {
			t.Error("this method shouldn't be called")
			return nil, nil
		}

		// Let it to iterate few times
		time.Sleep(delay * 2)
	})

	t.Run("handle 429 error when exchanging usernames to uuids", func(t *testing.T) {
		assert := testify.New(t)

		usernamesToUuids = createUsernameToUuidsMock(assert, []string{"maksimkurb"}, nil, &mojang.TooManyRequestsError{})

		queue := &JobsQueue{Storage: &NilStorage{}}
		result := queue.GetTexturesForUsername("maksimkurb")
		assert.Nil(result)
	})

	t.Run("handle 429 error when requesting user's textures", func(t *testing.T) {
		assert := testify.New(t)

		usernamesToUuids = createUsernameToUuidsMock(
			assert,
			[]string{"maksimkurb"},
			[]*mojang.ProfileInfo{
				{Id: "0d252b7218b648bfb86c2ae476954d32", Name: "maksimkurb"},
			},
			nil,
		)
		uuidToTextures = createUuidToTextures([]*createUuidToTexturesResult{
			createTexturesResult("0d252b7218b648bfb86c2ae476954d32", &mojang.TooManyRequestsError{}),
		})

		queue := &JobsQueue{Storage: &NilStorage{}}
		result := queue.GetTexturesForUsername("maksimkurb")
		assert.Nil(result)
	})
}

func createUsernameToUuidsMock(
	assert *testify.Assertions,
	expectedUsernames []string,
	result []*mojang.ProfileInfo,
	err error,
) func(usernames []string) ([]*mojang.ProfileInfo, error) {
	return func(usernames []string) ([]*mojang.ProfileInfo, error) {
		assert.ElementsMatch(expectedUsernames, usernames)
		return result, err
	}
}

type createUuidToTexturesResult struct {
	uuid   string
	result *mojang.SignedTexturesResponse
	err    error
}

func createTexturesResult(uuid string, result interface{}) *createUuidToTexturesResult {
	output := &createUuidToTexturesResult{uuid: uuid}
	if username, ok := result.(string); ok {
		output.result = &mojang.SignedTexturesResponse{Id: uuid, Name: username}
	} else if err, ok := result.(error); ok {
		output.err = err
	} else {
		log.Fatal("invalid result type passed")
	}

	return output
}

func createUuidToTextures(
	results []*createUuidToTexturesResult,
) func(uuid string, signed bool) (*mojang.SignedTexturesResponse, error) {
	return func(uuid string, signed bool) (*mojang.SignedTexturesResponse, error) {
		for _, result := range results {
			if result.uuid == uuid {
				return result.result, result.err
			}
		}

		return nil, errors.New("cannot find corresponding result")
	}
}

// https://stackoverflow.com/a/50581165
func randStr(len int) string {
	buff := make([]byte, len)
	rand.Read(buff)
	str := base64.StdEncoding.EncodeToString(buff)

	// Base 64 can be longer than len
	return str[:len]
}
