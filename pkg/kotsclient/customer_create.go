package kotsclient

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
	"time"
)

type GraphQLResponseCreateCustomer struct {
	Data   *CreateCustomerData `json:"data,omitempty"`
	Errors []graphql.GQLError  `json:"errors,omitempty"`
}

type CreateCustomerData struct {
	Customer *Customer `json:"createKotsCustomer"`
}

func (c *GraphQLClient) CreateCustomer(name, channel string, expiresIn time.Duration) (*types.Customer, error) {

	response := GraphQLResponseCreateCustomer{}

	request := graphql.Request{
		Query: `
	mutation createKotsCustomer($name: String!, $channelId: String!, $expiresAt: String) {
		createKotsCustomer(name: $name, channelId: $channelId, isKots: true, airgap: false, type: "dev", entitlements: [], expiresAt: $expiresAt) {
			id
			name 
			type
			expiresAt
			# cant fetch channels here. Even though its in the schema, it always comes back nil
        }
	}
	`,

		Variables: map[string]interface{}{
			"name":      name,
			"channelId": channel,
		},
	}
	if expiresIn > 0 {
		request.Variables["expiresAt"] = (time.Now().Add(expiresIn)).Format(time.RFC3339)
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, errors.Wrap(err, "execute gql request")
	}

	customer := &types.Customer{
		ID:      response.Data.Customer.ID,
		Name:    response.Data.Customer.Name,
		Type:    response.Data.Customer.Type,
	}
	if response.Data.Customer.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, response.Data.Customer.ExpiresAt)
		if err != nil {
			return nil, errors.Wrapf(err, "parse expiresAt timestamp %q", response.Data.Customer.ExpiresAt)
		}
		customer.Expires = &parsed
	}

	return customer, nil
}
