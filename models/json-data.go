package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONData object
type JSONData map[string]interface{}

// Value marshal JSONData
func (p JSONData) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}

	j, err := json.Marshal(p)
	return j, err
}

// Scan unmarshal JSONData
func (p *JSONData) Scan(src interface{}) error {
	if src != nil {
		source, ok := src.([]byte)
		if !ok {
			return errors.New("type assertion .([]byte) failed")
		}

		var i interface{}
		err := json.Unmarshal(source, &i)
		if err != nil {
			return err
		}

		if i != nil {
			*p, ok = i.(map[string]interface{})
			if !ok {
				return errors.New("type assertion .(map[string]interface{}) failed")
			}
		}
	}

	return nil
}
