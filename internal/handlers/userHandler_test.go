package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/matryer/is"

	"github.com/al-ce/goauth/internal/handlers"
	"github.com/al-ce/goauth/internal/models"
	"github.com/al-ce/goauth/internal/repository"
	"github.com/al-ce/goauth/internal/server"
	"github.com/al-ce/goauth/internal/services"
	"github.com/al-ce/goauth/internal/testutils"
	"github.com/al-ce/goauth/pkg/apperrors"
	"github.com/al-ce/goauth/pkg/config"
)

// TestHandlers_NewUserHandler checks the NewUserHandler constructor returns a valid UsrHandler
func TestHandlers_NewUserHandler(t *testing.T) {
	is := is.New(t)

	t.Run("err on nil user service", func(t *testing.T) {
		uh, err := handlers.NewUserHandler(nil)
		is.Equal(uh, nil)
		is.Equal(err, apperrors.ErrUserServiceIsNil)
	})

	t.Run("creates new user handler", func(t *testing.T) {
		uh := setupUserHandler(t)
		is.True(uh != nil)
	})
}

// TestUserHandle_RegisterUser checks that an http request can add a user to the database
func TestUserHandler_RegisterUser(t *testing.T) {
	is := is.New(t)

	email := "testRegisterUser@test.com"

	t.Run("no email", func(t *testing.T) {
		server := setupServer(t)
		rr, err := makeRequest(
			server.Router,
			"POST",
			"/register",
			UserCredentialsRequest{Password: testutils.TestingPassword},
		)
		is.NoErr(err)
		is.Equal(rr.Code, http.StatusBadRequest)
	})
	t.Run("no password", func(t *testing.T) {
		server := setupServer(t)
		rr, err := makeRequest(
			server.Router,
			"POST",
			"/register",
			UserCredentialsRequest{Email: email},
		)
		is.NoErr(err)
		is.Equal(rr.Code, http.StatusBadRequest)
	})

	t.Run("valid request", func(t *testing.T) {
		server := setupServer(t)
		rr, err := makeRequest(
			server.Router,
			"POST",
			"/register",
			UserCredentialsRequest{Email: email, Password: testutils.TestingPassword},
		)
		is.NoErr(err)
		is.Equal(rr.Code, http.StatusOK)

		// Check user is actually in database
		var user models.User
		result := server.DB.First(&user, "email = ?", email)
		is.NoErr(result.Error)
		is.Equal(user.Email, email)
	})
}

func TestUserHandler_Login(t *testing.T) {
	is := is.New(t)

	server := setupServer(t)

	// Register two test users directly to the DB
	email1 := "testUserHandlerLoginUser@test.com"
	password1 := testutils.TestingPassword // strong password for validator
	email2 := "SECONDARYtestUserHandlerLoginUser@test.com"
	password2 := "SECONDARY" + testutils.TestingPassword
	user1, err := models.NewUser(email1, password1)
	is.NoErr(err)
	user2, err := models.NewUser(email2, password2)
	is.NoErr(err)
	err = server.DB.Create(user1).Error
	is.NoErr(err)
	err = server.DB.Create(user2).Error
	is.NoErr(err)

	t.Run("valid request", func(t *testing.T) {
		rr, err := makeRequest(
			server.Router,
			"POST",
			"/login",
			UserCredentialsRequest{Email: email1, Password: password1},
		)
		is.NoErr(err)
		is.Equal(rr.Code, http.StatusOK)

		// Check cookie is set
		sessionCookie := getSessionCookie(rr)
		is.True(sessionCookie != nil)

		// Session Token is valid
		sessionToken := sessionCookie.Value
		parts := strings.Split(sessionToken, ".")
		is.True(len(parts) == 2)
		sessionID, signature := parts[0], parts[1]
		parsedID, err := uuid.Parse(sessionID)
		is.NoErr(err)
		is.True(models.ValidateSessionID(parsedID, signature))
	})

	t.Run("no email", func(t *testing.T) {
		rr, err := makeRequest(
			server.Router,
			"POST",
			"/login",
			UserCredentialsRequest{Password: password1},
		)
		is.NoErr(err)
		is.Equal(rr.Code, http.StatusBadRequest)
	})
	t.Run("no password", func(t *testing.T) {
		rr, err := makeRequest(
			server.Router,
			"POST",
			"/login",
			UserCredentialsRequest{Email: email1},
		)
		is.NoErr(err)
		is.Equal(rr.Code, http.StatusBadRequest)
	})
	t.Run("non-existent user", func(t *testing.T) {
		rr, err := makeRequest(
			server.Router,
			"POST",
			"/login",
			UserCredentialsRequest{Email: "doesNotExist@test.com", Password: password1},
		)
		is.NoErr(err)
		is.Equal(rr.Code, http.StatusBadRequest)
	})
	t.Run("incorrect password", func(t *testing.T) {
		rr, err := makeRequest(
			server.Router,
			"POST",
			"/login",
			UserCredentialsRequest{Email: email1, Password: "notthepassword"},
		)
		is.NoErr(err)
		is.Equal(rr.Code, http.StatusUnauthorized)
	})
	t.Run("existing password, mismatched existing user", func(t *testing.T) {
		rr, err := makeRequest(
			server.Router,
			"POST",
			"/login",
			UserCredentialsRequest{Email: email1, Password: password2},
		)
		is.NoErr(err)
		is.Equal(rr.Code, http.StatusUnauthorized)
	})
}

