/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/spi"
	"golang.org/x/crypto/bcrypt"

	"github.com/jmoiron/sqlx"
	"github.com/kisielk/sqlstruct"
	_ "github.com/mattn/go-sqlite3" // Needed to support sqlite
)

// Match to sqlx
func init() {
	sqlstruct.TagName = "db"
}

const (
	insertUser = `
INSERT INTO users (id, token, type, affiliation, attributes, state, max_enrollments)
	VALUES (:id, :token, :type, :affiliation, :attributes, :state, :max_enrollments);`

	deleteUser = `
DELETE FROM users
	WHERE (id = ?);`

	updateUser = `
UPDATE users
	SET token = :token, type = :type, affiliation = :affiliation, attributes = :attributes, state = :state
	WHERE (id = :id);`

	getUser = `
SELECT * FROM users
	WHERE (id = ?)`

	insertAffiliation = `
INSERT INTO affiliations (name, prekey)
	VALUES (?, ?)`

	deleteAffiliation = `
DELETE FROM affiliations
	WHERE (name = ?)`

	getAffiliation = `
SELECT name, prekey FROM affiliations
	WHERE (name = ?)`
)

// UserRecord defines the properties of a user
type UserRecord struct {
	Name           string `db:"id"`
	Pass           []byte `db:"token"`
	Type           string `db:"type"`
	Affiliation    string `db:"affiliation"`
	Attributes     string `db:"attributes"`
	State          int    `db:"state"`
	MaxEnrollments int    `db:"max_enrollments"`
}

// Accessor implements db.Accessor interface.
type Accessor struct {
	db *sqlx.DB
}

// NewDBAccessor is a constructor for the database API
func NewDBAccessor() *Accessor {
	return &Accessor{}
}

func (d *Accessor) checkDB() error {
	if d.db == nil {
		return errors.New("unknown db object, please check SetDB method")
	}
	return nil
}

// SetDB changes the underlying sql.DB object Accessor is manipulating.
func (d *Accessor) SetDB(db *sqlx.DB) {
	d.db = db
}

// InsertUser inserts user into database
func (d *Accessor) InsertUser(user spi.UserInfo) error {
	log.Debugf("DB: Add identity %s", user.Name)

	err := d.checkDB()
	if err != nil {
		return err
	}

	attrBytes, err := json.Marshal(user.Attributes)
	if err != nil {
		return err
	}

	// Hash the password before storing it
	pwd := []byte(user.Pass)
	pwd, err = bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Failed to hash password: %s", err)
	}

	// Store the user record in the DB
	res, err := d.db.NamedExec(insertUser, &UserRecord{
		Name:           user.Name,
		Pass:           pwd,
		Type:           user.Type,
		Affiliation:    user.Affiliation,
		Attributes:     string(attrBytes),
		State:          user.State,
		MaxEnrollments: user.MaxEnrollments,
	})

	if err != nil {
		log.Errorf("Error adding identity %s to the database: %s", user.Name, err)
		return err
	}

	numRowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if numRowsAffected == 0 {
		return fmt.Errorf("Failed to add identity %s to the database", user.Name)
	}

	if numRowsAffected != 1 {
		return fmt.Errorf("Expected to add one record to the database, but %d records were added", numRowsAffected)
	}

	log.Debugf("Successfully added identity %s to the database", user.Name)

	return nil

}

// DeleteUser deletes user from database
func (d *Accessor) DeleteUser(id string) error {
	log.Debugf("DB: Delete identity %s", id)
	err := d.checkDB()
	if err != nil {
		return err
	}

	_, err = d.db.Exec(deleteUser, id)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser updates user in database
func (d *Accessor) UpdateUser(user spi.UserInfo) error {
	log.Debugf("DB: Update identity %s", user.Name)
	err := d.checkDB()
	if err != nil {
		return err
	}

	attributes, err := json.Marshal(user.Attributes)
	if err != nil {
		return err
	}

	// Hash the password before storing it
	pwd := []byte(user.Pass)
	pwd, err = bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Failed to hash password: %s", err)
	}

	// Store the updated user entry
	res, err := d.db.NamedExec(updateUser, &UserRecord{
		Name:           user.Name,
		Pass:           pwd,
		Type:           user.Type,
		Affiliation:    user.Affiliation,
		Attributes:     string(attributes),
		State:          user.State,
		MaxEnrollments: user.MaxEnrollments,
	})

	if err != nil {
		log.Errorf("Failed to update identity record [error: %s]", err)
		return err
	}

	numRowsAffected, err := res.RowsAffected()

	if numRowsAffected == 0 {
		return fmt.Errorf("Failed to update the identity record")
	}

	if numRowsAffected != 1 {
		return fmt.Errorf("Expected one identity record to be updated, but %d records were updated", numRowsAffected)
	}

	return err

}

// GetUser gets user from database
func (d *Accessor) GetUser(id string, attrs []string) (spi.User, error) {
	log.Debugf("DB: Getting identity %s", id)

	err := d.checkDB()
	if err != nil {
		return nil, err
	}

	var userRec UserRecord
	err = d.db.Get(&userRec, d.db.Rebind(getUser), id)
	if err != nil {
		return nil, err
	}

	return d.newDBUser(&userRec), nil
}

// GetUserInfo gets user information from database
func (d *Accessor) GetUserInfo(id string) (spi.UserInfo, error) {
	log.Debugf("DB: Getting information for identity %s", id)

	var userInfo spi.UserInfo

	err := d.checkDB()
	if err != nil {
		return userInfo, err
	}

	var userRec UserRecord
	err = d.db.Get(&userRec, d.db.Rebind(getUser), id)
	if err != nil {
		return userInfo, err
	}

	var attributes []api.Attribute
	json.Unmarshal([]byte(userRec.Attributes), &attributes)

	userInfo.Name = userRec.Name
	userInfo.Type = userRec.Type
	userInfo.Affiliation = userRec.Affiliation
	userInfo.State = userRec.State
	userInfo.MaxEnrollments = userRec.MaxEnrollments
	userInfo.Attributes = attributes

	return userInfo, nil
}

