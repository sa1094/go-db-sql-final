package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func prepareStore() (ParcelStore, error) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return ParcelStore{}, err
	}
	store := NewParcelStore(db)
	err = cleanUp(store)
	return store, err
}

func cleanUp(p ParcelStore) error {
	_, err := p.db.Exec("delete from sqlite_sequence WHERE name='parcel'")
	if err != nil {
		return err
	}
	_, err = p.db.Exec("delete from 'parcel'")
	return err
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	store, err := prepareStore()
	require.NoError(t, err)
	defer cleanUp(store)

	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)
	parcel.Number = id

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	actualParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, parcel, actualParcel)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	require.NoError(t, store.Delete(parcel.Number))
	_, err = store.Get(parcel.Number)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	store, err := prepareStore()
	require.NoError(t, err)
	defer cleanUp(store)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)
	parcel.Number = id

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	actualParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, newAddress, actualParcel.Address)
}

func TestCannotSetAddress(t *testing.T) {
	// prepare
	store, err := prepareStore()
	require.NoError(t, err)
	defer cleanUp(store)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)
	parcel.Number = id

	statuses := []string{ParcelStatusSent, ParcelStatusDelivered}
	newAddress := "new test address"
	for _, s := range statuses {
		parcel.Status = s
		err = store.SetStatus(parcel.Number, parcel.Status)
		require.NoError(t, err)

		err = store.SetAddress(parcel.Number, newAddress)
		require.Error(t, err)

		actualParcel, err := store.Get(parcel.Number)
		require.NoError(t, err)
		require.Equal(t, parcel.Address, actualParcel.Address)
	}
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	store, err := prepareStore()
	require.NoError(t, err)
	defer cleanUp(store)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)
	parcel.Number = id

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	parcel.Status = ParcelStatusSent
	err = store.SetStatus(parcel.Number, parcel.Status)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	actualParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, parcel, actualParcel)
}

// TestNoDeleteStatus проверяет неудаление только статусов кроме "registered"
func TestNoDeleteStatus(t *testing.T) {
	// prepare
	store, err := prepareStore()
	require.NoError(t, err)
	defer cleanUp(store)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)
	parcel.Number = id

	// set status
	statuses := []string{ParcelStatusSent, ParcelStatusDelivered}
	for _, s := range statuses {
		parcel.Status = s
		err = store.SetStatus(parcel.Number, parcel.Status)
		require.NoError(t, err)
		err = store.Delete(parcel.Number)
		require.NoError(t, err)
		actualParcel, err := store.Get(parcel.Number)
		require.NoError(t, err)
		assert.Equal(t, parcel, actualParcel)
	}

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	store, err := prepareStore()
	require.NoError(t, err)
	defer cleanUp(store)
	setVolume := 3
	client := randRange.Intn(10_000_000)

	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	// client := randRange.Intn(10_000_000)

	// add
	for i := 0; i < setVolume; i++ {
		parcel := getTestParcel()
		parcel.Client = client
		id, err := store.Add(parcel)
		require.NoError(t, err)
		require.Greater(t, id, 0)
		// обновляем идентификатор добавленной у посылки
		parcel.Number = id
		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcel
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err)
	assert.Len(t, storedParcels, setVolume)
	// check
	for _, parcel := range storedParcels {
		actualParcel, err := store.Get(parcel.Number)
		require.NoError(t, err)
		assert.Equal(t, parcel, actualParcel)
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}
