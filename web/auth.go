package web

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	jwtSecretLen = 32
	tokenExpiry  = 7 * 24 * time.Hour
)

// User represents a web UI user
type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

// AuthManager handles user auth
type AuthManager struct {
	usersPath string
	jwtSecret []byte
}

// NewAuthManager creates auth manager, ensures jwt secret exists
func NewAuthManager() (*AuthManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".xray-go")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	secretPath := filepath.Join(dir, "jwt-secret")
	var secret []byte
	if data, err := os.ReadFile(secretPath); err == nil && len(data) == jwtSecretLen {
		secret = data
	} else {
		secret = make([]byte, jwtSecretLen)
		if _, err := rand.Read(secret); err != nil {
			return nil, err
		}
		if err := os.WriteFile(secretPath, secret, 0600); err != nil {
			return nil, err
		}
	}

	return &AuthManager{
		usersPath: filepath.Join(dir, "web-users.json"),
		jwtSecret: secret,
	}, nil
}

// HasUser returns true if at least one user exists
func (am *AuthManager) HasUser() bool {
	users, _ := am.loadUsers()
	return len(users) > 0
}

// CreateUser creates a new user with bcrypt-hashed password
func (am *AuthManager) CreateUser(username, password string) error {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return fmt.Errorf("username and password required")
	}
	users, err := am.loadUsers()
	if err != nil {
		return err
	}
	for _, u := range users {
		if u.Username == username {
			return fmt.Errorf("user already exists")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	users = append(users, &User{Username: username, PasswordHash: string(hash)})
	return am.saveUsers(users)
}

// ValidateUser checks username/password and returns JWT token
func (am *AuthManager) ValidateUser(username, password string) (string, error) {
	users, err := am.loadUsers()
	if err != nil {
		return "", err
	}
	for _, u := range users {
		if u.Username == username {
			if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
				return "", fmt.Errorf("invalid credentials")
			}
			return am.generateToken(username)
		}
	}
	return "", fmt.Errorf("invalid credentials")
}

// ValidateToken validates JWT and returns username
func (am *AuthManager) ValidateToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token")
	}
	return am.parseToken(token)
}

func (am *AuthManager) extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

func (am *AuthManager) loadUsers() ([]*User, error) {
	data, err := os.ReadFile(am.usersPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*User{}, nil
		}
		return nil, err
	}
	var users []*User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (am *AuthManager) saveUsers(users []*User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(am.usersPath, data, 0600)
}

type jwtPayload struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
}

func (am *AuthManager) generateToken(username string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadBytes, _ := json.Marshal(jwtPayload{
		Sub: username,
		Exp: time.Now().Add(tokenExpiry).Unix(),
	})
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	message := header + "." + payload
	mac := hmac.New(sha256.New, am.jwtSecret)
	mac.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return message + "." + signature, nil
}

func (am *AuthManager) parseToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}
	message := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, am.jwtSecret)
	mac.Write([]byte(message))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return "", fmt.Errorf("invalid signature")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	var p jwtPayload
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		return "", err
	}
	if time.Now().Unix() > p.Exp {
		return "", fmt.Errorf("token expired")
	}
	return p.Sub, nil
}
