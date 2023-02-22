package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type Object interface {
	// GetKind returns the type of the object.
	GetKind() string
	// GetID returns a unique UUID for the object.
	GetID() string
	// GetName returns the name of the object. Names are not unique.
	GetName() string
	// SetID sets the ID of the object.
	SetID(string)
	// SetName sets the name of the object.
	SetName(string) error
}

type Person struct {
	Name      string    `json:"name"`
	ID        string    `json:"id"`
	LastName  string    `json:"last_name"`
	Birthday  string    `json:"birthday"`
	BirthDate time.Time `json:"birthdate"`
}

type Animal struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Type    string `json:"type"`
	OwnerID string `json:"owner_id"`
}

func (p *Person) GetKind() string {
	return reflect.TypeOf(p).String()
}

func (p *Person) GetID() string {
	return p.ID
}

func (p *Person) GetName() string {
	return p.Name
}
func (p *Person) SetID(s string) {
	p.ID = s
}
func (p *Person) SetName(s string) error {
	if strings.Contains(s, "::") {
		return errors.New("name contains a restricted special character ::")
	}
	p.Name = s
	return nil
}

func (p *Animal) GetKind() string {
	return reflect.TypeOf(p).String()
}

func (p *Animal) GetID() string {
	return p.ID
}

func (p *Animal) GetName() string {
	return p.Name
}

func (p *Animal) SetID(s string) {
	p.ID = s
}

func (p *Animal) SetName(s string) error {
	if strings.Contains(s, "::") {
		return errors.New("name contains a restricted special character ::")
	}
	p.Name = s
	return nil
}

func getKind(key string) string {
	re := regexp.MustCompile(`^.*::(.*)$`)
	matches := re.FindStringSubmatch(key)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractObject(key, value string) (Object, error) {
	var object Object
	var animal Animal
	var person Person

	kind := getKind(key)
	switch kind {
	case animal.GetKind():
		object = &Animal{}
	case person.GetKind():
		object = &Person{}
	default:
		return nil, errors.New("unknown object type")
	}

	err := json.Unmarshal([]byte(value), object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func main() {
	// Quick test of all the functions

	ctx := context.Background()
	animal := Animal{Name: "tiger", Type: "wild-animal", OwnerID: "alice"}
	person1 := Person{Name: "alice", LastName: "jordon"}
	person2 := Person{Name: "alice", LastName: "macy"}

	db, err := NewRedisDB("127.0.0.1:6379", "")
	if err != nil {
		log.Fatalf(err.Error())
	}

	// We flush here for clean slate
	if _, err := db.client.FlushAll().Result(); err != nil {
		log.Fatalf(err.Error())
	}

	// Store 2 persons and 1 animal
	if err := db.Store(ctx, &animal); err != nil {
		log.Fatalf(err.Error())
	}
	if err := db.Store(ctx, &person1); err != nil {
		log.Fatalf(err.Error())
	}
	if err := db.Store(ctx, &person2); err != nil {
		log.Fatalf(err.Error())
	}

	// Get them by name
	tiger, err := db.GetObjectsByName(ctx, "tiger")
	if err != nil {
		log.Fatalf(err.Error())
	}
	persons, err := db.GetObjectsByName(ctx, "alice")
	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Println("Listing persons:")
	pKind := persons[0].GetKind()
	aKind := tiger[0].GetKind()
	persons, err = db.ListObjects(ctx, pKind)
	if err != nil {
		log.Fatalf(err.Error())
	}
	for i, alice := range persons {
		fmt.Println(i, "Person object: ", alice)
		err = db.DeleteObject(ctx, alice.GetID())
		if err != nil {
			log.Fatalf(err.Error())
		}
		fmt.Println(i, "Deleted person: ", alice)
	}

	tig, err := db.GetObjectByID(ctx, tiger[0].GetID())
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println("Getting animal by id:", tig)
	err = db.DeleteObject(ctx, tig.GetID())
	if err != nil {
		log.Fatalf(err.Error())
	}

	persons, err = db.ListObjects(ctx, pKind)
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println("After deletion, total persons:", len(persons))

	animals, err := db.ListObjects(ctx, aKind)
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println("After deletion, total animals:", len(animals))
}