func TestUserHandler_Logout(t *testing.T) {
	is := is.New(t)
	server := setupServer(t)

	// Register a test user
	email := "testUserHandlerLogoutUser@test.com"
	user, err := models.NewUser(email, testutils.TestingPassword)
	is.NoErr(err)
	err = server.DB.Create(user).Error
	is.NoErr(err)

	t.Run("valid token", func(t *testing.T) {
		// Login test user
		loginRR, err := makeRequest(
			server.Router,
			"POST",
			"/login",
			UserCredentialsRequest{Email: email, Password: testutils.TestingPassword},
		)
		is.NoErr(err)

		// Get cookie
		sessionCookie := getSessionCookie(loginRR)
		is.True(sessionCookie != nil)

		// Logout
		req, err := http.NewRequest("POST", "/logout", nil)
		is.NoErr(err)
		req.AddCookie(sessionCookie)
		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)
		is.Equal(rr.Code, http.StatusOK)

		// Check cookie is cleared
		logoutCookie := getSessionCookie(rr)
		is.True(logoutCookie != nil)
		is.Equal(logoutCookie.MaxAge, -1)

		// Check response message
		var response map[string]string
		err = json.NewDecoder(rr.Body).Decode(&response)
		is.NoErr(err)
		is.Equal(response["message"], "logged out successfully")
	})

	t.Run("no token", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/logout", nil)
		is.NoErr(err)
		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusUnauthorized)
	})

	t.Run("invalid token", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/logout", nil)
		is.NoErr(err)

		// Create an invalid cookie
		invalidCookie := &http.Cookie{
			Name:     config.SessionCookieName,
			Value:    "invalid-token",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		}
		req.AddCookie(invalidCookie)

		// Perform request
		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)
		is.Equal(rr.Code, http.StatusInternalServerError)
	})
}

