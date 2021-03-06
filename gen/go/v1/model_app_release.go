/*
 * Vendor API V1
 *
 * Apps documentation
 *
 * API version: 1.0.0
 * Contact: info@replicated.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package swagger

import (
	"time"
)

type AppRelease struct {
	Config    string    `json:"Config,omitempty"`
	CreatedAt time.Time `json:"CreatedAt,omitempty"`
	Editable  bool      `json:"Editable,omitempty"`
	EditedAt  time.Time `json:"EditedAt,omitempty"`
	Sequence  int64     `json:"Sequence,omitempty"`
}
