package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE parcel (
			number INTEGER PRIMARY KEY AUTOINCREMENT,
			client INTEGER,
			status TEXT,
			address TEXT,
			created_at TEXT
		)
	`)
	require.NoError(t, err)

	return db
}

func TestAddGetDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Get
	storedParcel, err := store.Get(id)
	require.NoError(t, err)

	// Сравнение структур, игнорируя поле Number
	expectedParcel := parcel
	expectedParcel.Number = id
	require.Equal(t, expectedParcel, storedParcel)

	// Delete
	err = store.Delete(id)
	require.NoError(t, err)

	// Verify Delete
	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Set Address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// Get and verify address update
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, storedParcel.Address)
}

func TestSetStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Set Status
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// Get and verify status update
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, storedParcel.Status)
}

func TestGetByClient(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	clientID := randRange.Intn(10_000_000)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	parcelMap := make(map[int]Parcel)

	for i := range parcels {
		parcels[i].Client = clientID
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// Get by client
	storedParcels, err := store.GetByClient(clientID)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// Check each parcel
	for _, storedParcel := range storedParcels {
		expectedParcel, exists := parcelMap[storedParcel.Number]
		require.True(t, exists)
		require.Equal(t, expectedParcel, storedParcel)
	}
}
