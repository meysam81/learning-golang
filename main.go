package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"strconv"
)

type contact struct {
	gorm.Model
	ContactName  name   `json:"contact_name"`
	SocialNumber string `json:"social_number"`
	Email        string `json:"email"`
}

type name struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

var contacts []contact

func getContacts(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(contacts); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func getContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr, _ := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	id := uint(idStr)
	for _, contact := range contacts {
		if contact.ID == id {
			if err := json.NewEncoder(w).Encode(contact); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func addContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var contact contact
	if err := json.NewDecoder(r.Body).Decode(&contact); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(contacts) == 0 {
		contact.ID = 1
	} else {
		// the new contact's ID is one more than the last contact
		// TODO: remove this, it will be handled by AUTO_INCREMENT of postgresql
		contact.ID = contacts[len(contacts)-1].ID + 1
	}
	contacts = append(contacts, contact)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(contact)
}

func deleteContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr, _ := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	id := uint(idStr)
	for i, contact := range contacts {
		if contact.ID == id {
			contacts = append(contacts[:i], contacts[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			_ = json.NewEncoder(w).Encode(contacts)
		}
	}
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(contacts)
}

func updateContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr, _ := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	id := uint(idStr)


	var cnt contact
	if err := json.NewDecoder(r.Body).Decode(&cnt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for i, contact := range contacts {
		if contact.ID == id {
			holdContacts := contacts
			cnt.ID = id
			contacts = append(contacts[:i], cnt)
			contacts = append(contacts, holdContacts[i+1:]...)
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(contact)
		}
	}
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(contacts)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/api/contacts", getContacts).Methods("GET")

	r.HandleFunc("/api/contacts", addContact).Methods("POST")
	r.HandleFunc("/api/contacts/{id}", getContact).Methods("GET")
	r.HandleFunc("/api/contacts/{id}", deleteContact).Methods("DELETE")
	r.HandleFunc("/api/contacts/{id}", updateContact).Methods("PUT")
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatalf("Could not start the server: %v", err)
	} else {
		log.Println("Starting the listener...")
	}
	//log.Fatal(http.ListenAndServe(":8000", r))
}
