package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"testing"
	"tournaments"
	"tournaments/tst"
)

type apartmentResponse struct {
	ID               uint    `json:"id"`
	Name             string  `json:"name"`
	Desc             string  `json:"description"`
	RealtorId        uint    `json:"realtorId"`
	FloorAreaMeters  float32 `json:"floorAreaMeters"`
	PricePerMonthUsd float32 `json:"pricePerMonthUSD"`
	RoomCount        int     `json:"roomCount"`
	Latitude         float32 `json:"latitude"`
	Longitude        float32 `json:"longitude"`
	Available        bool    `json:"available"`
}

func TestCRUDApartment(t *testing.T) {
	var wg sync.WaitGroup
	const addr = "localhost:8083"
	app, err := rentals.NewApp(addr)
	tst.Ok(t, err)
	tst.Ok(t, app.Setup())

	// Make sure we delete all things after we are done
	defer app.DropDB()
	serverUrl := fmt.Sprintf("http://%s", addr)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("[ERROR] %s", app.ServeHTTP())
	}()

	_, err = createUser("admin", "admin", "admin", app.Server.Db)
	tst.Ok(t, err)
	realtorId, err := createUser("realtor", "realtor", "realtor", app.Server.Db)
	tst.Ok(t, err)
	_, err = createUser("client", "client", "client", app.Server.Db)
	tst.Ok(t, err)

	t.Run("CRUD apartment no auth, fail", func(t *testing.T) {
		res, err := tst.MakeRequest("POST", serverUrl+"/apartments", "", []byte(""))
		tst.Ok(t, err)

		tst.Assert(t, res.StatusCode == http.StatusUnauthorized,
			fmt.Sprintf("Expected 401, got %d", res.StatusCode))

		for _, url := range []string{"/apartments", "/apartments/2"} {
			res, err := tst.MakeRequest("GET", serverUrl+url, "", []byte(""))
			tst.Ok(t, err)

			tst.Assert(t, res.StatusCode == http.StatusUnauthorized,
				fmt.Sprintf("Expected 401, got %d", res.StatusCode))
		}

		res, err = tst.MakeRequest("PATCH", serverUrl+"/apartments/1", "", []byte(""))
		tst.Ok(t, err)

		tst.Assert(t, res.StatusCode == http.StatusUnauthorized,
			fmt.Sprintf("Expected 401, got %d", res.StatusCode))

		res, err = tst.MakeRequest("DELETE", serverUrl+"/apartments/1", "", []byte(""))
		tst.Ok(t, err)

		tst.Assert(t, res.StatusCode == http.StatusUnauthorized,
			fmt.Sprintf("Expected 401, got %d", res.StatusCode))
	})

	t.Run("Create Update Delete apartment with client, fail", func(t *testing.T) {
		token, err := loginWithUser(t, serverUrl, "client", "client")
		tst.Ok(t, err)
		newApartmentPayload := newApartmentPayload("apt1", "desc", realtorId)

		res, err := tst.MakeRequest("POST", serverUrl+"/apartments", token, newApartmentPayload)
		tst.Ok(t, err)

		tst.Assert(t, res.StatusCode == http.StatusForbidden,
			fmt.Sprintf("Expected 403 got %d", res.StatusCode))

		res, err = tst.MakeRequest("PATCH", serverUrl+"/apartments/1", token, newApartmentPayload)
		tst.Ok(t, err)

		tst.Assert(t, res.StatusCode == http.StatusForbidden,
			fmt.Sprintf("Expected 403 got %d", res.StatusCode))

		res, err = tst.MakeRequest("DELETE", serverUrl+"/apartments/1", token, newApartmentPayload)
		tst.Ok(t, err)

		tst.Assert(t, res.StatusCode == http.StatusForbidden,
			fmt.Sprintf("Expected 403 got %d", res.StatusCode))
	})

	t.Run("CRUD apartment realtor admin, success", func(t *testing.T) {
		for _, user := range []string{"admin", "realtor"} {
			// Get user token
			token, err := loginWithUser(t, serverUrl, user, user)
			tst.Ok(t, err)

			// Create
			payload := newApartmentPayload("apt1", "desc", realtorId)
			res, err := tst.MakeRequest("POST", serverUrl+"/apartments", token, payload)
			tst.Ok(t, err)

			tst.Assert(t, res.StatusCode == http.StatusCreated, fmt.Sprintf("Expected 201 got %d", res.StatusCode))
			rawContent, err := ioutil.ReadAll(res.Body)
			tst.Ok(t, err)

			var aptRes apartmentResponse
			err = json.Unmarshal(rawContent, &aptRes)

			tst.Assert(t, aptRes.ID >= 1, "Expected id greater than 0")
			tst.Assert(t, aptRes.Name == "apt1", "Got name different name")
			tst.Assert(t, aptRes.RealtorId == realtorId, "Got unexpected realtor")
			tst.Assert(t, aptRes.Available, "Expected apartment to be available")

			// Read
			apartmentUrl := fmt.Sprintf("%s/apartments/%d", serverUrl, aptRes.ID)
			res, err = tst.MakeRequest("GET", apartmentUrl, token, []byte(""))
			tst.Ok(t, err)

			tst.Assert(t, res.StatusCode == http.StatusOK,
				fmt.Sprintf("Expected 200, got %d", res.StatusCode))

			var retApt apartmentResponse
			decoder := json.NewDecoder(res.Body)
			err = decoder.Decode(&retApt)
			tst.Ok(t, err)

			tst.Assert(t, retApt.ID == aptRes.ID, fmt.Sprintf("Expected id 1, got %d", retApt.ID))

			// Update
			newData := []byte(`{"id": 100, "name": "newName", "description": "newDesc"}`)
			res, err = tst.MakeRequest("PATCH", apartmentUrl, token, newData)
			tst.Ok(t, err)

			tst.Assert(t, res.StatusCode == http.StatusOK,
				fmt.Sprintf("Expected 200, got %d", res.StatusCode))

			var updApt apartmentResponse
			decoder = json.NewDecoder(res.Body)
			err = decoder.Decode(&updApt)
			tst.Ok(t, err)
			tst.Assert(t, updApt.ID == retApt.ID,
				fmt.Sprintf("Expected id to be %d, got %d", updApt.ID, retApt.ID))
			tst.Assert(t, updApt.Name == "newName",
				fmt.Sprintf("Expected name to be newName, got %s", updApt.Name))
			tst.Assert(t, updApt.Desc == "newDesc",
				fmt.Sprintf("Expected name to be newDesc, got %s", updApt.Desc))
			tst.Assert(t, updApt.FloorAreaMeters == retApt.FloorAreaMeters,
				fmt.Sprintf("Expected floorArea to be %f, got %f",
					retApt.FloorAreaMeters, updApt.FloorAreaMeters))
			tst.Assert(t, updApt.PricePerMonthUsd == retApt.PricePerMonthUsd,
				fmt.Sprintf("Expected pricePM to be %f, got %f",
					retApt.PricePerMonthUsd, updApt.PricePerMonthUsd))

			// Delete
			res, err = tst.MakeRequest("DELETE", apartmentUrl, token, []byte(""))
			tst.Ok(t, err)
			tst.Assert(t, res.StatusCode == http.StatusNoContent,
				fmt.Sprintf("Expected 204, got %d", res.StatusCode))
		}
	})
}

