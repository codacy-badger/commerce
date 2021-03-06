// Package subscription implements base subscription functionality
package subscription

import (
	"time"

	"github.com/ottemo/commerce/app/models/checkout"
	"github.com/ottemo/commerce/app/models/subscription"
	"github.com/ottemo/commerce/db"
	"github.com/ottemo/commerce/env"
)

// Package global constants
const (
	ConstErrorModule = "subscription"
	ConstErrorLevel  = env.ConstErrorLevelActor

	ConstCollectionNameSubscription = "subscription"

	ConstTimeDay = time.Hour * 24

	ConstSchedulerTaskName = "subscriptionProcess"
)

var (
	subscriptionProducts = make([]string, 0) // stores id's of products that should be subscriptional
	subscriptionEnabled  = false
)

// DefaultSubscription struct to hold subscription information and represent
// default implementer of InterfaceSubscription
type DefaultSubscription struct {
	id string

	VisitorID string
	OrderID   string

	items []subscription.StructSubscriptionItem

	CustomerEmail string
	CustomerName  string

	Status     string
	State      string
	ActionDate time.Time
	Period     int

	ShippingAddress map[string]interface{}
	BillingAddress  map[string]interface{}

	ShippingMethodCode string

	ShippingRate checkout.StructShippingRate

	// should be stored credit card info with payment method in it
	PaymentInstrument map[string]interface{}

	LastSubmit time.Time

	CreatedAt time.Time
	UpdatedAt time.Time

	Info map[string]interface{}
}

// DefaultSubscriptionCollection is a default implementer of InterfaceSubscriptionCollection
type DefaultSubscriptionCollection struct {
	listCollection     db.InterfaceDBCollection
	listExtraAtributes []string
}
