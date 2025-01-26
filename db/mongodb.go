package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

// 몽고디비 클라이언트 초기화 및 핑테스트
func InitMongoDB(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("MongoDB 연결 실패: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("MongoDB 핑 실패: %w", err)
	}

	MongoClient = client
	fmt.Println("MongoDB에 성공적으로 연결되었습니다!")
	return nil
}

// 몽고디비 컬렉션 가져옹기
func GetCollection(databaseName, collectionName string) *mongo.Collection {
	return MongoClient.Database(databaseName).Collection(collectionName)
}

// 몽고디비에 값 넣기
func InsertDocument(databaseName, collectionName string, document interface{}) error {
	collection := GetCollection(databaseName, collectionName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, document)
	if err != nil {
		return fmt.Errorf("문서 삽입 실패: %w", err)
	}

	fmt.Printf("문서가 %s 컬렉션에 성공적으로 저장되었습니다.\n", collectionName)
	return nil
}
