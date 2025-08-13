package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

const schema = `
CREATE TABLE parcel (
	number     INTEGER PRIMARY KEY AUTOINCREMENT,
	client     INTEGER,
	status     TEXT,
	address    TEXT,
	created_at TEXT
);`

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(schema)
	require.NoError(t, err)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// get
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, got.Client)
	require.Equal(t, parcel.Status, got.Status)
	require.Equal(t, parcel.Address, got.Address)
	require.Equal(t, parcel.CreatedAt, got.CreatedAt)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)
	_, err = store.Get(id)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(schema)
	require.NoError(t, err)

	store := NewParcelStore(db)

	// add
	id, err := store.Add(getTestParcel())
	require.NoError(t, err)
	require.NotZero(t, id)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, got.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(schema)
	require.NoError(t, err)

	store := NewParcelStore(db)

	// add
	id, err := store.Add(getTestParcel())
	require.NoError(t, err)
	require.NotZero(t, id)

	// set status (registered)
	err = store.SetStatus(id, ParcelStatusRegistered)
	require.NoError(t, err)
	r, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusRegistered, r.Status)

	// set status (sent)
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)
	s, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, s.Status)

	// check
	err = store.SetStatus(id, ParcelStatusDelivered)
	require.NoError(t, err)
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusDelivered, got.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(schema)
	require.NoError(t, err)

	store := NewParcelStore(db)

	// prepare parcels and map
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	parcelMap := make(map[int]Parcel)
	client := randRange.Intn(10_000_000)

	// add
	for i := range parcels {
		parcels[i].Client = client
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Equal(t, len(storedParcels), len(parcels))

	// check
	for _, p := range storedParcels {
		orig, ok := parcelMap[p.Number]
		require.True(t, ok)
		require.Equal(t, orig.Client, p.Client)
		require.Equal(t, orig.Status, p.Status)
		require.Equal(t, orig.Address, p.Address)
		require.Equal(t, orig.CreatedAt, p.CreatedAt)
	}
}
