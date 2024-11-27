package db

import (
	"context"
	"strings"
	"time"

	"github.com/adrieljss/go-serverus/result"
	"github.com/georgysavva/scany/v2/pgxscan"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/jackc/pgerrcode"
)

// CREATE TABLE IF NOT EXISTS users (
// 	uid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
// 	user_id VARCHAR(35) UNIQUE NOT NULL,
// 	username VARCHAR(35) NOT NULL,
// 	email VARCHAR(50) UNIQUE NOT NULL,
// 	password TEXT NOT NULL,
// 	biography VARCHAR(200),
// 	pfp_url TEXT,
// 	last_updated TIMESTAMP NOT NULL DEFAULT now(),
// 	created_at TIMESTAMP NOT NULL DEFAULT now(),
// 	role_level int4 DEFAULT 0 -- priviledges etc
//   );

type User struct {
	Uid         string     `json:"uid"`
	UserId      string     `json:"user_id"`
	Username    string     `json:"username"`
	Email       string     `json:"email,omitempty"`
	Password    string     `json:"-"`
	Biography   *string    `json:"biography"`
	PfpUrl      *string    `json:"pfp_url"`
	CreatedAt   *time.Time `json:"created_at"`
	LastUpdated *time.Time `json:"last_updated"`
	RoleLevel   int        `json:"role_level"`
}

// Required fields for users.
type RequiredUser struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	UserIdUniqueConstraint = "users_user_id_key"
	EmailUniqueConstraint  = "users_email_key"
)

func (a RequiredUser) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.UserId, validation.Required, validation.Length(3, 35), is.Alphanumeric),
		validation.Field(&a.Username, validation.Required, validation.Length(2, 35)),
		validation.Field(&a.Email, validation.Required, is.Email),
		validation.Field(&a.Password, validation.Required, validation.Length(5, 30)),
	)
}

// Clears out private information: emails and passwords.
func (user *User) ClearPrivateInfo() {
	user.Email = ""
	user.Password = ""
}

func UserExistsUserId(ctx context.Context, userId string) (bool, *result.Error) {
	var user User
	err := FetchOne(ctx, DB, &user, "select * from users where user_id = $1", userId)

	if pgxscan.NotFound(err) {
		return false, nil
	} else if err != nil {
		return false, result.ServerErr(err)
	}
	return true, nil
}

func UserExistsEmail(ctx context.Context, email string) (bool, *result.Error) {
	var user User
	err := FetchOne(ctx, DB, &user, "select * from users where email = $1", email)

	if pgxscan.NotFound(err) {
		return false, nil
	} else if err != nil {
		return false, result.ServerErr(err)
	}
	return true, nil
}

func FetchUserByUserId(ctx context.Context, userId string) (*User, *result.Error) {
	var user User
	err := FetchOne(ctx, DB, &user, "select * from users where user_id = $1", userId)

	if pgxscan.NotFound(err) {
		return nil, result.Err(404, err, "USER_NOT_FOUND", "user with the given user id is not found")
	} else if err != nil {
		return nil, result.ServerErr(err)
	}

	return &user, nil
}

func FetchUserByUid(ctx context.Context, uid string) (*User, *result.Error) {
	var user User
	err := FetchOne(ctx, DB, &user, "select * from users where uid = $1", uid)

	if pgxscan.NotFound(err) {
		return nil, result.Err(404, err, "USER_NOT_FOUND", "user with the given uid is not found")
	} else if err != nil {
		return nil, result.ServerErr(err)
	}

	return &user, nil
}

func FetchUserByEmail(ctx context.Context, email string) (*User, *result.Error) {
	var user User
	err := FetchOne(ctx, DB, &user, "select * from users where email = $1", email)

	if pgxscan.NotFound(err) {
		return nil, result.Err(404, err, "USER_NOT_FOUND", "user with the given email is not found")
	} else if err != nil {
		return nil, result.ServerErr(err)
	}

	return &user, nil
}

func FetchUserByUidAndHashedPassword(ctx context.Context, uid string, hashed_password string) (*User, *result.Error) {
	var user User
	err := FetchOne(ctx, DB, &user, "select * from users where uid = $1 AND password = $2", uid, hashed_password)

	if pgxscan.NotFound(err) {
		return nil, result.Err(404, err, "USER_NOT_FOUND", "user with the given uid is not found")
	} else if err != nil {
		return nil, result.ServerErr(err)
	}

	return &user, nil
}

