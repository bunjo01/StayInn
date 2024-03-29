package cache

import (
	"encoding/json"
	"io"
)

type Image struct {
	ID    string `json:"id"`
	AccID string `json:"acc_id"`
	Data  []byte `json:"data"`
}

type Images []*Image

func (i *Image) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
}

func (i *Image) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(i)
}

func (i *Images) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
}

func (i *Images) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(i)
}
