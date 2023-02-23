package turso

import (
	"fmt"
	"net/http"
)

type Database struct {
	ID            string `json:"dbId"`
	Name          string
	Type          string
	Region        string
	Regions       []string
	PrimaryRegion string
	Hostname      string
}

type DatabasesClient client

func (d *DatabasesClient) List() ([]Database, error) {
	r, err := d.client.Get("/v2/databases", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get database listing: %s", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get database listing: %s", r.Status)
	}

	type ListResponse struct {
		Databases []Database `json:"databases"`
	}
	resp, err := unmarshal[ListResponse](r)
	return resp.Databases, err
}

func (d *DatabasesClient) Delete(database string) error {
	url := fmt.Sprintf("/v2/databases/%s", database)
	r, err := d.client.Delete(url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete database: %s", err)
	}
	defer r.Body.Close()

	if r.StatusCode == http.StatusNotFound {
		return fmt.Errorf("database %s not found. List known databases using %s", Emph(database), Emph("turso db list"))
	}

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete database: %s", r.Status)
	}

	return nil
}

type CreateDatabaseResponse struct {
	Database Database
	Username string
	Password string
}

func (d *DatabasesClient) Create(name, region, image string) (*CreateDatabaseResponse, error) {
	type Body struct{ Name, Region, Image string }
	body, err := marshal(Body{name, region, image})
	if err != nil {
		return nil, fmt.Errorf("could not serialize request body: %w", err)
	}

	res, err := d.client.Post("/v2/databases", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnprocessableEntity {
		return nil, fmt.Errorf("Database name '%s' is not available", name)
	}

	if res.StatusCode != http.StatusOK {
		return nil, parseResponseError(res)
	}

	data, err := unmarshal[*CreateDatabaseResponse](res)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	return data, nil
}

func (d *DatabasesClient) ChangePassword(database string, newPassword string) error {
	type Body struct{ Password string }
	body, err := marshal(Body{newPassword})
	if err != nil {
		return fmt.Errorf("could not serialize request body: %w", err)
	}
	url := fmt.Sprintf("/v2/databases/%s/password", database)
	r, err := d.client.Post(url, body)
	if err != nil {
		return fmt.Errorf("failed to change database password: %s", err)
	}
	defer r.Body.Close()

	if r.StatusCode == http.StatusNotFound {
		return fmt.Errorf("database %s not found. List known databases using %s", Emph(database), Emph("turso db list"))
	}

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to change database password: %s", r.Status)
	}

	return nil
}
