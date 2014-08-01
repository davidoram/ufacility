// @APIVersion 1.0.0
// @Title Reward API
// @Description Rewards as resources
// @Contact api@contact.me
// @TermsOfServiceUrl http://google.com
// @License TODO
// @LicenseUrl http://google.com

// @SubApi Service alive API [/ping]
// @SubApi Panic test API [/panic]
package main

import (
	"flag"
	"fmt"
	"github.com/davidoram/ufacility/database"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
)

// This example show a complete (GET,PUT,POST,DELETE) conventional example of
// a REST Resource including documentation to be served by e.g. a Swagger UI
// It is recommended to create a Resource struct (UserResource) that can encapsulate
// an object that provide domain access (a DAO)
// It has a Register method including the complete Route mapping to methods together
// with all the appropriate documentation
//
// POST http://localhost:8080/users
// <User><Id>1</Id><Name>Melissa Raspberry</Name></User>
//
// GET http://localhost:8080/users/1
//
// PUT http://localhost:8080/users/1
// <User><Id>1</Id><Name>Melissa</Name></User>
//
// DELETE http://localhost:8080/users/1
//

type User struct {
	Id, Name string
}

type UserResource struct {
	// normally one would use DAO (data access object)
	users map[string]User
}

func (u UserResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/users").
		Doc("Manage Users").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	ws.Route(ws.GET("/{user-id}").To(u.findUser).
		// docs
		Doc("get a user").
		Operation("findUser").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
		Writes(User{})) // on the response

	ws.Route(ws.PUT("/{user-id}").To(u.updateUser).
		// docs
		Doc("update a user").
		Operation("updateUser").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
		Reads(User{})) // from the request

	ws.Route(ws.POST("").To(u.createUser).
		// docs
		Doc("create a user").
		Operation("createUser").
		Reads(User{})) // from the request

	ws.Route(ws.DELETE("/{user-id}").To(u.removeUser).
		// docs
		Doc("delete a user").
		Operation("removeUser").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")))

	container.Add(ws)
}

// GET http://localhost:8080/users/1
//
func (u UserResource) findUser(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	usr := u.users[id]
	if len(usr.Id) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "User could not be found.")
	}
	response.WriteEntity(usr)
}

// POST http://localhost:8080/users
// <User><Name>Melissa</Name></User>
//
func (u *UserResource) createUser(request *restful.Request, response *restful.Response) {
	usr := new(User)
	err := request.ReadEntity(usr)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	usr.Id = strconv.Itoa(len(u.users) + 1) // simple id generation
	u.users[usr.Id] = *usr
	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(usr)
}

// PUT http://localhost:8080/users/1
// <User><Id>1</Id><Name>Melissa Raspberry</Name></User>
//
func (u *UserResource) updateUser(request *restful.Request, response *restful.Response) {
	usr := new(User)
	err := request.ReadEntity(&usr)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	u.users[usr.Id] = *usr
	response.WriteEntity(usr)
}

// DELETE http://localhost:8080/users/1
//
func (u *UserResource) removeUser(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	delete(u.users, id)
}

func main() {
	var (
		dbHost     = flag.String("host", "localhost", "Postgress db host")
		dbName     = flag.String("db", "rewards_development", "Postgress db name")
		dbUser     = flag.String("user", os.Getenv("USER"), "Postgress db user")
		dbPassword = flag.String("password", "", "Postgress db password")
		httpPort   = flag.Int("port", 8080, "HTTP port to listen on")
		publicPath = flag.String("public", "", "Path to 'public' directory files, contains swagger API")
	)
	flag.Parse()

	log.Println("ufacility started")

	log.Printf("Connecting to database: %v, host: %v, user: %v ...", *dbName, *dbHost, *dbUser)
	connectionString := fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", *dbUser, *dbPassword, *dbHost, *dbName)
	db := sqlx.MustConnect("postgres", connectionString)

	/*
	 * Put your database migrations here, will be processed in order
	 */
	migrations := []database.Migration{
		database.Migration{
			`CREATE TABLE rewards(
				id 						integer,
				permalink 		varchar(255) not null,
				name					varchar(255) not null,

				PRIMARY KEY(id)
			)
			`,
		},
	}
	database.MigrateDatabase(db, &migrations)

	wsContainer := restful.NewContainer()
	u := UserResource{map[string]User{}}
	u.Register(wsContainer)

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	config := swagger.Config{
		WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: fmt.Sprintf("http://localhost:%d", *httpPort),
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: fmt.Sprintf("%v/swagger", *publicPath)}
	swagger.RegisterSwaggerService(config, wsContainer)

	log.Printf("start listening on localhost:%d", *httpPort)
	server := &http.Server{Addr: fmt.Sprintf(":%d", *httpPort), Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
