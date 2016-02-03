package store

import (
	"encoding/json"
	"fmt"
	"github.com/amassanet/gopad/model"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const (
    oldentriesPath = "/entries/old/"
	entriesPath    = "/entries/"
	jsonExt        = ".json"
	txtExt         = ".txt"
    dateTimeFormat  = "20060102150405"
)

type EntryStore struct {
	Config
}

func NewEntryStore(config Config) *EntryStore {
	if err := os.MkdirAll(config.path+entriesPath, 0744); err != nil {
		log.Fatalf("Cannot create folder %v", err)
	}
	if err := os.MkdirAll(config.path+oldentriesPath, 0744); err != nil {
		log.Fatalf("Cannot create folder %v", err)
	}
	return &EntryStore{config}
}

func (es *EntryStore) add(entry *model.Entry) error {

	filename := replaceFilenameChars(entry.Title) + "_" + entry.ID

	err := ioutil.WriteFile(es.path+entriesPath+filename+txtExt, []byte(entry.Markdown), 0644)

	if err != nil {
		return err
	}

	encoded, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(es.path+entriesPath+filename+jsonExt, encoded, 0644)

	if err != nil {
		return err
	}

	return nil
}

func (es *EntryStore) Add(entry *model.Entry) error {

	entry.ID = time.Now().Format(dateTimeFormat)

	err := es.add(entry)

	if err != nil {
		return err
	}

	return nil
}

func (es *EntryStore) Update(entry *model.Entry) error {

	filename, err := es.getFilenameForID(entry.ID)
	if err != nil {
		return err
	}

    now := time.Now().Format(dateTimeFormat)
	os.Rename(
		es.path+entriesPath+filename+txtExt,
		es.path+oldentriesPath+filename+"_"+now+txtExt,
	)
	os.Rename(
		es.path+entriesPath+filename+jsonExt,
		es.path+oldentriesPath+filename+"_"+now+jsonExt,
	)

	err = es.add(entry)
	if err != nil {
		return err
	}

	return nil
}

func (es *EntryStore) get(filename string) (*model.Entry, error) {

	rawjson, err := ioutil.ReadFile(es.path + entriesPath + filename + jsonExt)
	if err != nil {
		return nil, err
	}
	var entry model.Entry
	err = json.Unmarshal(rawjson, &entry)
	if err != nil {
		return nil, err
	}

	rawentry, err := ioutil.ReadFile(es.path + entriesPath + filename + txtExt)
	if err != nil {
		return nil, err
	}
	entry.Markdown = string(rawentry)

	return &entry, nil
}

func (es *EntryStore) Get(ID string) (*model.Entry, error) {

	filename, err := es.getFilenameForID(ID)
	if err != nil {
		return nil, err
	}

	entry, err := es.get(filename)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (es *EntryStore) getFilenameForID(ID string) (string, error) {
	fileInfos, err := ioutil.ReadDir(es.path + entriesPath)

	if err != nil {
		return "", err
	}

	suffix := "_" + ID + txtExt
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			name := fileInfo.Name()
			if strings.HasSuffix(name, suffix) {
				return name[:len(name)-len(txtExt)], nil
			}
		}
	}

	return "", fmt.Errorf("File not found for entry %v", ID)
}

func (es *EntryStore) List() ([]*model.Entry, error) {

	fileInfos, err := ioutil.ReadDir(es.path + entriesPath)

	if err != nil {
		return nil, err
	}

	entries := make([]*model.Entry, 0, len(fileInfos))

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), txtExt) {
			name := fileInfo.Name()
			entry, err := es.get(name[:len(name)-len(txtExt)])
			if err != nil {
				return nil, err
			}
			entries = append(entries, entry)
		}
	}
	return entries[:], nil
}