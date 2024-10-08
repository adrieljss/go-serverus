package db

import (
	"context"
	"strings"
	"time"

	"github.com/adrieljansen/go-serverus/result"
	"github.com/georgysavva/scany/v2/pgxscan"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

// CREATE TABLE IF NOT EXISTS users (
//   uid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
//   user_id VARCHAR(35) UNIQUE NOT NULL,
//   username VARCHAR(50) NOT NULL,
//   email VARCHAR(50) UNIQUE NOT NULL,
//   password TEXT NOT NULL,
//   created_at TIMESTAMP NOT NULL DEFAULT now()
// );

type User struct {
	Uid         string     `json:"uid"`
	UserId      string     `json:"user_id"`
	Username    string     `json:"username"`
	Email       string     `json:"email,omitempty"`
	PfpUrl      *string    `json:"pfp_url"`
	Password    string     `json:"-"`
	CreatedAt   *time.Time `json:"created_at"`
	LastUpdated *time.Time `json:"last_updated"`
}

// this struct is used when want to create users
type RequiredUser struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a RequiredUser) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.UserId, validation.Required, validation.Length(3, 35), is.Alphanumeric),
		validation.Field(&a.Username, validation.Required, validation.Length(2, 35)),
		validation.Field(&a.Email, validation.Required, is.Email),
		validation.Field(&a.Password, validation.Required, validation.Length(5, 30)),
	)
}

// clears out private information - emails and passwords
func (user *User) ClearPrivateInfo() {
	user.Email = ""
	user.Password = ""
}

// check for duplicated users
func HasUserWithSameId(userId string) (bool, *result.Error) {
	var user User
	err := pgxscan.Get(context.Background(), DB, &user, "select * from users where user_id = $1", userId)

	if pgxscan.NotFound(err) {
		return false, nil
	} else if err != nil {
		return false, result.ServerErr(err)
	}
	return true, nil
}

// check for duplicated users
func UserExistsEmail(email string) (bool, *result.Error) {
	var user User
	err := pgxscan.Get(context.Background(), DB, &user, "select * from users where email = $1", email)

	if pgxscan.NotFound(err) {
		return false, nil
	} else if err != nil {
		return false, result.ServerErr(err)
	}
	return true, nil
}

// no cache
func FetchUserByUserId(userId string) (*User, *result.Error) {
	var user User
	err := pgxscan.Get(context.Background(), DB, &user, "select * from users where user_id = $1", userId)

	if pgxscan.NotFound(err) {
		return nil, result.Err(404, err, "USER_NOT_FOUND", "user with the given user id is not found")
	} else if err != nil {
		return nil, result.ServerErr(err)
	}

	return &user, nil
}

func FetchUserByUid(uid string) (*User, *result.Error) {
	var user User
	err := pgxscan.Get(context.Background(), DB, &user, "select * from users where uid = $1", uid)

	if pgxscan.NotFound(err) {
		return nil, result.Err(404, err, "USER_NOT_FOUND", "user with the given uid is not found")
	} else if err != nil {
		return nil, result.ServerErr(err)
	}

	return &user, nil
}

func FetchUserByEmail(email string) (*User, *result.Error) {
	var user User
	err := pgxscan.Get(context.Background(), DB, &user, "select * from users where email = $1", email)

	if pgxscan.NotFound(err) {
		return nil, result.Err(404, err, "USER_NOT_FOUND", "user with the given email is not found")
	} else if err != nil {
		return nil, result.ServerErr(err)
	}

	return &user, nil
}

func FetchUserByUidAndHashedPassword(uid string, hashed_password string) (*User, *result.Error) {
	var user User
	err := pgxscan.Get(context.Background(), DB, &user, "select * from users where uid = $1 AND password = $2", uid, hashed_password)

	if pgxscan.NotFound(err) {
		return nil, result.Err(404, err, "USER_NOT_FOUND", "user with the given uid is not found")
	} else if err != nil {
		return nil, result.ServerErr(err)
	}

	return &user, nil
}

// the
//
//	emailOrUid string
//
// parameter checks if the string has a '@' in it, since Uid doesnt allow any special characters
func FetchUserByCredentials(emailOrUserId string, password string) (*User, *result.Error) {
	var user User
	var err error
	if strings.Contains(emailOrUserId, "@") {
		// email mode
		err = pgxscan.Get(context.Background(), DB, &user, "select * from users where email = $1 and password = crypt($2, password)", emailOrUserId, password)
	} else {
		// uid mode
		err = pgxscan.Get(context.Background(), DB, &user, "select * from users where user_id = $1 and password = crypt($2, password)", emailOrUserId, password)
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

// check for constrainted duplicates (in updates or inserts) and returns correspoding errors or nil
//
// example of constrained duplicates
//
//	`user_id`
//	`email`
func CheckDupeAndServerErr(err error) *result.Error {
	if err != nil {
		if res := IsDuplicateKeyError(err); res != "" {
			if res == "users_user_id_key" {
				return result.ErrWithMetadata(400, err, "USER_ID_TAKEN", "a user with the same user id already exists", map[string]string{
					"user_id": "user id already exists",
				})
			} else {
				// users_email_key
				return result.ErrWithMetadata(400, err, "EMAIL_TAKEN", "a user with the same email already exists", map[string]string{
					"email": "email already exists",
				})
			}
		} else {
			return result.ServerErr(err)
		}
	}
	return nil
}

func CreateUser(userId string, username string, email string, password string) (*User, *result.Error) {
	var user User

	err := pgxscan.Get(context.Background(),
		DB, &user,
		"INSERT INTO users (user_id, username, email, password) VALUES ($1,$2,$3,crypt($4, gen_salt('bf'))) RETURNING *",
		userId, username, email, password)

	derr := CheckDupeAndServerErr(err)
	if derr != nil {
		return nil, derr
	}
	return &user, nil
}

// ! THE FUNCTIONS BELOW DO NOT CHECK IF USER EXISTS BEFOREHAND
// ! SO ALWAYS USE IN PROTECTED ROUTE OR MAY CAUSE UNKNOWN SERVER_ERRORS

// updates a user (no credentials) using UPDATE - SET
//
// if a field is nil then it will use the previous value
//
//	current_user
func UpdateUserInfo(current_user User, user_id string, username string, pfp_url string) (*User, *result.Error) {
	var user User
	// TODO
	if user_id != "" {
		current_user.UserId = user_id
	}

	if username != "" {
		current_user.Username = username
	}

	if pfp_url != "" {
		current_user.PfpUrl = &pfp_url
	}

	err := pgxscan.Get(context.Background(), DB, &user, "UPDATE users SET user_id=$1, username=$2, pfp_url=$3 WHERE uid=$4 RETURNING *",
		current_user.UserId, current_user.Username, current_user.PfpUrl, current_user.Uid)

	derr := CheckDupeAndServerErr(err)
	if derr != nil {
		return nil, derr
	}
	return &user, nil
}

// updates a user password using UPDATE - SET
func UpdateUserPassword(current_user User, unhashed_new_password string) (*User, *result.Error) {
	var user User

	if unhashed_new_password == "" {
		return nil, result.Err(400, nil, "EMPTY_NEW_PASSWORD", "new password must not be empty")
	}

	err := pgxscan.Get(context.Background(), DB, &user, "UPDATE users SET password = crypt($1, gen_salt('bf')) WHERE uid=$2 RETURNING *",
		unhashed_new_password, current_user.Uid)

	if err != nil {
		return nil, result.ServerErr(err)
	}
	return &user, nil
}
