package models

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
)

// DatasetList represents a structure for a list of datasets
type DatasetList struct {
	Items           []Dataset `json:"items"`
	NumberOfResults int64     `json:"number_of_results"`
}

// Dataset represents information related to a single dataset
type Dataset struct {
	Contact     ContactDetails `json:"contact,omitempty"`
	Edition     string         `json:"edition,omitempty"`
	ID          string         `json:"id"`
	NextRelease string         `json:"next_release,omitempty"`
	ReleaseDate string         `json:"release_date,omitempty"`
	Title       string         `json:"title,omitempty"`
	URL         string         `json:"url,omitempty"`
	Version     string         `json:"version,omitempty"`
}

// ContactDetails represents an object containing information of the contact
type ContactDetails struct {
	Email     string `json:"email,omitempty"`
	Name      string `json:"name,omitempty"`
	Telephone string `json:"telephone,omitempty"`
}

// CreateDataset manages the creation of a dataset from a reader
func CreateDataset(reader io.Reader) (*Dataset, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Failed to read message body")
	}
	var datset Dataset
	err = json.Unmarshal(bytes, &datset)
	if err != nil {
		return nil, errors.New("Failed to parse json body")
	}

	return &datset, nil
}
