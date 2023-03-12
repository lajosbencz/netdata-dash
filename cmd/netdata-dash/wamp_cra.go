package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gammazero/nexus/v3/stdlog"
)

type authUser struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Secret string `json:"secret"`
}

type fileKeyStore struct {
	Users []authUser `json:"users"`
}

type watchFileKeyStore struct {
	fileKeyStore
	filePath string
	watcher  *fsnotify.Watcher
	logger   stdlog.StdLog
}

func WatchFileKeyStore(filePath string, logger stdlog.StdLog) (*watchFileKeyStore, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watcher.Add(filePath)
	if err != nil {
		return nil, err
	}
	ks := &watchFileKeyStore{
		filePath: filePath,
		watcher:  watcher,
		logger:   logger,
	}
	if err := ks.UpdateFromFile(); err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := ks.UpdateFromFile(); err != nil {
						logger.Println(err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Println(err)
			}
		}
	}()
	return ks, nil
}

func (ks *watchFileKeyStore) UpdateFromFile() error {
	fks := &fileKeyStore{}
	f, err := os.Open(ks.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			ks.logger.Println(err, ks.filePath)
			return nil
		}
		return err
	}
	if err := json.NewDecoder(f).Decode(&fks); err != nil {
		return err
	}
	newUsers := []authUser{}
	for _, u := range fks.Users {
		if u.Name != "" && u.Role != "" && u.Secret != "" {
			newUsers = append(newUsers, u)
		}
	}
	ks.fileKeyStore.Users = newUsers
	ks.logger.Printf("config updated from %s to %#v\n", ks.filePath, ks.fileKeyStore)
	return nil
}

func (ks *watchFileKeyStore) GetUser(user string) (*authUser, error) {
	for _, v := range ks.Users {
		if v.Name == user {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("no such user: %s", user)
}

func (ks *watchFileKeyStore) Provider() string {
	return "dynamic"
}

func (ks *watchFileKeyStore) AuthKey(authid, authmethod string) ([]byte, error) {
	switch authmethod {
	case "ticket":
		u, err := ks.GetUser(authid)
		if err != nil {
			return nil, err
		}
		return []byte(u.Secret), nil
	}
	return nil, errors.New("unsupported authmethod")
}

func (ks *watchFileKeyStore) PasswordInfo(authid string) (string, int, int) {
	return "", 0, 0
}

func (ks *watchFileKeyStore) AuthRole(authid string) (string, error) {
	u, err := ks.GetUser(authid)
	if err != nil {
		return "", err
	}
	return u.Role, nil
}
