package main

import (
	"encoding/json"
	"errors"
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
