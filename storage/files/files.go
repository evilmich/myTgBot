package files

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"my_tg_bot/libs/er"
	"my_tg_bot/storage"
	"os"
	"path/filepath"
	"time"
)

type Storage struct {
	basePath string
}

const defaultPerm = 0774

func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s Storage) Save(page *storage.Page) (err error) {
	defer func() { err = er.WrapIfErr("can't save page", err) }()

	fPath := filepath.Join(s.basePath, page.UserName)

	if err := os.MkdirAll(fPath, defaultPerm); err != nil {
		return err
	}

	fName, err := fileName(page)
	if err != nil {
		return err
	}

	fPath = filepath.Join(fPath, fName)

	file, err := os.Create(fPath)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	return nil
}

func (s Storage) PickAll(userName string) (pages []*storage.Page, err error) {
	defer func() { err = er.WrapIfErr("can't pick all URLs", err) }()

	path := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	mas := make([]*storage.Page, 0)

	for _, file := range files {
		var a, _ = s.decodePage(filepath.Join(path, file.Name()))
		mas = append(mas, a)
	}

	return mas, nil
}

func (s Storage) PickRandom(userName string) (page *storage.Page, err error) {
	defer func() { err = er.WrapIfErr("can't pick random", err) }()

	path := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))

	n := rand.Intn(len(files))

	file := files[n]

	return s.decodePage(filepath.Join(path, file.Name()))
}

func (s Storage) Remove(p *storage.Page) error {
	fileName, err := fileName(p)

	if err != nil {
		return er.Wrap("can't remove file", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	if err := os.Remove(path); err != nil {
		msg := fmt.Sprintf("can't remove file %s", path)
		return er.Wrap(msg, err)
	}

	return nil
}

func (s Storage) IsExists(p *storage.Page) (bool, error) {
	fileName, err := fileName(p)

	if err != nil {
		return false, er.Wrap("can't check file exists", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	switch _, err := os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can't check file %s exists", path)
		return false, er.Wrap(msg, err)
	}

	return true, nil
}

func (s Storage) decodePage(filePath string) (*storage.Page, error) {
	f, err := os.Open(filePath)

	if err != nil {
		return nil, er.Wrap("can't decode page", err)
	}

	defer func() { _ = f.Close() }()

	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, er.Wrap("can't decode page", err)
	}

	return &p, nil
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}
