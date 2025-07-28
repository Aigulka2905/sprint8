package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err, "Failed to connect to database")
	defer db.Close()
	err = db.Ping()
	assert.NoError(t, err, "Failed to ping database")

	store := NewParcelStore(db)

	// Добавление посылки
	client := 1000
	address := "Уфа, ул. Менделеева, д. 5"
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	number, err := store.Add(parcel)
	assert.NoError(t, err, "Failed to add parcel")
	parcel.Number = number // Обновляем номер посылки

	// Получение посылки
	retrievedParcel, err := store.Get(parcel.Number)
	assert.NoError(t, err, "Failed to get parcel")
	assert.Equal(t, parcel, retrievedParcel, "Retrieved parcel does not match added parcel")

	// Удаление посылки
	_, err = store.db.Exec("DELETE FROM parcel WHERE number = ?", parcel.Number)
	assert.NoError(t, err, "Failed to delete parcel")

	// Проверка, что посылка удалена
	_, err = store.Get(parcel.Number)
	assert.Error(t, err, "Parcel should not be found after deletion")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err, "Failed to connect to database")
	defer db.Close()
	err = db.Ping()
	assert.NoError(t, err, "Failed to ping database")

	store := NewParcelStore(db)

	// Добавление посылки
	client := 1000
	address := "test address"
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	number, err := store.Add(parcel)
	assert.NoError(t, err, "Failed to add parcel")
	parcel.Number = number // Обновляем номер посылки

	// Обновление адреса
	newAddress := "new test address"
	_, err = store.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", newAddress, parcel.Number)
	assert.NoError(t, err, "Failed to update address")

	// Проверка обновления адреса
	updatedParcel, err := store.Get(parcel.Number)
	assert.NoError(t, err, "Failed to get parcel")
	assert.Equal(t, newAddress, updatedParcel.Address, "Address was not updated")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err, "Failed to connect to database")
	defer db.Close()
	err = db.Ping()
	assert.NoError(t, err, "Failed to ping database")

	store := NewParcelStore(db)

	// Добавление посылки
	client := 1000
	address := "test address"
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	number, err := store.Add(parcel)
	assert.NoError(t, err, "Failed to add parcel")
	parcel.Number = number // Обновляем номер посылки

	// Обновление статуса
	_, err = store.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", ParcelStatusSent, parcel.Number)
	assert.NoError(t, err, "Failed to update status")

	// Проверка обновления статуса
	updatedParcel, err := store.Get(parcel.Number)
	assert.NoError(t, err, "Failed to get parcel")
	assert.Equal(t, ParcelStatusSent, updatedParcel.Status, "Status was not updated to sent")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err, "Failed to connect to database")
	defer db.Close()
	err = db.Ping()
	assert.NoError(t, err, "Failed to ping database")

	store := NewParcelStore(db)

	// Создаём одну тестовую посылку
	parcel := getTestParcel()

	// Формируем срез из трёх посылок на основе одной тестовой посылки
	client := randRange.Intn(10_000_000)
	parcels := []Parcel{parcel, parcel, parcel}
	for i := range parcels {
		parcels[i].Client = client
	}
	parcelMap := map[int]Parcel{}

	// Добавление посылок
	for i := 0; i < len(parcels); i++ {
		number, err := store.Add(parcels[i])
		assert.NoError(t, err, "Failed to add parcel")
		parcels[i].Number = number // Обновляем номер посылки
		parcelMap[number] = parcels[i]
	}

	// Получение списка посылок по идентификатору клиента
	storedParcels, err := store.GetByClient(client)
	assert.NoError(t, err, "Failed to get parcels by client")
	assert.Len(t, storedParcels, len(parcels), "Number of retrieved parcels does not match")

	// Проверка
	for _, parcel := range storedParcels {
		expectedParcel, exists := parcelMap[parcel.Number]
		assert.True(t, exists, "Parcel %d not found in parcelMap", parcel.Number)
		assert.Equal(t, expectedParcel, parcel, "Parcel data does not match")
	}
}
