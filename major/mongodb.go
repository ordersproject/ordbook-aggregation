package major

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ordbook-aggregation/conf"
	"ordbook-aggregation/config"
	"sync"
	"time"
)

type Database struct {
	MongoClientMap map[string]*mongo.Client
}

type MGOConfig struct {
	DsName    string
	Addrs     string
	Timeout   int64
	Database  string
	Username  string
	Password  string
	PoolLimit int
}


var _mongoOnce sync.Once
var mongoDB *Database

func InitMongo() {
	_mongoOnce.Do(func() {
		mongoDB = &Database{MongoClientMap: make(map[string]*mongo.Client, 0)}
		err := SetConnect()
		if err != nil {
			panic("Mongo init error")
		}
		fmt.Println("Mongodb init done")
	})
}

// 连接设置
func SetConnect() error {
	mongoConfigs := []MGOConfig{}
	if err := config.ReadJsonConfig(conf.MDBEnvironment(), &mongoConfigs); err != nil {
		fmt.Println("ReadJsonConfigError ", err)
		return nil
	}
	for _, mongoConfig := range mongoConfigs {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mongoConfig.Timeout)*time.Second)
		defer cancel()
		mongoUrl := "mongodb://" + mongoConfig.Username + ":" + mongoConfig.Password + "@" + mongoConfig.Addrs + "/" + mongoConfig.Database
		var err error
		mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUrl).SetMaxPoolSize(uint64(mongoConfig.PoolLimit)))
		if err != nil {
			return err
		}
		if mongoDB != nil {
			mongoDB.MongoClientMap[mongoConfig.DsName] = mongoClient
		} else {
			return err
		}
	}
	return nil
}

func GetDBWith(dbName string) (*mongo.Client, error) {
	if mongoDB.MongoClientMap[dbName] != nil {
		return mongoDB.MongoClientMap[dbName], nil
	}
	return nil, errors.New("MongoClientMap is nil")
}
