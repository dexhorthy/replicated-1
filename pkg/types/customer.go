package types

import (
	"github.com/pkg/errors"
	"time"
)

type Customer struct {
	ID       string
	Name     string
	Channels []Channel
	Type     string
	Expires  *time.Time
}


func (c Customer) WithExpiryTime(expiryTime string) (Customer, error) {
	customer := Customer{
		ID:       c.ID,
		Name:     c.Name,
		Channels: c.Channels,
		Type:     c.Type,
	}

	if expiryTime != "" {
		parsed, err := time.Parse(time.RFC3339, expiryTime)
		if err != nil {
			return Customer{}, errors.Wrapf(err, "parse expiresAt timestamp %q", expiryTime)
		}
		customer.Expires = &parsed
	}
	return customer, nil
}
