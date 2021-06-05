/*
Q: 我们在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。为什么，应该怎么做请写出代码？

A: 应该 Wrap ，sql.ErrNoRows 是根因，写 DAO 的时候就是做 Application 的开发，
   需要保留返回给上层，这样才能兼容 Sentinel Error 和 type assertions 的处理，
   并且能够附带上业务的信息，而且使用 Wrap 的话，能够拿到 Error 的整个堆栈信息。

Error Handling Note:
1. 主动：开发应用时，自己产生的 Error ，使用 errors.New 或者 errors.Errorf 返回带堆栈信息的 Error;
2. 协作：使用官方库或第三方库，就使用 pkg/errors 的 Wrap 包起来;

 */
package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/pkg/errors"

	_ "github.com/mattn/go-sqlite3"

)

type Device struct {
	Id int32
	Uuid string
	DeviceName string
	IsOnline int32
}

type IDeviceDAO interface {
	Get(id string) (Device, error)
}

type DeviceDAO struct {
	db sql.DB
}

func (d *DeviceDAO) Init() {
	db, err := GetSqliteConn()
	if err != nil {
		return
	}
	d.db = db
}

func (d *DeviceDAO) Get(id string) (*Device, error) {
	sqlString := "SELECT id, uuid, \"device name\", \"is online\" FROM device WHERE id = " + id
	device := Device{}
	err := d.db.QueryRow(sqlString).Scan(&device.Id, &device.Uuid, &device.DeviceName, &device.IsOnline)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrapf(err, "Can not found device, id: %s", id)
		}
	}
	return &device, nil
}

func GetSqliteConn() (sql.DB, error) {
	dbPath := "./ch2.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Println(err)
	}
	return *db, nil
}

func main() {
	deviceDao := DeviceDAO{}
	deviceDao.Init()
	device, err := deviceDao.Get("1")
	if err != nil {
		fmt.Printf("original error: %T %v\n", errors.Cause(err), errors.Cause(err))
		fmt.Printf("stack trace: \n%+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("uuid: %s\ndevice_name: %s", device.Uuid, device.DeviceName)
	defer deviceDao.db.Close()
}