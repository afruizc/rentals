# Rentals

Manage rental properties

## Usage

The application is written with a Go backend and a Javascript (vue.js) frontend.
Make sure `baseUrl` is set correctly in `frontend/src/components/http.js`, otherwise
requests to the server will fail.

The server takes the database host, name and user from env variable, specifically:

```
RENTALS_DB_HOST
RENTALS_DB_NAME
RENTALS_DB_USER
```

See `scripts/run.sh` for an example on how to start the server. `scripts/rentals-cli`
is a compiled binary that can be used directly to run the server. Otherwise, you can install
go > 1.11 (New modules are used) and build all the project. To obtain a binary, cd into
`cmd/rentals-cli` and run `go build ./...`.

Scripts in `scripts` DO NOT WORK if you are outside the directory. Make sure you cd into
it.

To run all tests, run `scripts/all_tests.sh`. As previously stated, make sure you
are inside the `scripts` directory.

The frontend is run using `vue-cli`. `cd` into the `frontend` folder, install all dependencies
with `npm install --saveDev` and then run `npm run serve`. This will bring up a dev server
for the frontend.

## Spec

*Write an application that manages apartment rentals using a REST API*

* Users must be able to create an account and log in.

* Implement `client`, `realtor` and `admin` role:
   * Clients: browse rentable apartments in a list and on a map.
   * Realtors: client + CRUD all apartments and set the apartment state to available/rented.
   * Admins: realtor +  CRUD realtors, and clients.
   
* Apartments have:
    * Name.
    * Description.
    * Floor area size.
    * Price per month.
    * Number of rooms.
    * Valid geolocation coordinates (either lat/log or geocode).
    * Date added.
    * Associated realtor.

* Apartments are searchable by:
    * Floor area size.
    * Price per month.
    * Number of rooms.
 
- Single-page application. All actions need to be done client side using AJAX,
refreshing the page is not acceptable. Functional UI/UX design is needed. You are
not required to create a unique design, however, do follow best practices to make
the project as functional as possible.

- Bonus: unit and e2e tests.

## Attack plan:

- ~~User creation.~~
- ~~Authentication for user creation.~~
- ~~Apartment creation.~~
- ~~Add authorization to user and apartment creation.~~
- ~~Add read/update/delete tournaments.~~
- ~~Add read/update/delete users.~~
- ~~Search by floor area size, price, rooms.~~
- ~~Write frontend.~~
- ~~Create endpoint to return user info.~~
- ~~Frontend: Only show create to admins/realtors.~~
- ~~Change available state.~~
- ~~Make username unique in database.~~
- ~~Fix frontend.~~
- ~~Create client account endpoint.~~
- ~~Date added for apartments.~~
- ~~Drop down/Select self for realtor.~~
- ~~Validate apartment info.~~
- Bonus: create users with admin.
- Bonus: Do geocoding in the frontend.