func TestUserHandler_LogoutEverywhere(t *testing.T) {
	is := is.New(t)
	server := setupServer(t)

	// Register a test user
	email := "testUserHandlerLogoutEverywhere@test.com"
	user, err := models.NewUser(email, testutils.TestingPassword)
	is.NoErr(err)
	err = server.DB.Create(user).Error
	is.NoErr(err)

	// Login on one "device"
	rr, err := makeRequest(
		server.Router,
		"POST",
		"/login",
		UserCredentialsRequest{Email: email, Password: testutils.TestingPassword},
	)
	is.NoErr(err)
	is.Equal(rr.Code, http.StatusOK)

	// Get token from cookies
	firstCookie := getSessionCookie(rr)
	is.True(firstCookie != nil)
	firstToken := firstCookie.Value
	is.True(firstToken != "")

	// Login on another "device"
	rr, err = makeRequest(
		server.Router,
		"POST",
		"/login",
		UserCredentialsRequest{Email: email, Password: testutils.TestingPassword},
	)
	is.NoErr(err)
	is.Equal(rr.Code, http.StatusOK)

	// Get token from cookies
	secondCookie := getSessionCookie(rr)
	is.True(secondCookie != nil)
	secondToken := secondCookie.Value
	is.True(secondToken != "")
	is.True(firstToken != secondToken)

	t.Run("successfully logout everywhere", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/logouteverywhere", nil)
		is.NoErr(err)

		// Add auth cookie
		req.AddCookie(firstCookie)

		// Logout everywhere
		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)
		is.Equal(rr.Code, http.StatusOK)

		// Cookie is cleared
		clearedCookie := getSessionCookie(rr)
		is.True(clearedCookie != nil)
		is.Equal(clearedCookie.MaxAge, -1) // Cookie should be expired

		// Verify response body
		var response map[string]string
		err = json.NewDecoder(rr.Body).Decode(&response)
		is.NoErr(err)
		is.Equal(response["message"], "logged out everywhere")

		// Check first token is invalidated
		req, err = http.NewRequest("POST", "/logouteverywhere", nil)
		is.NoErr(err)
		req.AddCookie(firstCookie)
		rr = httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)
		is.Equal(rr.Code, http.StatusUnauthorized)

		// Check second token is invalidated
		req, err = http.NewRequest("POST", "/logouteverywhere", nil)
		is.NoErr(err)
		req.AddCookie(secondCookie)
		rr = httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)
		is.Equal(rr.Code, http.StatusUnauthorized)
	})

	t.Run("logout everywhere without token", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/logouteverywhere", nil)
		is.NoErr(err)

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusUnauthorized)
	})

	t.Run("logout everywhere with invalid token", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/logouteverywhere", nil)
		is.NoErr(err)

		req.AddCookie(&http.Cookie{
			Name:     config.SessionCookieName,
			Value:    "invalid-token",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		})

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusUnauthorized)
	})
}

