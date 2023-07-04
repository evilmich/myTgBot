package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"my_tg_bot/libs/er"
	"my_tg_bot/storage"
	"time"
)

type Storage struct {
	pages Pages
}

type Pages struct {
	*mongo.Collection
}

type Page struct {
	URL      string `bson:"url"`
	UserName string `bson:"username"`
}

func New(connectString string, connectTimeout time.Duration) Storage {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectString))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	pages := Pages{
		Collection: client.Database("read-adviser").Collection("pages"),
	}

	return Storage{
		pages: pages,
	}
}

func (s Storage) Save(ctx context.Context, page *storage.Page) error {
	_, err := s.pages.InsertOne(ctx, Page{
		URL:      page.URL,
		UserName: page.UserName,
	})
	if err != nil {
		return er.Wrap("can't save page", err)
	}

	return nil
}

func (s Storage) PickAll(ctx context.Context, UserName string) (pages []*storage.Page, err error) {
	defer func() { err = er.WrapIfErr("can't pick all URLs", err) }()

	filter := UserNameFilter(UserName)

	options := options.Find()
	options.SetSort(bson.M{"created": -1})

	cursor, err := s.pages.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}

	var mas []*storage.Page

	for cursor.Next(ctx) {
		p := &storage.Page{}
		err := cursor.Decode(&p)

		if err != nil {
			return nil, err
		} else {
			mas = append(mas, p)
		}
	}

	if mas == nil {
		return nil, storage.ErrNoSavedPages
	}

	return mas, nil
}

func (s Storage) PickRandom(ctx context.Context, UserName string) (page *storage.Page, err error) {
	defer func() { err = er.WrapIfErr("can't pick random page", err) }()

	filter := UserNameFilter(UserName)

	pipe := bson.A{bson.M{"$match": filter}, bson.M{"$sample": bson.M{"size": 1}}}

	cursor, err := s.pages.Aggregate(ctx, pipe)

	if err != nil {
		return nil, err
	}

	var p Page

	cursor.Next(ctx)

	err = cursor.Decode(&p)

	switch {
	case errors.Is(err, io.EOF):
		return nil, storage.ErrNoSavedPages
	case err != nil:
		return nil, err
	}

	return &storage.Page{
		URL:      p.URL,
		UserName: p.UserName,
	}, nil
}

func (s Storage) Remove(ctx context.Context, storagePage *storage.Page) error {
	_, err := s.pages.DeleteOne(ctx, toPage(storagePage).Filter())
	if err != nil {
		return er.Wrap("can't remove page", err)
	}

	return nil
}

func (s Storage) IsExists(ctx context.Context, storagePage *storage.Page) (bool, error) {
	count, err := s.pages.CountDocuments(ctx, toPage(storagePage).Filter())
	if err != nil {
		return false, er.Wrap("can't check if page exists", err)
	}

	return count > 0, nil
}

func toPage(p *storage.Page) Page {
	return Page{
		URL:      p.URL,
		UserName: p.UserName,
	}
}

func (p Page) Filter() bson.M {
	return bson.M{
		"url":      p.URL,
		"username": p.UserName,
	}
}

func UserNameFilter(UserName string) bson.M {
	var filters = []bson.M{
		{
			"username": UserName,
		},
	}
	return bson.M{"$and": filters}
}
