package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"net/http"
	"os"
	"strconv"
)

var db *gorm.DB

type configuration struct {
	Dialect  string `json:"dialect"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Dbname   string `json:"dbname"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type contact struct {
	ID           uint
	Name         name `json:"contact_name" gorm:"foreignkey:NameID,association_foreignkey:ID"`
	NameID       uint
	SocialNumber string `json:"social_number"`
	Email        string `json:"email"`
}

type name struct {
	ID        uint   `sql:"index"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/api/contacts", http.StatusPermanentRedirect)
}

func getContacts(w http.ResponseWriter, _ *http.Request) {
	// Set HTTP Header
	w.Header().Set("Content-Type", "application/json")

	// Fetch from PostgreSQL (using GORM)
	var contacts []contact
	db.Debug().Find(&contacts)

	// To populate the foreign key (GORM doesn't do it automatically)
	db.Debug().Preload("Name").Find(&contacts)

	// Return response
	w.WriteHeader(http.StatusFound)
	_ = json.NewEncoder(w).Encode(contacts)

	// A final log for debugging
	log.Printf("All contacts: %v", &contacts)
}

func getContact(w http.ResponseWriter, r *http.Request) {
	// Set HTTP Header
	w.Header().Set("Content-Type", "application/json")

	// Get id from url params
	idStr, _ := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	id := uint(idStr)

	// Fetch from PostgreSQL (using GORM)
	var c contact
	db.Debug().First(&c, id)

	// To populate the foreign key (GORM doesn't do it automatically)
	db.Debug().Preload("Name").Find(&c)

	// Return response
	w.WriteHeader(http.StatusFound)
	_ = json.NewEncoder(w).Encode(c)

	// A final log for debugging
	log.Printf("Get contact %v: %v", id, &c)
}

func addContact(w http.ResponseWriter, r *http.Request) {
	// Set HTTP Header
	w.Header().Set("Content-Type", "application/json")

	// Get info from HTTP body (in JSON format)
	var c contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Insert to PostgreSQL (using GORM)
	db.Debug().Create(&c)

	// Return response
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)

	// A final log for debugging
	log.Printf("Add contact %v: %v", &c.ID, &c)
}

func deleteContact(w http.ResponseWriter, r *http.Request) {
	// Set HTTP Header
	w.Header().Set("Content-Type", "application/json")

	// Get id from url params
	idStr, _ := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	id := uint(idStr)

	// Delete from PostgreSQL (using GORM) ---> Hard delete
	db.Debug().Delete(&contact{}, "id = ?", id)

	// Return response
	w.WriteHeader(http.StatusNoContent)

	// A final log for debugging
	log.Printf("Delete contact: %v", id)
}

func updateContact(w http.ResponseWriter, r *http.Request) {
	// Set HTTP Header
	w.Header().Set("Content-Type", "application/json")

	// Get id from url params
	idStr, _ := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	id := uint(idStr)

	// Get info from HTTP body (in JSON format)
	var c contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Update field in PostgreSQL (using GORM)
	var hold contact
	db.Debug().Where("id = ?", id).First(&hold)
	hold.Name.FirstName = c.Name.FirstName
	hold.Name.LastName = c.Name.LastName
	hold.Email = c.Email
	hold.SocialNumber = c.SocialNumber
	db.Save(&hold)

	// Return response
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(&hold)

	// A final log for debugging
	log.Printf("Update contact %v: %v", &hold.ID, &hold)
}

func readConfig(fileName string) (*configuration, error) {
	if fileName == "" {
		fileName = "conf.json"
	}
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	config := &configuration{}
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

func createTables() error {
	for _, model := range []interface{}{&contact{}, &name{}} {
		db.Debug().DropTableIfExists(model)
		db.Debug().CreateTable(model)
	}
	db.Debug().Model(&contact{}).AddForeignKey("name_id", "names(id)", "CASCADE", "CASCADE")
	return nil
}

func init() {
	config, err := readConfig("")
	if err != nil {
		panic(err)
	}

	var args string
	if config.Host != "" {
		args += fmt.Sprintf("host=%v ", config.Host)
	}
	if config.Port != 0 {
		args += fmt.Sprintf("port=%v ", config.Port)
	}
	args += fmt.Sprintf("dbname=%v user=%v password=%v", config.Dbname, config.User, config.Password)
	db, err = gorm.Open(config.Dialect, args)
	if err != nil {
		panic(err)
	}

	if err = db.Debug().DB().Ping(); err != nil {
		panic(err)
	}

	err = createTables()
	if err != nil {
		panic(err)
	}
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", index).Methods("GET")

	r.HandleFunc("/api/contacts", getContacts).Methods("GET")
	r.HandleFunc("/api/contacts", addContact).Methods("POST")
	r.HandleFunc("/api/contacts/{id}", getContact).Methods("GET")
	r.HandleFunc("/api/contacts/{id}", deleteContact).Methods("DELETE")
	r.HandleFunc("/api/contacts/{id}", updateContact).Methods("PUT")

	log.Println("Starting the server...")
	log.Println(http.ListenAndServe(":8000", r))
}