func TestUserHandler_PermanentlyDeleteUser(t *testing.T) {
	is := is.New(t)

	server := setupServer(t)

	// Register a test user
	email := "TestUserHandler_PermanentlyDeleteUser@test.com"
	user, err := models.NewUser(email, testutils.TestingPassword)
	is.NoErr(err)
	err = server.DB.Create(user).Error
	is.NoErr(err)

	// Read registered user from DB so we can get its ID
	var dbUser models.User
	server.DB.First(&dbUser, "email = ?", user.Email)

	t.Run("set userID in gin context", func(t *testing.T) {
		path := "/deleteAccountValid"
		rr := httptest.NewRecorder()
		_, r := gin.CreateTestContext(rr)

		r.DELETE(path, func(c *gin.Context) {
			c.Set("userID", dbUser.ID.String())
			server.HandlerRegistry.User.PermanentlyDeleteUser(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, path, nil)
		r.ServeHTTP(rr, req)
		is.Equal(http.StatusOK, rr.Code)
	})

	t.Run("non-existent user ID", func(t *testing.T) {
		path := "/deleteAccountInvalidID"
		rr := httptest.NewRecorder()
		_, r := gin.CreateTestContext(rr)

		r.DELETE(path, func(c *gin.Context) {
			randUUID := uuid.New()
			c.Set("userID", randUUID.String())
			server.HandlerRegistry.User.PermanentlyDeleteUser(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, path, nil)
		r.ServeHTTP(rr, req)
		is.Equal(http.StatusInternalServerError, rr.Code)
	})

	t.Run("no userID in gin context", func(t *testing.T) {
		path := "/deleteAccountNoUserID"
		rr := httptest.NewRecorder()
		_, r := gin.CreateTestContext(rr)

		r.DELETE(path, func(c *gin.Context) {
			server.HandlerRegistry.User.PermanentlyDeleteUser(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, path, nil)
		r.ServeHTTP(rr, req)
		is.Equal(http.StatusUnauthorized, rr.Code)

		var response map[string]any
		json.Unmarshal(rr.Body.Bytes(), &response)
		is.Equal(0, len(response))
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	is := is.New(t)
	server := setupServer(t)

	// Register a test user
	email := "testUpdateUser@test.com"
	user, err := models.NewUser(email, testutils.TestingPassword)
	is.NoErr(err)
	err = server.DB.Create(user).Error
	is.NoErr(err)

	// Login test user
	_rr, err := makeRequest(
		server.Router,
		"POST",
		"/login",
		UserCredentialsRequest{Email: email, Password: testutils.TestingPassword},
	)
	is.NoErr(err)
	is.Equal(_rr.Code, http.StatusOK)

	// Get token from cookies
	sessionCookie := getSessionCookie(_rr)
	is.True(sessionCookie != nil)
	sessionToken := sessionCookie.Value
	is.True(sessionToken != "")

	t.Run("update email and password", func(t *testing.T) {
		// Create update request body
		newEmail := "newemail2@test.com"
		newPassword := "AnotherSecure" + testutils.TestingPassword
		updateBody := map[string]string{
			"email":    newEmail,
			"password": newPassword,
		}
		jsonData, _ := json.Marshal(updateBody)
		body := bytes.NewBuffer(jsonData)

		// Create request
		req, err := http.NewRequest("POST", "/updateuser", body)
		is.NoErr(err)
		req.Header.Set("Content-Type", "application/json")

		// Get token from cookies
		sessionCookie := getSessionCookie(_rr)
		is.True(sessionCookie != nil)
		sessionToken := sessionCookie.Value
		is.True(sessionToken != "")

		// Add auth cookie to update request
		req.AddCookie(sessionCookie)

		// Make request
		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)
		is.Equal(rr.Code, http.StatusOK)

		// Check we can login with the new credentials
		_rr, err = makeRequest(
			server.Router,
			"POST",
			"/login",
			UserCredentialsRequest{Email: newEmail, Password: newPassword},
		)
		is.NoErr(err)
		is.Equal(_rr.Code, http.StatusOK)
	})

	t.Run("update with empty request", func(t *testing.T) {
		// Create empty request body
		updateBody := map[string]string{}
		jsonData, _ := json.Marshal(updateBody)
		body := bytes.NewBuffer(jsonData)

		// Create request
		req, err := http.NewRequest("POST", "/updateuser", body)
		is.NoErr(err)
		req.Header.Set("Content-Type", "application/json")

		// Add auth cookie
		req.AddCookie(sessionCookie)

		// Make request
		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)
		is.Equal(rr.Code, http.StatusBadRequest)

		var response map[string]string
		err = json.NewDecoder(rr.Body).Decode(&response)
		is.NoErr(err)
		is.Equal(response["error"], "no valid fields provided")
	})

	t.Run("update without auth", func(t *testing.T) {
		// Create request body
		updateBody := map[string]string{
			"email": "valid@email.com",
		}
		jsonData, _ := json.Marshal(updateBody)
		body := bytes.NewBuffer(jsonData)

		// Create request
		req, err := http.NewRequest("POST", "/updateuser", body)
		is.NoErr(err)
		req.Header.Set("Content-Type", "application/json")

		// Make request
		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)
		is.Equal(rr.Code, http.StatusUnauthorized)
	})
}

func setupUserHandler(t *testing.T) *handlers.UserHandler {
	t.Helper()

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	t.Cleanup(func() { tx.Rollback() })

	ur, err := repository.NewUserRepository(tx)
	if err != nil {
		t.Fatalf("failed to create user repository: %v", err)
	}
	sr, err := repository.NewSessionRepository(tx)
	if err != nil {
		t.Fatalf("failed to create session repository: %v", err)
	}
	us, err := services.NewUserService(ur, sr)
	if err != nil {
		t.Fatalf("failed to create user service: %v", err)
	}
	uh, err := handlers.NewUserHandler(us)
	if err != nil {
		t.Fatalf("failed to create user handler: %v", err)
	}
	return uh
}

func setupServer(t *testing.T) *server.APIServer {
	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	t.Cleanup(func() { tx.Rollback() })

	server, err := server.NewAPIServer(tx)
	if err != nil {
		t.Fatalf("failed to init api server: %v", err)
	}
	server.SetupRoutes()
	return server
}

func makeRequest(router *gin.Engine, method, path string, body any) (*httptest.ResponseRecorder, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		method,
		path,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr, nil
}

func getSessionCookie(rr *httptest.ResponseRecorder) *http.Cookie {
	var sessionCookie *http.Cookie
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == config.SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	return sessionCookie
}
