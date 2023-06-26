package telegram

import (
	"context"
	"errors"
	"log"
	"my_tg_bot/libs/er"
	"my_tg_bot/storage"
	"net/url"
	"strings"
)

const (
	AllCmd   = "/all"
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

func (p *ProcessorOver) doCmd(ctx context.Context, text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command '%s' from '%s", text, username)

	if isAddCmd(text) {
		return p.savePage(ctx, chatID, text, username)
	}

	switch text {
	//case AllCmd:
	//	return p.sendAll(ctx, chatID, username)
	case RndCmd:
		return p.sendRandom(ctx, chatID, username)
	case HelpCmd:
		return p.sendHelp(ctx, chatID)
	case StartCmd:
		return p.sendHello(ctx, chatID)
	default:
		return p.tg.SendMessage(ctx, chatID, msgUnknownCommand)
	}
}

func (p *ProcessorOver) savePage(ctx context.Context, chatID int, pageURL string, username string) (err error) {
	defer func() { err = er.WrapIfErr("can't do command: save page", err) }()

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(ctx, page)
	if err != nil {
		return err
	}
	if isExists {
		return p.tg.SendMessage(ctx, chatID, msgAlreadyExists)
	}

	if err := p.storage.Save(ctx, page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(ctx, chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

//func (p *ProcessorOver) sendAll(ctx context.Context, chatID int, username string) (err error) {
//	defer func() { err = er.WrapIfErr("can't do command: can't send all URLs", err) }()
//
//	pages, err := p.storage.PickAll(ctx, username)
//
//	log.Printf("pages: %s", pages)
//
//	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
//		return err
//	}
//
//	if errors.Is(err, storage.ErrNoSavedPages) {
//		return p.tg.SendMessage(ctx, chatID, msgNoSavedPages)
//	}
//
//	for _, page := range pages {
//		if err := p.tg.SendMessage(ctx, chatID, page.URL); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}

func (p *ProcessorOver) sendRandom(ctx context.Context, chatID int, username string) (err error) {
	defer func() { err = er.WrapIfErr("can't do command: can't send random", err) }()

	page, err := p.storage.PickRandom(ctx, username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(ctx, chatID, msgNoSavedPages)
	}

	if err := p.tg.SendMessage(ctx, chatID, page.URL); err != nil {
		return err
	}

	return p.storage.Remove(ctx, page)
}

func (p *ProcessorOver) sendHelp(ctx context.Context, chatID int) error {
	return p.tg.SendMessage(ctx, chatID, msgHelp)
}

func (p *ProcessorOver) sendHello(ctx context.Context, chatID int) error {
	return p.tg.SendMessage(ctx, chatID, msgHello)
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
