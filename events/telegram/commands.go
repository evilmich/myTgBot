package telegram

import (
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

func (p *ProcessorOver) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command '%s' from '%s'", text, username)

	if isAddCmd(text) {
		return p.savePage(chatID, text, username)
	}

	switch text {
	case AllCmd:
		return p.sendAll(chatID, username)
	case RndCmd:
		return p.sendRandom(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	default:
		return p.tg.SendMessage(chatID, msgUnknownCommand)
	}
}

func (p *ProcessorOver) savePage(chatID int, pageUrl string, username string) (err error) {
	defer func() { err = er.WrapIfErr("can't do command: save page", err) }()

	page := &storage.Page{
		URL:      pageUrl,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(page)
	if err != nil {
		return err
	}

	if isExists {
		return p.tg.SendMessage(chatID, msgAlreadyExists)
	}

	if err := p.storage.Save(page); err != nil {
		return err
	}
	if err := p.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}
	return nil
}

func (p *ProcessorOver) sendAll(chatID int, username string) (err error) {
	defer func() { err = er.WrapIfErr("can't do command: can't send all URLs", err) }()

	pages, err := p.storage.PickAll(username)

	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	for _, page := range pages {
		if err := p.tg.SendMessage(chatID, page.URL); err != nil {
			return err
		}
	}

	return nil
}

func (p *ProcessorOver) sendRandom(chatID int, username string) (err error) {
	defer func() { err = er.WrapIfErr("can't do command: can't send random", err) }()

	page, err := p.storage.PickRandom(username)

	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	if err := p.tg.SendMessage(chatID, page.URL); err != nil {
		return err
	}

	return p.storage.Remove(page)
}

func (p *ProcessorOver) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *ProcessorOver) sendHello(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
