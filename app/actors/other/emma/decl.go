package emma

import "github.com/ottemo/commerce/env"

// Package constants for Emma module
const (
	ConstErrorModule = "emma"
	ConstErrorLevel  = env.ConstErrorLevelAPI

	ConstConfigPathEmma                = "general.emma"
	ConstConfigPathEmmaEnabled         = "general.emma.enabled"
	ConstConfigPathEmmaPublicAPIKey    = "general.emma.public_api_key"
	ConstConfigPathEmmaPrivateAPIKey   = "general.emma.private_api_key"
	ConstConfigPathEmmaAccountID       = "general.emma.account_id"
	ConstConfigPathEmmaSKU             = "general.emma.trigger_sku"
	ConstConfigPathEmmaDefaultGroupIds = "general.emma.default_group_ids"
)

var (
	emmaService emmaServiceType
)

type emmaCredentialsGetter interface {
	get() (emmaCredentialsType, error)
}
