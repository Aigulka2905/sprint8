package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

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
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Failed to connect to database")
	defer db.Close()
	err = db.Ping()
	require.NoError(t, err, "Failed to ping database")

	store := NewParcelStore(db)
	service := NewParcelService(store)

	// Добавление посылки
	client := 1000
	address := "Уфа, ул. Менделеева, д. 5"
	parcel, err := service.Register(client, address)
	require.NoError(t, err, "Failed to register parcel")
	require.NotZero(t, parcel.Number, "Parcel number should not be zero")

	// Получение посылки
	retrievedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err, "Failed to get parcel")
	require.Equal(t, parcel, retrievedParcel, "Retrieved parcel does not match added parcel")

	// Удаление посылки
	err = service.Delete(parcel.Number)
	require.NoError(t, err, "Failed to delete parcel")

	// Проверка, что посылка удалена
	_, err = store.Get(parcel.Number)
	require.Error(t, err, "Parcel should not be found after deletion")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Failed to connect to database")
	defer db.Close()
	err = db.Ping()
	require.NoError(t, err, "Failed to ping database")

	store := NewParcelStore(db)
	service := NewParcelService(store)

	// Добавление посылки
	client := 1000
	parcel, err := service.Register(client, "test address")
	require.NoError(t, err, "Failed to register parcel")
	require.NotZero(t, parcel.Number, "Parcel number should not be zero")

	// Обновление адреса
	newAddress := "new test address"
	err = service.ChangeAddress(parcel.Number, newAddress)
	require.NoError(t, err, "Failed to update address")

	// Проверка обновления адреса
	updatedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err, "Failed to get parcel")
	require.Equal(t, newAddress, updatedParcel.Address, "Address was not updated")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Failed to connect to database")
	defer db.Close()
	err = db.Ping()
	require.NoError(t, err, "Failed to ping database")

	store := NewParcelStore(db)
	service := NewParcelService(store)

	// Добавление посылки
	client := 1000
	parcel, err := service.Register(client, "test address")
	require.NoError(t, err, "Failed to register parcel")
	require.NotZero(t, parcel.Number, "Parcel number should not be zero")

	// Обновление статуса
	err = service.NextStatus(parcel.Number)
	require.NoError(t, err, "Failed to update status")

	// Проверка обновления статуса
	updatedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err, "Failed to get parcel")
	require.Equal(t, ParcelStatusSent, updatedParcel.Status, "Status was not updated to sent")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Failed to connect to database")
	defer db.Close()
	err = db.Ping()
	require.NoError(t, err, "Failed to ping database")

	store := NewParcelStore(db)
	service := NewParcelService(store)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// Задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// Добавление посылок
	for i := 0; i < len(parcels); i++ {
		parcel, err := service.Register(client, parcels[i].Address)
		require.NoError(t, err, "Failed to register parcel")
		require.NotZero(t, parcel.Number, "Parcel number should not be zero")

		// Обновляем идентификатор добавленной посылки
		parcels[i].Number = parcel.Number
		parcelMap[parcel.Number] = parcels[i]
	}

	// Получение списка посылок по идентификатору клиента
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "Failed to get parcels by client")
	require.Len(t, storedParcels, len(parcels), "Number of retrieved parcels does not match")

	// Проверка
	for _, parcel := range storedParcels {
		expectedParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists, "Parcel %d not found in parcelMap", parcel.Number)
		require.Equal(t, expectedParcel, parcel, "Parcel data does not match")
	}
}
