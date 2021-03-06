// Package product is a implementation of interfaces declared in
// "github.com/ottemo/commerce/app/models/product" package
package product

import (
	"github.com/ottemo/commerce/app/helpers/attributes"
	"github.com/ottemo/commerce/db"
	"github.com/ottemo/commerce/env"
)

// Package global constants
const (
	ConstCollectionNameProduct = "product"

	ConstErrorModule = "product"
	ConstErrorLevel  = env.ConstErrorLevelActor

	ConstProductMediaTypeImage = "image"

	ConstSwatchImageDefaultFormat    = "jpeg"
	ConstSwatchImageDefaultExtention = "jpeg"
)

// DefaultProduct is a default implementer of InterfaceProduct
type DefaultProduct struct {
	id string

	Enabled bool

	Sku  string
	Name string

	ShortDescription string
	Description      string

	DefaultImage string

	Price float64

	Weight float64

	Options map[string]interface{}

	RelatedProductIds []string

	Visible bool

	// appliedOptions tracks options were applied to current instance
	appliedOptions map[string]interface{}

	// updatedQty holds qty should be updated during save operation ("" item holds qty value)
	updatedQty []map[string]interface{}

	customAttributes   *attributes.ModelCustomAttributes
	externalAttributes *attributes.ModelExternalAttributes
}

// DefaultProductCollection is a default implementer of InterfaceProduct
type DefaultProductCollection struct {
	listCollection     db.InterfaceDBCollection
	listExtraAtributes []string
}
