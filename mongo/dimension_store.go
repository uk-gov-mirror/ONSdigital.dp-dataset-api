package mongo

import (
	"fmt"
	"time"

	errs "github.com/ONSdigital/dp-dataset-api/apierrors"
	"github.com/ONSdigital/dp-dataset-api/models"
	"gopkg.in/mgo.v2/bson"
)

const dimensionOptions = "dimension.options"

// GetDimensionNodesFromInstance which are stored in a mongodb collection
func (m *Mongo) GetDimensionNodesFromInstance(id string) (*models.DimensionNodeResults, error) {
	s := m.Session.Copy()
	defer s.Close()

	var dimensions []models.DimensionOption
	iter := s.DB(m.Database).C(dimensionOptions).Find(bson.M{"instance_id": id}).Select(bson.M{"id": 0, "last_updated": 0, "instance_id": 0}).Iter()

	err := iter.All(&dimensions)
	if err != nil {
		return nil, err
	}

	return &models.DimensionNodeResults{Items: dimensions}, nil
}

// GetUniqueDimensionValues which are stored in mongodb collection
func (m *Mongo) GetUniqueDimensionValues(id, dimension string) (*models.DimensionValues, error) {
	s := m.Session.Copy()
	defer s.Close()

	var values []string
	err := s.DB(m.Database).C(dimensionOptions).Find(bson.M{"instance_id": id, "name": dimension}).Distinct("option", &values)
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, errs.ErrDimensionNodeNotFound
	}

	return &models.DimensionValues{Name: dimension, Values: values}, nil
}

// AddDimensionToInstance to the dimension collection
func (m *Mongo) AddDimensionToInstance(opt *models.CachedDimensionOption) error {
	s := m.Session.Copy()
	defer s.Close()

	option := models.DimensionOption{InstanceID: opt.InstanceID, Option: opt.Option, Name: opt.Name, Label: opt.Label}
	option.Links.CodeList = models.LinkObject{ID: opt.CodeList, HRef: fmt.Sprintf("%s/code-lists/%s", m.CodeListURL, opt.CodeList)}
	option.Links.Code = models.LinkObject{ID: opt.Code, HRef: fmt.Sprintf("%s/code-lists/%s/codes/%s", m.CodeListURL, opt.CodeList, opt.Code)}

	option.LastUpdated = time.Now().UTC()
	_, err := s.DB(m.Database).C(dimensionOptions).Upsert(bson.M{"instance_id": option.InstanceID, "name": option.Name,
		"option": option.Option}, &option)
	if err != nil {
		return err
	}

	return nil
}

// GetDimensions returns a list of all dimensions from a dataset
func (m *Mongo) GetDimensions(datasetID, editionID, versionID string) (*models.DatasetDimensionResults, error) {
	s := m.Session.Copy()
	defer s.Close()

	version, err := m.GetVersion(datasetID, editionID, versionID, models.PublishedState)
	if err != nil {
		return nil, err
	}

	var results []models.Dimension
	// To get all unique values an aggregation is needed, as using distinct() will only return the distinct values and
	// not the documents.
	// Match by instance_id
	match := bson.M{"$match": bson.M{"instance_id": version.ID}}
	// Then group the values by name.
	group := bson.M{"$group": bson.M{"_id": "$name", "doc": bson.M{"$first": "$$ROOT"}}}
	res := []bson.M{}
	err = s.DB(m.Database).C(dimensionOptions).Pipe([]bson.M{match, group}).All(&res)
	if err != nil {
		return nil, err
	}

	for _, dim := range res {
		opt := convertBSonToDimension(dim["doc"])
		dimension := models.Dimension{Name: opt.Name}
		dimension.Links.CodeList = opt.Links.CodeList
		dimension.Links.Options = models.LinkObject{ID: opt.Name, HRef: fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/dimensions/%s/options",
			m.DatasetURL, version.Links.Dataset.ID, version.Edition, versionID, opt.Name)}
		dimension.Links.Version = *version.Links.Self

		results = append(results, dimension)
	}

	return &models.DatasetDimensionResults{Items: results}, nil
}

// GetDimensionOptions returns all dimension options for a dimensions within a dataset.
func (m *Mongo) GetDimensionOptions(datasetID, editionID, versionID, dimension string) (*models.DimensionOptionResults, error) {
	s := m.Session.Copy()
	defer s.Close()

	version, err := m.GetVersion(datasetID, editionID, versionID, models.PublishedState)
	if err != nil {
		return nil, err
	}

	var values []models.PublicDimensionOption
	iter := s.DB(m.Database).C(dimensionOptions).Find(bson.M{"instance_id": version.ID, "name": dimension}).Iter()
	err = iter.All(&values)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(values); i++ {
		values[i].Links.Version = *version.Links.Self
	}

	return &models.DimensionOptionResults{Items: values}, nil
}

func convertBSonToDimension(data interface{}) *models.DimensionOption {
	var dim models.DimensionOption
	bytes, err := bson.Marshal(data)
	if err != nil {

	}

	bson.Unmarshal(bytes, &dim)

	return &dim
}
