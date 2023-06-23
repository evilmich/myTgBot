package storage

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"my_tg_bot/libs/er"
)

type Storage interface {
	Save(p *Page) error
	PickAll(UserName string) ([]*Page, error)
	PickRandom(UserName string) (*Page, error)
	Remove(p *Page) error
	IsExists(p *Page) (bool, error)
}

var ErrNoSavedPages = errors.New("no saves pages")

type Page struct {
	URL      string
	UserName string
}

func (p Page) Hash() (string, error) {
	h := sha256.New()

	if _, err := io.WriteString(h, p.URL); err != nil {
		return "", er.Wrap("can't calculate hash(URL)", err)
	}

	if _, err := io.WriteString(h, p.UserName); err != nil {
		return "", er.Wrap("can't calculate hash(UserName)", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
