package store

import (
	"encoding/json"
	"github.com/amassanet/gopad/model"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
    "errors"
    "fmt"
    "sort"
    "regexp"
)

const (
    oldentriesPath = "/entries/old/"
	entriesPath    = "/entries/"
	jsonExt        = ".json"
	mdExt          = ".md"
    dateTimeFormat  = "20060102150405"
)

var (
    errNotExists = errors.New("File does not exist")
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

	filename := entry.ID + "_" + replaceFilenameChars(entry.Title)
    txtPath := es.path+entriesPath+filename+mdExt
    jsonPath := es.path+entriesPath+filename+jsonExt

    if _, err := os.Stat(txtPath); err == nil {
        panic("Oops, cannot override files!")
    }
    if _, err := os.Stat(jsonPath); err == nil {
        panic("Oops, cannot override files!")
    }

	err := ioutil.WriteFile(txtPath, []byte(entry.Markdown), 0644)

	if err != nil {
		return err
	}

	encoded, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(jsonPath, encoded, 0644)

	if err != nil {
		return err
	}

	return nil
}

func (es *EntryStore) NewID() string {

	return time.Now().Format(dateTimeFormat)

}

func (es *EntryStore) Store(entry *model.Entry) error {

	filename, err := es.getFilenameForID(entry.ID)

    if err != nil && err != errNotExists {
		return err
	}

    if err == nil {
        now := time.Now().Format(dateTimeFormat)
        os.Rename(
            es.path+entriesPath+filename+mdExt,
            es.path+oldentriesPath+filename+"_"+now+mdExt,
        )
        os.Rename(
            es.path+entriesPath+filename+jsonExt,
            es.path+oldentriesPath+filename+"_"+now+jsonExt,
        )
    }

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

	rawentry, err := ioutil.ReadFile(es.path + entriesPath + filename + mdExt)
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

    prefix := ID + "_"
    suffix := mdExt
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			name := fileInfo.Name()
			if strings.HasPrefix(name,prefix) && strings.HasSuffix(name, suffix) {
				return name[:len(name)-len(suffix)], nil
			}
		}
	}

	return "", errNotExists
}

func (es *EntryStore) List() ([]*model.Entry, error) {

	fileInfos, err := ioutil.ReadDir(es.path + entriesPath)

	if err != nil {
		return nil, err
	}

    sort.Sort(FileInfos(fileInfos))

	entries := make([]*model.Entry, 0, len(fileInfos))

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), mdExt) {
			name := fileInfo.Name()
            pos := strings.Index(name,".")
            if pos == -1 {
                return nil,fmt.Errorf("Bad filename %v",name)
            }
            entry, err := es.get(name[:pos])
			if err != nil {
				return nil, err
			}
			entries = append(entries, entry)
		}
	}
	return entries[:], nil
}

type SearchResult struct {
    ID string
    Title string
    Matches []string
}

func (es *EntryStore) Search(expr string) ( []SearchResult, error) {

    rg,err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	fileInfos, err := ioutil.ReadDir(es.path + entriesPath)

	if err != nil {
		return nil, err
	}

    sort.Sort(FileInfos(fileInfos))

	results := []SearchResult{}

	for _, fileInfo := range fileInfos {

        if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), mdExt) {

            name := fileInfo.Name()
            pos := strings.Index(name,".")
            if pos == -1 {
                return nil,fmt.Errorf("Bad filename %v",name)
            }

            entry, err := es.get(name[:pos])
			if err != nil {
				return nil, err
			}

            matches := []string{}
            lines := strings.Split(entry.Markdown,"\n")
            for _,line := range lines {
                if rg.MatchString(line) {
                    matches = append (matches, line)
                }
            }

            if len(matches) > 0 {
                results = append(results,
                    SearchResult{entry.ID,entry.Title,matches})
            }
		}
	}
	return results, nil
}


