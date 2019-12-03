package kotsclient

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListCustomers struct {
	Data   *CustomerDataWrapper `json:"data,omitempty"`
	Errors []graphql.GQLError   `json:"errors,omitempty"`
}

type CustomerDataWrapper struct {
	Customers CustomerData `json:"customers"`
}

type CustomerData struct {
	Customers []*Customer `json:"customers"`
}

type Customer struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Channels  []*KotsChannel `json:"channels"`
	Type      string         `json:"type"`
	ExpiresAt string         `json:"expiresAt"`
}

func (c *GraphQLClient) ListCustomers(appID string) ([]types.Customer, error) {
	response := GraphQLResponseListCustomers{}

	request := graphql.Request{
		Query: `
	query customers($appId: String!, $appType: String!) {
		customers(appId: $appId, appType: $appType) {
            customers {
		        id
		        name 
				type
				expiresAt
		        channels {
		            id
		            name
		            currentVersion
		        }
            }
        }
	}
	`,

		Variables: map[string]interface{}{
			"appId":   appID,
			"appType": "kots",
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, errors.Wrap(err, "execute gql request")
	}

	customers := make([]types.Customer, 0, 0)
	for _, kotsCustomer := range response.Data.Customers.Customers {

		kotsChannels := make([]types.Channel, 0, 0)
		for _, kotsChannel := range kotsCustomer.Channels {
			channel := types.Channel{
				ID:              kotsChannel.ID,
				Name:            kotsChannel.Name,
				ReleaseLabel:    kotsChannel.CurrentVersion,
				ReleaseSequence: kotsChannel.ReleaseSequence,
			}
			kotsChannels = append(kotsChannels, channel)
		}
		customer := types.Customer{
			ID:       kotsCustomer.ID,
			Name:     kotsCustomer.Name,
			Type:     kotsCustomer.Type,
			Channels: kotsChannels,
		}


		customers = append(customers, customer)
	}

	return customers, nil
}
