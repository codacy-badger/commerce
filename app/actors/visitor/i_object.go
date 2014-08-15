package visitor

import (
	"errors"
	"strings"

	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/app/utils"
)

func (it *DefaultVisitor) Get(attribute string) interface{} {
	switch strings.ToLower(attribute) {
	case "_id", "id":
		return it.id
	case "fname", "first_name":
		return it.FirstName
	case "lname", "last_name":
		return it.LastName
	case "billing_address_id":
		return it.BillingAddress.GetId()
	case "shipping_address_id":
		return it.ShippingAddress.GetId()
	case "billing_address":
		return it.BillingAddress
	case "shipping_address":
		return it.ShippingAddress
	case "validate":
		return it.ValidateKey
	case "facebook_id":
		return it.FacebookId
	case "google_id":
		return it.GoogleId
	}

	return nil
}

func (it *DefaultVisitor) Set(attribute string, value interface{}) error {
	attribute = strings.ToLower(attribute)

	switch attribute {
	case "_id", "id":
		it.id = value.(string)
	case "email", "e_mail", "e-mail":
		it.Email = value.(string)
	case "fname", "first_name":
		it.FirstName = value.(string)
	case "lname", "last_name":
		it.LastName = value.(string)
	case "password", "passwd":
		it.SetPassword(value.(string))
	case "validate":
		it.ValidateKey = value.(string)
	case "facebook_id":
		it.FacebookId = value.(string)
	case "google_id":
		it.GoogleId = value.(string)

	// only address id coming - trying to get it from DB
	case "billing_address_id", "shipping_address_id":
		address := it.getVisitorAddressById(utils.InterfaceToString(value))
		if address != nil && address.GetId() != "" {

			if attribute == "billing_address_id" {
				it.BillingAddress = address
			} else {
				it.ShippingAddress = address
			}

		}

	// address with details coming
	case "billing_address", "shipping_address":
		switch value := value.(type) {

		// we have already have structure
		case visitor.I_VisitorAddress:
			if attribute == "billing_address" {
				it.BillingAddress = value
			} else {
				it.ShippingAddress = value
			}

		// we have sub-map, supposedly I_VisitorAddress capable
		case map[string]interface{}:
			model, err := models.GetModel("VisitorAddress")
			if err != nil {
				return err
			}

			if address, ok := model.(visitor.I_VisitorAddress); ok {
				err := address.FromHashMap(value)
				if err != nil {
					return err
				}

				if attribute == "billing_address" {
					it.BillingAddress = address
				} else {
					it.ShippingAddress = address
				}
			} else {
				errors.New("unsupported visitor addres model " + model.GetImplementationName())
			}

		default:
			return errors.New("unsupported 'billing_address' value")
		}
	}
	return nil
}

func (it *DefaultVisitor) FromHashMap(input map[string]interface{}) error {

	for attribute, value := range input {
		if err := it.Set(attribute, value); err != nil {
			return err
		}
	}

	return nil
}

func (it *DefaultVisitor) ToHashMap() map[string]interface{} {

	result := make(map[string]interface{})

	result["_id"] = it.id

	result["email"] = it.Email
	result["first_name"] = it.FirstName
	result["last_name"] = it.LastName

	result["billing_address"] = nil
	result["shipping_address"] = nil

	//result["billing_address_id"] = it.BillingAddressId
	//result["shipping_address_id"] = it.ShippingAddressId

	if it.BillingAddress != nil {
		result["billing_address"] = it.BillingAddress.ToHashMap()
	}

	if it.ShippingAddress != nil {
		result["shipping_address"] = it.ShippingAddress.ToHashMap()
	}

	return result
}

func (it *DefaultVisitor) GetAttributesInfo() []models.T_AttributeInfo {

	info := []models.T_AttributeInfo{
		models.T_AttributeInfo{
			Model:      "Visitor",
			Collection: "Visitor",
			Attribute:  "_id",
			Type:       "id",
			IsRequired: false,
			IsStatic:   true,
			Label:      "ID",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Visitor",
			Collection: "Visitor",
			Attribute:  "email",
			Type:       "text",
			IsRequired: true,
			IsStatic:   true,
			Label:      "E-mail",
			Group:      "General",
			Editors:    "line_text",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Visitor",
			Collection: "Visitor",
			Attribute:  "first_name",
			Type:       "text",
			IsRequired: true,
			IsStatic:   true,
			Label:      "First Name",
			Group:      "General",
			Editors:    "line_text",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Visitor",
			Collection: "Visitor",
			Attribute:  "last_name",
			Type:       "text",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Last Name",
			Group:      "General",
			Editors:    "line_text",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Visitor",
			Collection: "Visitor",
			Attribute:  "password",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Password",
			Group:      "Password",
			Editors:    "password",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Visitor",
			Collection: "Visitor",
			Attribute:  "billing_address_id",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Billing Address",
			Group:      "General",
			Editors:    "model_selector",
			Options:    "model=VisitorAddress",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Visitor",
			Collection: "Visitor",
			Attribute:  "shipping_address_id",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Shipping Address",
			Group:      "General",
			Editors:    "model_selector",
			Options:    "model:VisitorAddress",
			Default:    "",
		},
	}

	return info
}