// The `emailOrUserId` parameter checks if the string has a '@' in it, since Uid doesnt allow any special characters.
func FetchUserByCredentials(ctx context.Context, emailOrUserId string, password string) (*User, *result.Error) {
	var user User
	var err error
	if strings.Contains(emailOrUserId, "@") {
		// email mode
		err = FetchOne(ctx, DB, &user, "select * from users where email = $1 and password = crypt($2, password)", emailOrUserId, password)
	} else {
		// uid mode
		err = FetchOne(ctx, DB, &user, "select * from users where user_id = $1 and password = crypt($2, password)", emailOrUserId, password)
	}

	if pgxscan.NotFound(err) {
		return nil, result.ErrWithMetadata(404, err, "USER_NOT_FOUND", "user with the given credentials do not exist", map[string]string{
			"email_or_user_id": "user with the given credentials do not exist",
			"password":         "user with the given credentials do not exist",
		})
	} else if err != nil {
		return nil, result.ServerErr(err)
	}

	return &user, nil
}

// Check for unique violation errors (in updates or inserts) and returns correspoding errors or nil.
// Only checks for user_id and email.
func CheckDupeAndServerErr(err error) *result.Error {
	if PgErrorIs(err, pgerrcode.UniqueViolation) {
		if res := ToPgError(err).ConstraintName; res != "" {
			if res == UserIdUniqueConstraint {
				return result.ErrWithMetadata(400, err, "USER_ID_TAKEN", "a user with the same user id already exists", map[string]string{
					"user_id": "user id already exists",
				})
			} else if res == EmailUniqueConstraint {
				return result.ErrWithMetadata(400, err, "EMAIL_TAKEN", "a user with the same email already exists", map[string]string{
					"email": "email already exists",
				})
			} else {
				return result.ServerErr(err)
			}
		} else {
			return result.ServerErr(err)
		}
	}
	return nil
}

func CreateUser(ctx context.Context, userId string, username string, email string, password string) (*User, *result.Error) {
	var user User

	err := FetchOne(ctx,
		DB, &user,
		"INSERT INTO users (user_id, username, email, password) VALUES ($1,$2,$3,crypt($4, gen_salt('bf'))) RETURNING *",
		userId, username, email, password)

	derr := CheckDupeAndServerErr(err)
	if derr != nil {
		return nil, derr
	}
	return &user, nil
}

// Updates a user (no credentials).
// If a field is empty then it will use the previous value.
// Only use in protected routes.
func UpdateUserInfo(ctx context.Context, current_user User, user_id string, username string, biography string, pfp_url string) (*User, *result.Error) {
	var user User
	// TODO
	if user_id != "" {
		current_user.UserId = user_id
	}

	if username != "" {
		current_user.Username = username
	}

	if biography != "" {
		current_user.Biography = &biography
	}

	if pfp_url != "" {
		current_user.PfpUrl = &pfp_url
	}

	if user_id == "" && username == "" && biography == "" && pfp_url == "" {
		return nil, result.Err(400, nil, "ZERO_UPDATE_FIELDS", "a field must be updated")
	}

	err := FetchOne(ctx, DB, &user, "UPDATE users SET user_id=$1, username=$2, pfp_url=$3, biography=$4 WHERE uid=$5 RETURNING *",
		current_user.UserId, current_user.Username, current_user.PfpUrl, current_user.Biography, current_user.Uid)

	derr := CheckDupeAndServerErr(err)
	if derr != nil {
		return nil, derr
	}
	return &user, nil
}

// Updates a user's password.
// Only use in protected routes.
func UpdateUserPassword(ctx context.Context, current_user User, unhashed_new_password string) (*User, *result.Error) {
	var user User

	if unhashed_new_password == "" {
		return nil, result.Err(400, nil, "EMPTY_NEW_PASSWORD", "new password must not be empty")
	}

	err := FetchOne(ctx, DB, &user, "UPDATE users SET password = crypt($1, gen_salt('bf')) WHERE uid=$2 RETURNING *",
		unhashed_new_password, current_user.Uid)

	if err != nil {
		return nil, result.ServerErr(err)
	}
	return &user, nil
}
