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

//func (s Storage) PickAll(ctx context.Context, userName string) (pages []*storage.Page, err error) {
//	defer func() { err = er.WrapIfErr("can't pick all URLs", err) }()
//
//	path := filepath.Join(s.basePath, userName)
//
//	files, err := os.ReadDir(path)
//
//	if err != nil {
//		return nil, err
//	}
//
//	if len(files) == 0 {
//		return nil, storage.ErrNoSavedPages
//	}
//
//	mas := make([]*storage.Page, 0)
//
//	for _, file := range files {
//		var a, _ = s.decodePage(filepath.Join(path, file.Name()))
//		mas = append(mas, a)
//	}
//
//	return mas, nil
//
//}

//func (s Storage) PickAll(ctx context.Context, userName string) (pages []*storage.Page, err error) {
//	defer func() { err = er.WrapIfErr("can't pick all URLs", err) }()
//
//	pipe := bson.A{
//		bson.M{"$sample": bson.M{"size": 1}},
//	}
//
//	cursor, err := s.pages.Aggregate(ctx, pipe)
//	if err != nil {
//		return nil, err
//	}
//
//	var p Page
//
//	cursor.Next(ctx)
//
//	err = cursor.Decode(&p)
//	switch {
//	case errors.Is(err, io.EOF):
//		return nil, storage.ErrNoSavedPages
//	case err != nil:
//		return nil, err
//	}
//
//	return &storage.Page{
//		URL:      p.URL,
//		UserName: p.UserName,
//	}, nil
//}

func (s Storage) PickRandom(ctx context.Context, UserName string) (page *storage.Page, err error) {
	defer func() { err = er.WrapIfErr("can't pick random page", err) }()

	pipe := bson.A{
		bson.M{"$sample": bson.M{"size": 1}},
	}

	log.Printf("pipe: %s\n", pipe)

	cursor, err := s.pages.Aggregate(ctx, pipe)

	if err != nil {
		return nil, err
	}

	log.Printf("cursor: %s\n", cursor)

	var p Page

	cursor.Next(ctx)

	err = cursor.Decode(&p)

	log.Printf("page: %s\n", p)

	log.Printf("err: %s\n", err)

	switch {
	case errors.Is(err, io.EOF):
		return nil, storage.ErrNoSavedPages
	case err != nil:
		return nil, err
	}

	mas := &storage.Page{
		URL:      p.URL,
		UserName: p.UserName,
	}

	log.Printf("storage page: %s\n", mas)

	return mas, nil
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
