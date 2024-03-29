package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type user_doc struct {
	Username UserId
	Hash     string
	Balance  float32
	Stonks   map[string]float64
}

// ======== redis ==========
func setupRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		DB:       0,
		Password: "",
		Addr:     redisHOST + ":6379",
	})
	return client
}

// Wrapper around Redis for keeping track of pending purchases
type TransactionStore struct {
	rdb *redis.Client
	ctx context.Context
}

func (t *TransactionStore) ShutDown() {
	t.rdb.Close()
}
func NewTransactionStore() *TransactionStore {
	return &TransactionStore{setupRedis(), context.Background()}

}
func (b *Notification) Pending(t PendingTransactorSource) error {
	err := t.Store(b.Userid, b.Topic, b)
	return err
}

func (t *TransactionStore) lastPending(uid UserId, topic CommandType) (*Notification, error) {
	ctx := context.Background()
	val, err := t.rdb.GetDel(ctx, string(uid)+"#"+string(topic)).Bytes()
	if err != nil {
		return nil, err
	}

	var n Notification
	err = json.Unmarshal(val, &n)
	return &n, err
}
func (t TransactionStore) Store(uid UserId, topic CommandType, n *Notification) error {
	ctx := context.Background()
	err := t.rdb.Set(ctx, string(uid)+"#"+string(topic), n, 60*time.Second).Err()
	return err

}

func (i *Notification) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}
func (i *Stock) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

// ===== mongo ====

const database = "day_trading"

func connect() (*mongo.Client, context.Context) {
	clientOptions := options.Client()
	clientOptions.ApplyURI("mongodb://admin:admin@10.9.0.3:27017")
	// clientOptions.ApplyURI("mongodb://admin:admin@localhost:27017")
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Hour)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		fmt.Println("Error connecting to DB")
		panic(err)
	}
	return client, ctx
}

// Wrapper around mongo for reading and updating users
type UserStore struct {
	db  *mongo.Client
	ctx context.Context
}

func NewUserStore() *UserStore {
	db, ctx := connect()
	return &UserStore{db, ctx}
}

func (u *UserStore) Execute(t func(context.Context) error) error {
	transactionError := u.db.UseSession(context.TODO(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		err = t(sessionContext)
		if err != nil {
			sessionContext.AbortTransaction(context.TODO())
			return err
		}

		err = sessionContext.CommitTransaction(context.TODO())
		return err
	})

	if transactionError != nil {
		transactionError = errors.New("ERROR: Failed Transaction with error " + transactionError.Error())
	}
	return transactionError
}

func (s *UserStore) getUser(note CommandType, uid UserId) (user_collection *user_doc, err error) {
	if s.db == nil {
		s.db, s.ctx = connect()
	}
	var result user_doc
	err = s.db.Database(database).Collection("users").FindOne(s.ctx, bson.D{{"username", uid}}).Decode(&result)

	if err != nil {
		if err.Error() == "mongo: no documents in result" && note == notifyADD {

			var new_doc = new(user_doc)
			new_doc.Username = uid
			new_doc.Hash = "unsecure_this_user_never_made_account_via_backend"
			new_doc.Balance = 0
			new_doc.Stonks = make(map[string]float64)

			collection := s.db.Database(database).Collection("users")
			_, err = collection.InsertOne(context.TODO(), new_doc)

			if err != nil {
				fmt.Println("Error adding user to db: ", err)
				panic(err)
			}

			//defer db.Disconnect(ctx)
			return new_doc, nil

		} else {
			return nil, (err)
		}

	}
	return &result, err
}
func (n *Notification) ReadUser(s UserTransactorSource) (user_collection *user_doc, err error) {

	return s.getUser(n.Topic, n.Userid)
}

func (s *UserStore) setUser(username UserId, balance float32, stocks map[string]float64) error {
	if s.db == nil {
		s.db, s.ctx = connect()
	}
	collection := s.db.Database(database).Collection("users")

	selected_user := bson.M{"username": username}
	updated_user := bson.M{"$set": bson.M{"balance": balance, "stonks": stocks}}
	_, err := collection.UpdateOne(context.TODO(), selected_user, updated_user)

	if err != nil {
		fmt.Println("Error inserting into db: ", err)
		return err
	}
	return err

}

func (u *user_doc) Backup(s UserTransactorSource) error {
	return s.setUser(u.Username, u.Balance, u.Stonks)
}
func (us *UserStore) Disconnect() {
	us.db.Disconnect(us.ctx)
}
