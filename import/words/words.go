package words

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

type Word struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type Loader interface {
	LoadWords() ([]Word, error)
}

func NewLoader(wordsFilePath string) Loader {
	return &loader{
		wordsFilePath: wordsFilePath,
	}
}

type loader struct {
	wordsFilePath string
}

func (l *loader) LoadWords() ([]Word, error) {
	if _, err := os.Stat(l.wordsFilePath); err != nil {
		return nil, errors.Wrapf(err, "file cannot be found, path: %v", l.wordsFilePath)
	}
	if jsonFile, err := os.Open(l.wordsFilePath); err != nil {
		return nil, errors.Wrapf(err, "file cannot be opened, path: %v", l.wordsFilePath)
	} else {
		byteValue, _ := ioutil.ReadAll(jsonFile)
		type wrapper struct {
			Words []Word
		}
		w := wrapper{}
		err := json.Unmarshal(byteValue, &w)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot parse JSON with flip words, path: %v", l.wordsFilePath)
		}
		return w.Words, nil
	}
}
