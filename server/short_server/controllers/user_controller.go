package controllers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/beheryahmed1991/ClipsStream/server/short_server/database"
	model "github.com/beheryahmed1991/ClipsStream/server/short_server/models"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type (
	GetUsersOutput struct {
		Body []model.User `json:"body"`
	}

	GetUserOutput struct {
		Body model.User `json:"body"`
	}

	GetUserInput struct {
		ID string `path:"id"`
	}

	AddUserInput struct {
		Body AddUserRequestBody
	}

	AddUserOutput struct {
		Body model.User `json:"body"`
	}
	//DTO
	AddUserRequestBody struct {
		FirstName       string        `json:"first_name" validate:"required,min=2,max=100"`
		LastName        string        `json:"last_name" validate:"required,min=2,max=100"`
		Email           string        `json:"email" validate:"required,email"`
		Password        string        `json:"password" validate:"required,min=6"`
		Role            string        `json:"role" validate:"required,oneof=ADMIN USER"`
		Token           string        `json:"token" bson:"token"`
		RefreshToken    string        `json:"refresh_token" bson:"refresh_token"`
		FavouriteGenres []model.Genre `json:"favourite_genres" validate:"required,dive"`
	}
)

func RegisterUserRoutes(api huma.API) {
	huma.Get(api, "/users", GetUsers)
	huma.Register(api, huma.Operation{
		OperationID: "get-user",
		Method:      "GET",
		Path:        "/users/{id}",
		Summary:     "Get one user by ID",
		Errors:      []int{400, 404, 500},
	}, GetUser)
	huma.Register(api, huma.Operation{
		OperationID:   "add-user",
		Method:        "POST",
		Path:          "/users",
		Summary:       "Add one user",
		DefaultStatus: http.StatusCreated,
		Errors:        []int{400, 409, 500},
	}, AddUser)
}

func getUserCol() (*mongo.Collection, error) {
	return database.OpenCollection("users")
}

func GetUsers(ctx context.Context, in *struct{}) (*GetUsersOutput, error) {
	col, err := getUserCol()
	if err != nil {
		slog.Error("open users collection failed", "op", "GetUsers", "err", err)
		return nil, fmt.Errorf("open users collection: %w", err)
	}

	qctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := col.Find(qctx, bson.M{})
	if err != nil {
		slog.Error("find users failed", "op", "GetUsers", "err", err)
		return nil, fmt.Errorf("find users: %w", err)
	}
	defer cursor.Close(qctx)

	users := make([]model.User, 0)
	if err := cursor.All(qctx, &users); err != nil {
		slog.Error("decode users failed", "op", "GetUsers", "err", err)
		return nil, fmt.Errorf("decode users: %w", err)
	}

	return &GetUsersOutput{Body: users}, nil
}

func GetUser(ctx context.Context, in *GetUserInput) (*GetUserOutput, error) {
	objID, err := bson.ObjectIDFromHex(in.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid user ID")
	}
	col, err := getUserCol()
	if err != nil {
		slog.Error("open users collection failed", "op", "GetUser", "user_id", in.ID, "err", err)
		return nil, fmt.Errorf("open users collection: %w", err)
	}

	qctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var user model.User
	if err := col.FindOne(qctx, bson.M{"_id": objID}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, huma.Error404NotFound("user not found")
		}
		slog.Error("find user failed", "op", "GetUser", "user_id", in.ID, "err", err)
		return nil, fmt.Errorf("find user: %w", err)
	}

	return &GetUserOutput{Body: user}, nil
}

func AddUser(ctx context.Context, in *AddUserInput) (*AddUserOutput, error) {
	if err := validate.Struct(in.Body); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			details := make([]error, len(ve))
			for i, fe := range ve {
				details[i] = fmt.Errorf("field '%s' failed '%s'", fe.Field(), fe.Tag())
			}
			return nil, huma.Error400BadRequest("validation failed", details...)
		}
		return nil, huma.Error400BadRequest("validation failed")
	}

	col, err := getUserCol()
	if err != nil {
		slog.Error("open users collection failed", "op", "AddUser", "err", err)
		return nil, fmt.Errorf("open users collection: %w", err)
	}

	qctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	hashedPassword, err := HashPassword(in.Body.Password)
	if err != nil {
		slog.Error("hash password failed", "op", "AddUser", "email", in.Body.Email, "err", err)
		return nil, huma.Error500InternalServerError("failed to secure user password")
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(in.Body.Email))

	user := model.User{
		FirstName:       in.Body.FirstName,
		LastName:        in.Body.LastName,
		Email:           normalizedEmail,
		Password:        hashedPassword,
		Role:            in.Body.Role,
		Token:           in.Body.Token,
		RefreshToken:    in.Body.RefreshToken,
		FavouriteGenres: in.Body.FavouriteGenres,
	}
	assignUserIdentityAndTimestamps(&user)

	if _, err := col.InsertOne(qctx, user); err != nil {
		if isDuplicateKeyError(err) {
			return nil, huma.Error409Conflict("user already registered")
		}
		slog.Error("insert user failed", "op", "AddUser", "email", user.Email, "err", err)
		return nil, fmt.Errorf("insert user: %w", err)
	}

	// Do not return password hash to clients.
	user.Password = ""
	return &AddUserOutput{Body: user}, nil
}

func assignUserIdentityAndTimestamps(user *model.User) {
	user.ID = bson.NewObjectID()
	if user.UserID == "" {
		user.UserID = user.ID.Hex()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func isDuplicateKeyError(err error) bool {
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		for _, e := range writeErr.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}

	var bulkWriteErr mongo.BulkWriteException
	if errors.As(err, &bulkWriteErr) {
		for _, e := range bulkWriteErr.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}
	return false
}
