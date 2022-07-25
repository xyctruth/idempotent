package idempotent

import (
	"errors"
	"fmt"
	"github.com/xyctruth/idempotent/drive"
	"log"
	"runtime/debug"
	"time"
)

type Idempotent struct {
	global drive.Drive
	config *Config
}

func New(d drive.Drive, opts ...ConfigOpt) (*Idempotent, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	i := &Idempotent{global: d, config: config}
	if err := i.start(); err != nil {
		return nil, err
	}
	return i, nil
}

func (i *Idempotent) Acquire(key string, engine drive.Drive, opts ...ConfigOpt) bool {
	config := &Config{
		TTL: i.config.TTL,
	}
	for _, opt := range opts {
		opt(config)
	}
	ok, err := engine.Acquire(key, config.TTL)
	if err != nil {
		log.Println(err)
	}
	return ok
}

func (i *Idempotent) start() error {
	err := i.global.Init()
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			if p := recover(); p != nil {
				pncMsg := fmt.Sprintf("%v\n%s", p, debug.Stack())
				err := errors.New(pncMsg)
				log.Println(err)
			}
		}()

		clear := time.NewTicker(i.config.ClearExpiryDuration)
		for {
			select {
			case <-clear.C:
				err := i.global.Clear()
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()
	return nil
}
