package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/reactivex/rxgo/v2"
	y3 "github.com/yomorun/y3-codec-golang"
	"github.com/yomorun/yomo/pkg/rx"
)

// NoiseDataKey represents the Tag of a Y3 encoded data packet
const NoiseDataKey = 0x10

var (
	MysqlDB *sql.DB = nil
)

func init() {
	var err error = nil
	MysqlDB, err = sql.Open("mysql", "yomo:yomo@tcp(127.0.0.1:4000)/demo")

	if err != nil {
		panic(err)
	}
}

type NoiseData struct {
	Noise float32 `y3:"0x11"`
	Time  int64   `y3:"0x12"`
	From  string  `y3:"0x13"`
}

// save data to tidb
var saveDocs = func(_ context.Context, i interface{}) (interface{}, error) {
	execstring := "INSERT INTO noises(`noise`, `time`) VALUES "

	for j, v := range i.([]interface{}) {
		noise := v.(NoiseData)
		if j == len(i.([]interface{}))-1 {
			execstring = execstring + fmt.Sprintf("(%f,%d);", noise.Noise, noise.Time)
		} else {
			execstring = execstring + fmt.Sprintf("(%f,%d),", noise.Noise, noise.Time)
		}
	}

	_, err := MysqlDB.Exec(execstring)

	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("⚡️ %d successfully stored in the tidb", len(i.([]interface{}))), nil
}

// y3 callback
var callback = func(v []byte) (interface{}, error) {
	var mold NoiseData
	err := y3.ToObject(v, &mold)

	if err != nil {
		return nil, err
	}
	return mold, nil
}

// Handler will handle data in Rx way
func Handler(rxstream rx.RxStream) rx.RxStream {
	if MysqlDB == nil {
		panic("not found MysqlDB.")
	}

	stream := rxstream.
		Subscribe(NoiseDataKey).
		OnObserve(callback).
		BufferWithTimeOrCount(rxgo.WithDuration(5*time.Second), 30).
		Map(saveDocs).
		StdOut().
		Encode(NoiseDataKey)
	return stream
}
