package handlers

import (
	"encoding/json"
	"github.com/csye7125/team01/internal/store"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type UserHandler struct {
	Store *store.Storage
}

func NewUserHandler(store *store.Storage) *UserHandler {
	return &UserHandler{Store: store}
}

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user store.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if err := h.Store.Users.CreateUser(r.Context(), &user); err != nil {
		if err == gorm.ErrDuplicatedKey {
			http.Error(w, `{"error": "User already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error": "Could not create user"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Step 1: Extract Basic Auth credentials (username & password)
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, `{"error": "Unauthorized: Missing credentials"}`, http.StatusUnauthorized)
		return
	}

	// Step 2: Retrieve the user by username
	user, err := h.Store.Users.GetUserByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, `{"error": "Unauthorized: Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Step 3: Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		http.Error(w, `{"error": "Unauthorized: Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Step 4: Extract user ID from URL
	idStr := chi.URLParam(r, "userId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Step 5: Ensure authenticated user can only access their own data
	if user.ID != uint(id) {
		http.Error(w, `{"error": "Forbidden: You can only access your own user data"}`, http.StatusForbidden)
		return
	}

	// Step 6: Remove password before returning user data
	user.Password = ""

	// Step 7: Return user JSON response
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID from the URL
	idStr := chi.URLParam(r, "userId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Authenticate user using Basic Auth
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Retrieve user from the database using the username
	authUser, err := h.Store.Users.GetUserByUsername(r.Context(), username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(authUser.Password), []byte(password)) != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Ensure that the authenticated user can only update their own details
	if authUser.ID != uint(id) {
		http.Error(w, `{"error": "You are not authorized to update this user"}`, http.StatusForbidden)
		return
	}

	// Decode request body into the update struct
	var updateData store.User
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// Perform the update
	if err := h.Store.Users.UpdateUser(r.Context(), uint(id), &updateData); err != nil {
		http.Error(w, `{"error": "Could not update user"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updateData)
}

func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID from the URL
	idStr := chi.URLParam(r, "userId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Authenticate user using Basic Auth
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Retrieve user from the database using the username
	authUser, err := h.Store.Users.GetUserByUsername(r.Context(), username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(authUser.Password), []byte(password)) != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Ensure that the authenticated user can only delete their own account
	if authUser.ID != uint(id) {
		http.Error(w, `{"error": "You are not authorized to delete this user"}`, http.StatusForbidden)
		return
	}

	// Perform deletion
	if err := h.Store.Users.DeleteUser(r.Context(), uint(id)); err != nil {
		http.Error(w, `{"error": "Could not delete user"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
