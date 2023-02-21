package main

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
	_, ok := obj.(*Person)
	assert.Equal(t, true, ok)

	obj, err = db.GetObjectByID(context.Background(), animal.GetID())
	assert.NoError(t, err)
	_, ok = obj.(*Animal)
	assert.Equal(t, true, ok)

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

	for i := 0; i < 5; i++ {
		err = db.Store(context.Background(), person)
		assert.NoError(t, err)

		err = db.Store(context.Background(), animal)
		assert.NoError(t, err)
	}

	// Happy path
	objs, err := db.GetObjectsByName(context.Background(), pName)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(objs))

	objs, err = db.GetObjectsByName(context.Background(), aName)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(objs))

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

	for i := 0; i < 5; i++ {
		err = db.Store(context.Background(), person)
		assert.NoError(t, err)

		err = db.Store(context.Background(), animal)
		assert.NoError(t, err)
	}

	// Happy path
	objs, err := db.ListObjects(context.Background(), person.GetKind())
	assert.NoError(t, err)
	assert.Equal(t, 5, len(objs))

	objs, err = db.ListObjects(context.Background(), animal.GetKind())
	assert.NoError(t, err)
	assert.Equal(t, 5, len(objs))

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
}
