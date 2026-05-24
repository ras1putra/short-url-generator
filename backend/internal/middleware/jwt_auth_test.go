package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/token"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestJWTAuthEnv(t *testing.T) (*repository.Queries, *testutil.FakeCacher) {
	_, queries := testutil.SetupTestDB(t)
	fakeCache := testutil.NewFakeCacher()
	return queries, fakeCache
}

func createUserWithRole(t *testing.T, queries *repository.Queries, role string) repository.User {
	user, err := queries.CreateUser(context.Background(), repository.CreateUserParams{
		Name:     "Test " + role,
		Email:    role + "-" + uuid.New().String() + "@example.com",
		Password: "hashed",
		Role:     role,
	})
	require.NoError(t, err)
	return user
}

func TestJWTAuth_NoToken(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	queries, fakeCache := newTestJWTAuthEnv(t)

	app := fiber.New()
	app.Use(JWTAuth("secret", fakeCache, queries))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestJWTAuth_ValidToken(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	secret := "test-secret"

	queries, fakeCache := newTestJWTAuthEnv(t)
	user := createUserWithRole(t, queries, "user")

	tokenStr, err := token.IssueToken(user.ID.String(), user.Role, secret, "access", 15*time.Minute)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(JWTAuth(secret, fakeCache, queries))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString(c.Locals("user_id").(string))
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: tokenStr})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	queries, fakeCache := newTestJWTAuthEnv(t)

	app := fiber.New()
	app.Use(JWTAuth("secret", fakeCache, queries))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: "invalid-token"})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestJWTAuth_RevokedToken(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	secret := "test-secret"

	queries, fakeCache := newTestJWTAuthEnv(t)
	user := createUserWithRole(t, queries, "user")

	tokenStr, err := token.IssueToken(user.ID.String(), user.Role, secret, "access", 15*time.Minute)
	require.NoError(t, err)

	// Revoke the token by adding it to fake cache
	err = fakeCache.Set(nil, constants.RedisPrefixBlacklist+tokenStr, "1", 15*time.Minute)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(JWTAuth(secret, fakeCache, queries))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: tokenStr})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestJWTAuth_WrongTokenType(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	secret := "test-secret"

	queries, fakeCache := newTestJWTAuthEnv(t)

	refreshToken, err := token.IssueToken(uuid.New().String(), "user", secret, "refresh", 7*24*time.Hour)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(JWTAuth(secret, fakeCache, queries))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: refreshToken})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestRequireRole(t *testing.T) {
	secret := "test-secret"
	queries, fakeCache := newTestJWTAuthEnv(t)

	app := fiber.New()
	app.Use(JWTAuth(secret, fakeCache, queries))

	app.Get("/admin", RequireRole("admin"), func(c *fiber.Ctx) error {
		return c.SendString("admin-ok")
	})

	t.Run("Authorized as admin", func(t *testing.T) {
		user := createUserWithRole(t, queries, "admin")
		tokenStr, _ := token.IssueToken(user.ID.String(), user.Role, secret, "access", time.Hour)
		req := httptest.NewRequest("GET", "/admin", nil)
		req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: tokenStr})
		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Unauthorized as user", func(t *testing.T) {
		user := createUserWithRole(t, queries, "user")
		tokenStr, _ := token.IssueToken(user.ID.String(), user.Role, secret, "access", time.Hour)
		req := httptest.NewRequest("GET", "/admin", nil)
		req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: tokenStr})
		resp, _ := app.Test(req)
		assert.Equal(t, 403, resp.StatusCode)
	})
}
