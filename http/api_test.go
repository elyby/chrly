package http

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/elyby/chrly/auth"
	"github.com/elyby/chrly/db"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"
)

func TestConfig_PostSkin(t *testing.T) {
	t.Run("Upload new identity with textures info", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		resultModel := createSkinModel("mock_user", false)
		resultModel.SkinId = 5
		resultModel.Hash = "94a457d92a61460cb9cb5d6f29732d2a"
		resultModel.Url = "http://ely.by/minecraft/skins/default.png"
		resultModel.MojangTextures = ""
		resultModel.MojangSignature = ""

		form := url.Values{
			"identityId": {"1"},
			"username":   {"mock_user"},
			"uuid":       {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":     {"5"},
			"hash":       {"94a457d92a61460cb9cb5d6f29732d2a"},
			"is1_8":      {"0"},
			"isSlim":     {"0"},
			"url":        {"http://ely.by/minecraft/skins/default.png"},
		}

		req := httptest.NewRequest("POST", "http://skinsystem.ely.by/api/skins", bytes.NewBufferString(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(nil)
		mocks.Skins.EXPECT().FindByUserId(1).Return(nil, &db.SkinNotFoundError{Who: "unknown"})
		mocks.Skins.EXPECT().FindByUsername("mock_user").Return(nil, &db.SkinNotFoundError{Who: "mock_user"})
		mocks.Skins.EXPECT().Save(resultModel).Return(nil)
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.success", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.request", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.success", int64(1))

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(201, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.Empty(response)
	})

	t.Run("Upload new identity with skin file", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, _ := writer.CreateFormFile("skin", "char.png")
		_, _ = part.Write(loadSkinFile())

		_ = writer.WriteField("identityId", "1")
		_ = writer.WriteField("username", "mock_user")
		_ = writer.WriteField("uuid", "0f657aa8-bfbe-415d-b700-5750090d3af3")
		_ = writer.WriteField("skinId", "5")

		err := writer.Close()
		if err != nil {
			panic(err)
		}

		req := httptest.NewRequest("POST", "http://skinsystem.ely.by/api/skins", body)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(nil)
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.success", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.request", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.validation_failed", int64(1))

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(400, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`{
			"errors": {
				"skin": [
					"Skin uploading is temporary unavailable"
				]
			}
		}`, string(response))
	})

	t.Run("Keep the same identityId, uuid and username, but change textures information", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		resultModel := createSkinModel("mock_user", false)
		resultModel.SkinId = 5
		resultModel.Hash = "94a457d92a61460cb9cb5d6f29732d2a"
		resultModel.Url = "http://textures-server.com/skin.png"
		resultModel.MojangTextures = ""
		resultModel.MojangSignature = ""

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(nil)
		mocks.Skins.EXPECT().FindByUserId(1).Return(createSkinModel("mock_user", false), nil)
		mocks.Skins.EXPECT().Save(resultModel).Return(nil)
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.success", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.request", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.success", int64(1))

		form := url.Values{
			"identityId": {"1"},
			"username":   {"mock_user"},
			"uuid":       {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":     {"5"},
			"hash":       {"94a457d92a61460cb9cb5d6f29732d2a"},
			"is1_8":      {"0"},
			"isSlim":     {"0"},
			"url":        {"http://textures-server.com/skin.png"},
		}

		req := httptest.NewRequest("POST", "http://chrly/api/skins", bytes.NewBufferString(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(201, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.Empty(response)
	})

	t.Run("Keep the same uuid and username, but change identityId", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		resultModel := createSkinModel("mock_user", false)
		resultModel.UserId = 2
		resultModel.SkinId = 5
		resultModel.Hash = "94a457d92a61460cb9cb5d6f29732d2a"
		resultModel.Url = "http://ely.by/minecraft/skins/default.png"
		resultModel.MojangTextures = ""
		resultModel.MojangSignature = ""

		form := url.Values{
			"identityId": {"2"},
			"username":   {"mock_user"},
			"uuid":       {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":     {"5"},
			"hash":       {"94a457d92a61460cb9cb5d6f29732d2a"},
			"is1_8":      {"0"},
			"isSlim":     {"0"},
			"url":        {"http://ely.by/minecraft/skins/default.png"},
		}

		req := httptest.NewRequest("POST", "http://skinsystem.ely.by/api/skins", bytes.NewBufferString(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(nil)
		mocks.Skins.EXPECT().FindByUserId(2).Return(nil, &db.SkinNotFoundError{Who: "unknown"})
		mocks.Skins.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", false), nil)
		mocks.Skins.EXPECT().RemoveByUsername("mock_user").Return(nil)
		mocks.Skins.EXPECT().Save(resultModel).Return(nil)
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.success", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.request", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.success", int64(1))

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(201, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.Empty(response)
	})

	t.Run("Keep the same identityId and uuid, but change username", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		resultModel := createSkinModel("changed_username", false)
		resultModel.SkinId = 5
		resultModel.Hash = "94a457d92a61460cb9cb5d6f29732d2a"
		resultModel.Url = "http://ely.by/minecraft/skins/default.png"
		resultModel.MojangTextures = ""
		resultModel.MojangSignature = ""

		form := url.Values{
			"identityId": {"1"},
			"username":   {"changed_username"},
			"uuid":       {"0f657aa8-bfbe-415d-b700-5750090d3af3"},
			"skinId":     {"5"},
			"hash":       {"94a457d92a61460cb9cb5d6f29732d2a"},
			"is1_8":      {"0"},
			"isSlim":     {"0"},
			"url":        {"http://ely.by/minecraft/skins/default.png"},
		}

		req := httptest.NewRequest("POST", "http://skinsystem.ely.by/api/skins", bytes.NewBufferString(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(nil)
		mocks.Skins.EXPECT().FindByUserId(1).Return(createSkinModel("mock_user", false), nil)
		mocks.Skins.EXPECT().RemoveByUserId(1).Return(nil)
		mocks.Skins.EXPECT().Save(resultModel).Return(nil)
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.success", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.request", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.success", int64(1))

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(201, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.Empty(response)
	})

	t.Run("Get errors about required fields", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		form := url.Values{
			"mojangTextures": {"someBase64EncodedString"},
		}

		req := httptest.NewRequest("POST", "http://skinsystem.ely.by/api/skins", bytes.NewBufferString(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(nil)
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.success", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.request", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.post.validation_failed", int64(1))

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(400, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`{
			"errors": {
				"identityId": [
					"The identityId field is required",
					"The identityId field must be numeric",
					"The identityId field must be minimum 1 char"
				],
				"skinId": [
					"The skinId field is required",
					"The skinId field must be numeric",
					"The skinId field must be minimum 1 char"
				],
				"username": [
					"The username field is required"
				],
				"uuid": [
					"The uuid field is required",
					"The uuid field must contain valid UUID"
				],
				"url": [
					"One of url or skin should be provided, but not both"
				],
				"skin": [
					"One of url or skin should be provided, but not both"
				],
				"mojangSignature": [
					"The mojangSignature field is required"
				]
			}
		}`, string(response))
	})

	t.Run("Perform request without authorization", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		req := httptest.NewRequest("POST", "http://skinsystem.ely.by/api/skins", nil)
		req.Header.Add("Authorization", "Bearer invalid.jwt.token")
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(&auth.Unauthorized{Reason: "Cannot parse passed JWT token"})
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.failed", int64(1))

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(403, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`{
			"error": "Cannot parse passed JWT token"
		}`, string(response))
	})
}

func TestConfig_DeleteSkinByUserId(t *testing.T) {
	t.Run("Delete skin by its identity id", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		req := httptest.NewRequest("DELETE", "http://skinsystem.ely.by/api/skins/id:1", nil)
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(nil)
		mocks.Skins.EXPECT().FindByUserId(1).Return(createSkinModel("mock_user", false), nil)
		mocks.Skins.EXPECT().RemoveByUserId(1).Return(nil)
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.success", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.delete.request", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.delete.success", int64(1))

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(204, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.Empty(response)
	})

	t.Run("Try to remove not exists identity id", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		req := httptest.NewRequest("DELETE", "http://skinsystem.ely.by/api/skins/id:2", nil)
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(nil)
		mocks.Skins.EXPECT().FindByUserId(2).Return(nil, &db.SkinNotFoundError{Who: "unknown"})
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.success", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.delete.request", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.delete.not_found", int64(1))

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(404, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`[
			"Cannot find record for requested user id"
		]`, string(response))
	})
}

func TestConfig_DeleteSkinByUsername(t *testing.T) {
	t.Run("Delete skin by its identity username", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		req := httptest.NewRequest("DELETE", "http://skinsystem.ely.by/api/skins/mock_user", nil)
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(nil)
		mocks.Skins.EXPECT().FindByUsername("mock_user").Return(createSkinModel("mock_user", false), nil)
		mocks.Skins.EXPECT().RemoveByUserId(1).Return(nil)
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.success", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.delete.request", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.delete.success", int64(1))

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(204, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.Empty(response)
	})

	t.Run("Try to remove not exists identity username", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		req := httptest.NewRequest("DELETE", "http://skinsystem.ely.by/api/skins/mock_user_2", nil)
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(nil)
		mocks.Skins.EXPECT().FindByUsername("mock_user_2").Return(nil, &db.SkinNotFoundError{Who: "mock_user_2"})
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.success", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.delete.request", int64(1))
		mocks.Log.EXPECT().IncCounter("api.skins.delete.not_found", int64(1))

		config.CreateHandler().ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(404, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`[
		"Cannot find record for requested username"
	]`, string(response))
	})
}

func TestConfig_Authenticate(t *testing.T) {
	t.Run("Test behavior when signing key is not set", func(t *testing.T) {
		assert := testify.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		config, mocks := setupMocks(ctrl)

		req := httptest.NewRequest("POST", "http://localhost", nil)
		w := httptest.NewRecorder()

		mocks.Auth.EXPECT().Check(gomock.Any()).Return(&auth.Unauthorized{Reason: "signing key not available"})
		mocks.Log.EXPECT().IncCounter("authentication.challenge", int64(1))
		mocks.Log.EXPECT().IncCounter("authentication.failed", int64(1))

		res := config.AuthenticationMiddleware(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {}))
		res.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(403, resp.StatusCode)
		response, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(`{
			"error": "signing key not available"
		}`, string(response))
	})
}

// base64 https://github.com/mathiasbynens/small/blob/0ca3c51/png-transparent.png
var OnePxPng = []byte("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQottAAAAABJRU5ErkJggg==")

func loadSkinFile() []byte {
	result := make([]byte, 92)
	_, err := base64.StdEncoding.Decode(result, OnePxPng)
	if err != nil {
		panic(err)
	}

	return result
}
