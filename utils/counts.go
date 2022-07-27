package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type Count struct {
	Filename string
	data     map[string]uint64
}

func (c *Count) Load() error {
	content, err := ioutil.ReadFile(c.Filename)
	c.data = make(map[string]uint64)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	err = json.Unmarshal(content, &c.data)
	if err != nil {
		return err
	}

	return nil
}

func (c *Count) Get(key string) uint64 {
	value, ok := c.data[key]
	if ok {
		return value
	}

	return 0
}

func (c *Count) Set(key string, value uint64) {
	c.data[key] = value
}

func (c *Count) Save() error {
	counts, err := json.Marshal(c.data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.Filename, counts, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
