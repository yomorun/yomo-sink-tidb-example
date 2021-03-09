# TIDBCONNECTOR

Tidb 🙌 YoMo. Demonstrates how to integrate Tidb to YoMo and bulk insert data into Tidb after stream processing.

## About Tidb

TiDB ("Ti" stands for Titanium) is an open-source NewSQL database that supports Hybrid Transactional and Analytical Processing (HTAP) workloads. It is MySQL compatible and features horizontal scalability, strong consistency, and high availability.

For more information, please visit [Tidb homepage](https://github.com/pingcap/tidb).

## About YoMo

[YoMo](https://github.com/yomorun/yomo) is an open-source Streaming Serverless Framework for building Low-latency Edge Computing applications. Built atop QUIC Transport Protocol and Functional Reactive Programming interface. makes real-time data processing reliable, secure, and easy.

## Quick Start

### Install yomo cli

please visit [Yomo homepage](https://github.com/yomorun/yomo#1-install-cli).

### Create your serverless app

please visit [Yomo homepage](https://github.com/yomorun/yomo#2-create-your-serverless-app).

### Copy `app.go` to your serverless app

```go
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


```

### Run Tidb with Docker

```bash
=> docker run --name tidb-server -d -v /tidb/data:/tmp/tidb -p 4000:4000 -p 10080:10080 pingcap/tidb:latest

=> mysql -h 127.0.0.1 -P 4000 -u root -D test --prompt="tidb> "
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 30
Server version: 5.7.25-TiDB-v4.0.10 TiDB Server (Apache License 2.0) Community Edition, MySQL 5.7 compatible

Copyright (c) 2000, 2016, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

tidb>

```

### Create user && Create database && Create table

```bash
CREATE DATABASE demo CHARACTER SET utf8 COLLATE utf8_general_ci;
CREATE USER 'yomo'@'%' IDENTIFIED BY 'yomo';
GRANT ALL PRIVILEGES ON demo.* TO 'yomo'@'%';
FLUSH PRIVILEGES;
CREATE TABLE IF NOT EXISTS `noises`(`id` INT UNSIGNED AUTO_INCREMENT,`noise` FLOAT(6,2) NOT NULL, `time` BIGINT NOT NULL,PRIMARY KEY ( `id` ))ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

### Run your serverless app in development

```bash
⇒  yomo dev
2021/03/09 18:49:31 Building the Serverless Function File...
2021/03/09 18:49:33 ✅ Listening on 0.0.0.0:4242
[StdOut]:  ⚡️ 30 successfully stored in the tidb
[StdOut]:  ⚡️ 30 successfully stored in the tidb
[StdOut]:  ⚡️ 30 successfully stored in the tidb
```
### Verify data in Macrometa
```bash
tidb> select * from noises;
+----+--------+---------------+
| id | noise  | time          |
+----+--------+---------------+
|  1 | 146.73 | 1615286973445 |
|  2 |  22.21 | 1615286973545 |
|  3 |  55.20 | 1615286973645 |
|  4 |  38.34 | 1615286973746 |
|  5 |  79.72 | 1615286973846 |
|  6 |   3.55 | 1615286973946 |
|  7 | 162.90 | 1615286974046 |
|  8 |  34.47 | 1615286974146 |
|  9 |  19.12 | 1615286974246 |
| 10 |  44.86 | 1615286974346 |
| 11 |   5.50 | 1615286974447 |
```