// InsertAffiliation inserts affiliation into database
func (d *Accessor) InsertAffiliation(name string, prekey string) error {
	log.Debugf("DB: Add affiliation %s", name)
	err := d.checkDB()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(d.db.Rebind(insertAffiliation), name, prekey)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAffiliation deletes affiliation from database
func (d *Accessor) DeleteAffiliation(name string) error {
	log.Debugf("DB: Delete affiliation %s", name)
	err := d.checkDB()
	if err != nil {
		return err
	}

	_, err = d.db.Exec(deleteAffiliation, name)
	if err != nil {
		return err
	}

	return nil
}

// GetAffiliation gets affiliation from database
func (d *Accessor) GetAffiliation(name string) (spi.Affiliation, error) {
	log.Debugf("DB: Get affiliation %s", name)
	err := d.checkDB()
	if err != nil {
		return nil, err
	}

	var affiliation spi.AffiliationImpl

	err = d.db.Get(&affiliation, d.db.Rebind(getAffiliation), name)
	if err != nil {
		return nil, err
	}

	return &affiliation, nil
}

// Creates a DBUser object from the DB user record
func (d *Accessor) newDBUser(userRec *UserRecord) *DBUser {
	var user = new(DBUser)
	user.Name = userRec.Name
	user.pass = userRec.Pass
	user.State = userRec.State
	user.MaxEnrollments = userRec.MaxEnrollments
	user.Affiliation = userRec.Affiliation
	user.Type = userRec.Type

	var attrs []api.Attribute
	json.Unmarshal([]byte(userRec.Attributes), &attrs)
	user.Attributes = attrs

	user.attrs = make(map[string]string)
	for _, attr := range attrs {
		user.attrs[attr.Name] = attr.Value
	}

	user.db = d.db
	return user
}

// DBUser is the databases representation of a user
type DBUser struct {
	spi.UserInfo
	pass  []byte
	attrs map[string]string
	db    *sqlx.DB
}

// GetName returns the enrollment ID of the user
func (u *DBUser) GetName() string {
	return u.Name
}

// Login the user with a password
func (u *DBUser) Login(pass string, caMaxEnrollments int) error {
	var stateUpdateSQL string
	var args []interface{}

	log.Debugf("DB: Login user %s with max enrollments of %d and state of %d", u.Name, u.MaxEnrollments, u.State)

	// Check the password by comparing to stored hash
	err := bcrypt.CompareHashAndPassword(u.pass, []byte(pass))
	if err != nil {
		return fmt.Errorf("Password mismatch: %s", err)
	}

	if u.MaxEnrollments == 0 {
		return fmt.Errorf("Zero is an invalid value for maximum enrollments on identity '%s'", u.Name)
	}

	if u.State == -1 {
		return fmt.Errorf("User %s is revoked; access denied", u.Name)
	}

	// If max enrollment value of user is greater than allowed by CA, using CA max enrollment value for user
	if u.MaxEnrollments > caMaxEnrollments || (u.MaxEnrollments == -1 && caMaxEnrollments != -1) {
		log.Debugf("Max enrollment value (%d) of identity is greater than allowed by CA, using CA max enrollment value of %d", u.MaxEnrollments, caMaxEnrollments)
		u.MaxEnrollments = caMaxEnrollments
	}

	// If maxEnrollments is set to -1, user has unlimited enrollment
	// If the maxEnrollments is set (i.e. >= 1), make sure we haven't exceeded this number of logins.
	// The state variable keeps track of the number of previously successful logins.
	if u.MaxEnrollments != -1 && u.State >= u.MaxEnrollments {
		return fmt.Errorf("The identity %s has already enrolled %d times, it has reached its maximum enrollment allowance", u.Name, u.MaxEnrollments)
	}

	// Not exceeded, so attempt to increment the count
	state := u.State + 1
	args = append(args, u.Name)
	if u.MaxEnrollments == -1 {
		// unlimited so no state check
		stateUpdateSQL = "UPDATE users SET state = state + 1 WHERE (id = ?)"
	} else {
		// state must be less than max enrollments
		stateUpdateSQL = "UPDATE users SET state = state + 1 WHERE (id = ? AND state < ?)"
		args = append(args, u.MaxEnrollments)
	}
	res, err := u.db.Exec(u.db.Rebind(stateUpdateSQL), args...)
	if err != nil {
		return fmt.Errorf("Failed to update state of identity %s to %d: %s", u.Name, state, err)
	}

	numRowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db.RowsAffected failed: %s", err)
	}

	if numRowsAffected == 0 {
		return fmt.Errorf("No rows were affected when updating the state of identity %s", u.Name)
	}

	if numRowsAffected != 1 {
		return fmt.Errorf("%d rows were affected when updating the state of identity %s", numRowsAffected, u.Name)
	}

	log.Debugf("Successfully incremented state for identity %s to %d", u.Name, state)

	log.Debugf("DB: identity %s successfully logged in", u.Name)

	return nil

}

// GetAffiliationPath returns the complete path for the user's affiliation.
func (u *DBUser) GetAffiliationPath() []string {
	affiliationPath := strings.Split(u.Affiliation, ".")
	return affiliationPath
}

// GetAttribute returns the value for an attribute name
func (u *DBUser) GetAttribute(name string) string {
	return u.attrs[name]
}
