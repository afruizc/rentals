package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"net/url"
	"rentals"
	"rentals/crypto"
	"strconv"
)

type NotFoundError string

func (e NotFoundError) Error() string {
	return string(e)
}

type UserResource struct {
	Db *gorm.DB
}

func (ur *UserResource) Name() string {
	return "users"
}

func (ur *UserResource) Create(jsonData []byte) ([]byte, error) {
	var newUserSchema struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	err := json.Unmarshal(jsonData, &newUserSchema)
	if err != nil {
		return nil, fmt.Errorf("error calling json.Unmarshall(): %v", err)
	}

	user, err := createUser(newUserSchema.Username, newUserSchema.Password, newUserSchema.Role, ur.Db)
	if err != nil {
		return nil, err
	}

	rawJson, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	return rawJson, nil
}

func (ur *UserResource) Find(query string) ([]byte, error) {
	var users []rentals.User
	ur.Db.Find(&users)

	return json.Marshal(users)
}

func (ur *UserResource) Read(id string) ([]byte, error) {
	user, err := getUserIdStr(id, ur.Db)
	if err != nil {
		return nil, err
	}
	return json.Marshal(user)
}

func (ur *UserResource) Update(id string, jsonData []byte) ([]byte, error) {
	user, err := getUserIdStr(id, ur.Db)
	if err != nil {
		return nil, err
	}

	var updateUserSchema struct {
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	err = json.Unmarshal(jsonData, &updateUserSchema)
	if err != nil {
		return nil, err
	}

	if updateUserSchema.Password != "" {
		user.PasswordHash, err = crypto.EncryptPassword(updateUserSchema.Password)
		if err != nil {
			return nil, fmt.Errorf("[UserResource.Update] error encrypting password %v", err)
		}
	}

	if updateUserSchema.Role != "" {
		if !validRole(updateUserSchema.Role) {
			return nil, errors.New(
				fmt.Sprintf("[UserResource.Update] error creating user. Unknown role %s",
					updateUserSchema.Role))
		}

		user.Role = updateUserSchema.Role
	}

	// Save to DB
	ur.Db.Save(&user)
	return json.Marshal(user)
}

func (ur *UserResource) Delete(id string) error {
	user, err := getUserIdStr(id, ur.Db)
	if err != nil {
		return err
	}

	ur.Db.Delete(&user)
	return nil
}

type ApartmentResource struct {
	Db *gorm.DB
}

func (ar *ApartmentResource) Name() string {
	return "apartments"
}

func (ar *ApartmentResource) Create(jsonData []byte) ([]byte, error) {
	newApartment, err := ar.createApartment(jsonData)
	if err != nil {
		return nil, err
	}

	ar.Db.Create(&newApartment)

	rawJson, err := json.Marshal(newApartment)
	if err != nil {
		return nil, err
	}

	return rawJson, nil
}

func (ar *ApartmentResource) Read(id string) ([]byte, error) {
	apartment, err := getApartment(id, ar.Db)
	if err != nil {
		return nil, err
	}

	return json.Marshal(apartment)
}

func (ar *ApartmentResource) Find(query string) ([]byte, error) {
	values, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	filters := map[string]string{
		"floor_area_meters":   getJsonTag(rentals.Apartment{}, "FloorAreaMeters"),
		"price_per_month_usd": getJsonTag(rentals.Apartment{}, "PricePerMonthUsd"),
		"room_count":          getJsonTag(rentals.Apartment{}, "RoomCount"),
	}

	tx := ar.Db.New()
	tx.Debug()
	for dbField, jsonTag := range filters {
		if v, ok := values[jsonTag]; ok {
			if !ok || len(v) == 0 {
				continue
			}

			tx = tx.Where(fmt.Sprintf("%s = ?", dbField), v[0])
		}
	}

	var apartments []rentals.Apartment
	tx.Find(&apartments)
	return json.Marshal(apartments)
}

func (ar *ApartmentResource) Update(id string, jsonData []byte) ([]byte, error) {
	apartment, err := getApartment(id, ar.Db)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &apartment)
	if err != nil {
		return nil, err
	}

	// Save to DB
	ar.Db.Save(&apartment)

	return json.Marshal(apartment)
}

func (ar *ApartmentResource) Delete(id string) error {
	apartment, err := getApartment(id, ar.Db)
	if err != nil {
		return err
	}

	ar.Db.Delete(&apartment)
	return nil
}

func getApartment(id string, db *gorm.DB) (*rentals.Apartment, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	var apartment rentals.Apartment
	db.First(&apartment, idInt)

	if int(apartment.ID) != idInt {
		return nil, NotFoundError(fmt.Sprintf("apartment %d not found", idInt))
	}

	return &apartment, nil
}

func (ar *ApartmentResource) createApartment(jsonData []byte) (*rentals.Apartment, error) {
	var newApartment rentals.Apartment

	err := json.Unmarshal(jsonData, &newApartment)
	if err != nil {
		return nil, fmt.Errorf("[createApartment] error calling json.Unmarshall(): %v", err)
	}

	realtor := getUser(newApartment.RealtorId, ar.Db)
	if realtor == nil {
		return nil, fmt.Errorf("realtor not found (id=%d)", newApartment.RealtorId)
	}

	err = newApartment.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating apartment: %v", err)
	}

	return &newApartment, nil
}

func getUserIdStr(id string, db *gorm.DB) (*rentals.User, error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	usr := getUser(uint(intId), db)
	if usr == nil {
		return nil, NotFoundError(fmt.Sprintf("user %d not found", intId))
	}

	return usr, nil
}

func getUser(userId uint, db *gorm.DB) *rentals.User {
	var user rentals.User
	db.First(&user, userId)

	if uint(user.ID) != userId {
		return nil
	}

	return &user
}

func createUser(username, password, role string, db *gorm.DB) (*rentals.User, error) {
	if !validRole(role) {
		return nil, errors.New(
			fmt.Sprintf("error creating user. Unknown role %s", role))
	}

	pwdHash, err := crypto.EncryptPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error encrypting password %v", err)
	}

	user := rentals.User{
		Username:     username,
		PasswordHash: pwdHash,
		Role:         role,
	}

	err = db.Create(&user).Error
	if err != nil {
		return nil, fmt.Errorf("error creating user %v", err)
	}

	return &user, nil
}

func validRole(role string) bool {
	return contains([]string{"admin", "realtor", "client"}, role)
}
