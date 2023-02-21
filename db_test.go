package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func randomString(length int) string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func getPerson(name string) *Person {
	return &Person{
		Name:      name,
		LastName:  "Johnson",
		Birthday:  "01-02-1990",
		BirthDate: time.Date(1989, 2, 1, 0, 0, 0, 0, time.UTC),
	}
}

func getAnimal(name string) *Animal {
	return &Animal{
		Name:    name,
		Type:    "Cat",
		OwnerID: "Johnson",
	}
}

func getDBCred() (string, string) {
	return getEnv("REDIS_HOST", "127.0.0.1:6379"), getEnv("REDIS_PASSWORD", "")
}

func flushDB() error {
	db, err := NewRedisDB(getDBCred())
	if err != nil {
		return err
	}

	if _, err := db.client.FlushAll().Result(); err != nil {
		return err
	}
	return nil
}

func init() {
	if err := flushDB(); err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestNameSet(t *testing.T) {
	person := getPerson("alice")
	err := person.SetName("")
	assert.NoError(t, err)
	err = person.SetName("ab:c:")
	assert.NoError(t, err)
	err = person.SetName("ab::c")
	assert.Error(t, err)
}

func TestStore(t *testing.T) {
	db, err := NewRedisDB(getDBCred())
	assert.NoError(t, err)

	person := getPerson("alice")
	animal := getAnimal("fluffy")

	err = db.Store(context.Background(), person)
	assert.NoError(t, err)

	err = db.Store(context.Background(), animal)
	assert.NoError(t, err)

	person.Name = "test::test"
	err = db.Store(context.Background(), person)
	assert.Error(t, err)

	person.Name = ""
	err = db.Store(context.Background(), person)
	assert.Error(t, err)
}

func TestGetObjectByID(t *testing.T) {
	db, err := NewRedisDB(getDBCred())
	assert.NoError(t, err)

	person := getPerson("alice")
	animal := getAnimal("fluffy")

	err = db.Store(context.Background(), person)
	assert.NoError(t, err)

	err = db.Store(context.Background(), animal)
	assert.NoError(t, err)

	// Happy case
	obj, err := db.GetObjectByID(context.Background(), person.GetID())
	assert.NoError(t, err)
	per, ok := obj.(*Person)
	assert.Equal(t, true, ok)
	assert.Equal(t, person.Name, per.Name)
	assert.Equal(t, person.Birthday, per.Birthday)

	obj, err = db.GetObjectByID(context.Background(), animal.GetID())
	assert.NoError(t, err)
	res, ok := obj.(*Animal)
	assert.Equal(t, true, ok)
	assert.Equal(t, animal.Name, res.Name)
	assert.Equal(t, animal.Type, res.Type)

	// Test with empty id
	_, err = db.GetObjectByID(context.Background(), "")
	assert.Error(t, err)

	// Test with unknown id
	_, err = db.GetObjectByID(context.Background(), "unknown")
	assert.Error(t, err)
}

func TestGetObjectsByName(t *testing.T) {
	db, err := NewRedisDB(getDBCred())
	assert.NoError(t, err)

	pName := randomString(10)
	aName := randomString(10)

	person := getPerson(pName)
	animal := getAnimal(aName)

	var ap = make(map[string]int)

	for i := 0; i < 5; i++ {
		ap[person.Name]++
		ap[animal.Name]++

		err = db.Store(context.Background(), person)
		assert.NoError(t, err)

		err = db.Store(context.Background(), animal)
		assert.NoError(t, err)
	}

	// Happy path
	objs, err := db.GetObjectsByName(context.Background(), pName)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(objs))

	for _, obj := range objs {
		ap[obj.GetName()]--
	}
	assert.Equal(t, 0, ap[pName])

	objs, err = db.GetObjectsByName(context.Background(), aName)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(objs))

	for _, obj := range objs {
		ap[obj.GetName()]--
	}
	assert.Equal(t, 0, ap[aName])

	// Empty name
	_, err = db.GetObjectsByName(context.Background(), "")
	assert.Error(t, err)

	// Unknown name
	objs, err = db.GetObjectsByName(context.Background(), "unknown")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(objs))
}

func TestListObjects(t *testing.T) {
	db, err := NewRedisDB(getDBCred())
	assert.NoError(t, err)

	assert.NoError(t, flushDB())

	pName := randomString(10)
	aName := randomString(10)

	person := getPerson(pName)
	animal := getAnimal(aName)

	var pMap = make(map[string]bool)
	var aMap = make(map[string]bool)

	for i := 0; i < 5; i++ {
		person.Name = pName + fmt.Sprintf("%d", i)
		pMap[person.Name] = true

		animal.Name = pName + fmt.Sprintf("%d", i)
		aMap[animal.Name] = true

		err = db.Store(context.Background(), person)
		assert.NoError(t, err)

		err = db.Store(context.Background(), animal)
		assert.NoError(t, err)
	}

	// Happy path
	objs, err := db.ListObjects(context.Background(), person.GetKind())
	assert.NoError(t, err)
	assert.Equal(t, 5, len(objs))

	for _, obj := range objs {
		delete(pMap, obj.GetName())
	}
	assert.Equal(t, 0, len(pMap))

	objs, err = db.ListObjects(context.Background(), animal.GetKind())
	assert.NoError(t, err)
	assert.Equal(t, 5, len(objs))

	for _, obj := range objs {
		delete(aMap, obj.GetName())
	}
	assert.Equal(t, 0, len(aMap))

	// Unknown name
	objs, err = db.ListObjects(context.Background(), "unknown")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(objs))
}

func TestDeleteObject(t *testing.T) {
	db, err := NewRedisDB(getDBCred())
	assert.NoError(t, err)

	person := getPerson("alice")
	animal := getAnimal("fluffy")

	err = db.Store(context.Background(), person)
	assert.NoError(t, err)

	err = db.Store(context.Background(), animal)
	assert.NoError(t, err)

	// Happy case
	err = db.DeleteObject(context.Background(), person.GetID())
	assert.NoError(t, err)

	// Delete again
	err = db.DeleteObject(context.Background(), person.GetID())
	assert.NoError(t, err)

	// Try to get the object
	_, err = db.GetObjectByID(context.Background(), person.GetID())
	assert.Error(t, err)

	// Delete unknown object
	err = db.DeleteObject(context.Background(), "unknown")
	assert.NoError(t, err)
}
