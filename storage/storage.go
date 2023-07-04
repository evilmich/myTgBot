package storage

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"my_tg_bot/libs/er"
)

type Storage interface {
	Save(ctx context.Context, p *Page) error
	PickAll(ctx context.Context, UserName string) ([]*Page, error)
	PickRandom(ctx context.Context, UserName string) (*Page, error)
	Remove(ctx context.Context, p *Page) error
	IsExists(ctx context.Context, p *Page) (bool, error)
}

var ErrNoSavedPages = errors.New("no saved pages")

type Page struct {
	URL      string
	UserName string
}

func (p Page) Hash() (string, error) {
	h := sha256.New()

	if _, err := io.WriteString(h, p.URL); err != nil {
		return "", er.Wrap("can't calculate hash", err)
	}

	if _, err := io.WriteString(h, p.UserName); err != nil {
		return "", er.Wrap("can't calculate hash", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