func TestReadAllApartments(t *testing.T) {
	var wg sync.WaitGroup
	const addr = "localhost:8083"
	app, err := rentals.NewApp(addr)
	tst.Ok(t, err)
	tst.Ok(t, app.Setup())

	// Make sure we delete all things after we are done
	defer app.DropDB()
	serverUrl := fmt.Sprintf("http://%s", addr)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("[ERROR] %s", app.ServeHTTP())
	}()

	_, err = createUser("admin", "admin", "admin", app.Server.Db)
	tst.Ok(t, err)
	realtorId, err := createUser("realtor", "realtor", "realtor", app.Server.Db)
	tst.Ok(t, err)
	_, err = createUser("client", "client", "client", app.Server.Db)
	tst.Ok(t, err)

	create10Apartments(t, realtorId, app.Server.Db)

	t.Run("Read all apartments client, realtor, admin, success", func(t *testing.T) {
		for _, user := range []string{"client", "realtor", "admin"} {
			token, err := loginWithUser(t, serverUrl, user, user)
			tst.Ok(t, err)

			// Act
			res, err := tst.MakeRequest("GET", serverUrl+"/apartments", token, []byte(""))
			tst.Ok(t, err)

			// Assert
			tst.Assert(t, res.StatusCode == http.StatusOK,
				fmt.Sprintf("Expected 200, got %d", res.StatusCode))

			var returnedApartments []apartmentResponse
			decoder := json.NewDecoder(res.Body)
			err = decoder.Decode(&returnedApartments)
			tst.Ok(t, err)

			tst.Assert(t, len(returnedApartments) == 10,
				fmt.Sprintf("Expected 10 apartments, got %d", len(returnedApartments)))
		}
	})
}

func newApartmentPayload(name, desc string, realtorId uint) []byte {
	return []byte(fmt.Sprintf(
		`{
"name":"%s",
"description": "%s",
"floorAreaMeters": 50.0,
"pricePerMonthUSD": 500.0,
"roomCount": 4,
"latitude": 41.761536,
"longitude": 12.315237,
"available": true,
"realtorId": %d}`, name, desc, realtorId))
}

func create10Apartments(t *testing.T, realtorId uint, db *gorm.DB) {
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("apt%d", i)
		desc := fmt.Sprintf("desc%d", i)
		_, err := createApartment(name, desc, realtorId, db)
		tst.Ok(t, err)
	}
}

// Creates a user. Returns its id.
func createApartment(name, desc string, realtorId uint, db *gorm.DB) (uint, error) {
	apartmentResource := &rentals.ApartmentResource{Db: db}

	apartmentData := newApartmentPayload(name, desc, realtorId)
	jsonData, err := apartmentResource.Create([]byte(apartmentData))
	if err != nil {
		return 0, err
	}

	var apartmentId struct {
		Id uint `json:"id"`
	}

	err = json.Unmarshal(jsonData, &apartmentId)
	if err != nil {
		return 0, err
	}

	return apartmentId.Id, nil
}
