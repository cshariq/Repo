package middlewares

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/hackclub/hackatime/config"

	"github.com/hackclub/hackatime/mocks"
	"github.com/hackclub/hackatime/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticateMiddleware_tryGetUserByApiKeyHeader_Success(t *testing.T) {
	testApiKey := "86648d74-19c5-452b-ba01-fb3ec70d4c2f"
	testToken := base64.StdEncoding.EncodeToString([]byte(testApiKey))
	testUser := &models.User{ApiKey: testApiKey}

	mockRequest := &http.Request{
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("Basic %s", testToken)},
		},
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByKey", testApiKey).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, err := sut.tryGetUserByApiKeyHeader(mockRequest)

	assert.Nil(t, err)
	assert.Equal(t, testUser, result)
}

func TestAuthenticateMiddleware_tryGetUserByApiKeyHeader_Invalid(t *testing.T) {
	testApiKey := "86648d74-19c5-452b-ba01-fb3ec70d4c2f"
	testToken := base64.StdEncoding.EncodeToString([]byte(testApiKey))

	mockRequest := &http.Request{
		Header: http.Header{
			// 'Basic' prefix missing here
			"Authorization": []string{testToken},
		},
	}

	userServiceMock := new(mocks.UserServiceMock)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, err := sut.tryGetUserByApiKeyHeader(mockRequest)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByApiKeyQuery_Success(t *testing.T) {
	testApiKey := "86648d74-19c5-452b-ba01-fb3ec70d4c2f"
	testUser := &models.User{ApiKey: testApiKey}

	params := url.Values{}
	params.Add("api_key", testApiKey)
	mockRequest := &http.Request{
		URL: &url.URL{
			RawQuery: params.Encode(),
		},
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByKey", testApiKey).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, err := sut.tryGetUserByApiKeyQuery(mockRequest)

	assert.Nil(t, err)
	assert.Equal(t, testUser, result)
}

func TestAuthenticateMiddleware_tryGetUserByApiKeyQuery_Invalid(t *testing.T) {
	testApiKey := "86648d74-19c5-452b-ba01-fb3ec70d4c2f"

	params := url.Values{}
	params.Add("token", testApiKey)
	mockRequest := &http.Request{
		URL: &url.URL{
			RawQuery: params.Encode(),
		},
	}

	userServiceMock := new(mocks.UserServiceMock)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, actualErr := sut.tryGetUserByApiKeyQuery(mockRequest)

	assert.Error(t, actualErr)
	assert.Equal(t, errEmptyKey, actualErr)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByTrustedHeader_Disabled(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.TrustedHeaderAuth = false
	cfg.Security.TrustedHeaderAuthKey = "Remote-User"
	cfg.Security.TrustReverseProxyIps = "127.0.0.1,::1"
	cfg.Security.ParseTrustReverseProxyIPs()
	config.Set(cfg)

	testUser := &models.User{ID: "user01"}

	mockRequest := &http.Request{
		Header:     http.Header{"Remote-User": []string{testUser.ID}},
		RemoteAddr: "127.0.0.1:54654",
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserById", testUser.ID).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest)
	assert.Error(t, actualErr)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByTrustedHeader_Untrusted(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.TrustedHeaderAuth = true
	cfg.Security.TrustedHeaderAuthKey = "Remote-User"
	cfg.Security.TrustReverseProxyIps = "127.0.0.1,::1,192.168.0.1,192.168.178.0/24,33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafb,1ddc:e2d6:dcce:ab6c::1/64"
	cfg.Security.ParseTrustReverseProxyIPs()
	config.Set(cfg)

	testIps := []string{"192.168.0.2", "192.168.179.35", "[33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafa]", "[1ddc:e2d6:dcce:ab6b:1ba7:7aaa:58dc:a42b]"} // none of these should be authorized
	testUser := &models.User{ID: "user01"}

	for _, ip := range testIps {
		mockRequest := &http.Request{
			Header: http.Header{
				"Remote-User":     []string{testUser.ID},
				"X-Forwarded-For": []string{"192.168.0.1"}, // forward for some trusted ip -> header should be ignored for auth. checks, because only actual reverse proxy must be legitimized
			},
			RemoteAddr: fmt.Sprintf("%s:54654", ip),
		}

		userServiceMock := new(mocks.UserServiceMock)
		userServiceMock.On("GetUserById", testUser.ID).Return(testUser, nil)

		sut := NewAuthenticateMiddleware(userServiceMock)

		result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest)
		assert.Error(t, actualErr)
		assert.Nil(t, result)
	}
}

func TestAuthenticateMiddleware_tryGetUserByTrustedHeader_Success(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.TrustedHeaderAuth = true
	cfg.Security.TrustedHeaderAuthKey = "Remote-User"
	cfg.Security.TrustReverseProxyIps = "127.0.0.1,::1,192.168.0.1,192.168.178.0/24,33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafb,1ddc:e2d6:dcce:ab6c::1/64"
	cfg.Security.ParseTrustReverseProxyIPs()
	config.Set(cfg)

	testIps := []string{"127.0.0.1", "[::1]", "192.168.0.1", "192.168.178.1", "192.168.178.35", "[33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafb]", "[1ddc:e2d6:dcce:ab6c:2ba7:7aaa:58dc:a42b]", "[1ddc:e2d6:dcce:ab6c:1ba7:7aaa:58dc:a42b]"} // all of these should be authorized
	testUser := &models.User{ID: "user01"}

	for _, ip := range testIps {
		mockRequest := &http.Request{
			Header:     http.Header{"Remote-User": []string{testUser.ID}},
			RemoteAddr: fmt.Sprintf("%s:54654", ip),
		}

		userServiceMock := new(mocks.UserServiceMock)
		userServiceMock.On("GetUserById", testUser.ID).Return(testUser, nil)

		sut := NewAuthenticateMiddleware(userServiceMock)

		result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest)
		assert.Equal(t, testUser, result)
		assert.Nil(t, actualErr)
	}
}

// TODO: somehow test cookie auth function