package main

import (
	"database/sql"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	sqlExpr := "INSERT INTO parcel (client ,status, address, created_at) values (:client, :status, :address, :created_at)"

	r, err := s.db.Exec(sqlExpr,
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}
	// верните идентификатор последней добавленной записи
	lastInsertedID, err := r.LastInsertId()
	return int(lastInsertedID), err
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка

	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	sqlExpr := "SELECT number, client, status, address, created_at FROM parcel WHERE number=:number"

	r := s.db.QueryRow(sqlExpr,
		sql.Named("number", number))
	err := r.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	return p, err
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	var res []Parcel
	sqlExpr := "SELECT number,client, status, address, created_at FROM parcel WHERE client=:client"

	rows, err := s.db.Query(sqlExpr,
		sql.Named("client", client))
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {

		p := Parcel{}
		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	// заполните срез Parcel данными из таблицы
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	sqlExpr := "UPDATE parcel SET status=:status WHERE number=:number"

	_, err := s.db.Exec(sqlExpr,
		sql.Named("status", status),
		sql.Named("number", number))

	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	// отдаем ошибку, если нельзя обновить адрес
	var t int
	sqlExpr := "SELECT 1 FROM parcel WHERE number=:number AND status=:status"
	r := s.db.QueryRow(sqlExpr,
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))

	err := r.Scan(&t)
	if err != nil {
		return err
	}

	sqlExpr = "UPDATE parcel SET address=:address WHERE number=:number AND status=:status"
	_, err = s.db.Exec(sqlExpr,
		sql.Named("number", number),
		sql.Named("address", address),
		sql.Named("status", ParcelStatusRegistered))

	return err
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	sqlExpr := "DELETE FROM parcel WHERE number=:number and status=:status"

	_, err := s.db.Exec(sqlExpr,
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))

	return err
}
