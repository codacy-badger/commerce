package product

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/utils"
)

func (it *DefaultProduct) Get(attribute string) interface{} {
	switch strings.ToLower(attribute) {
	case "_id", "id":
		return it.id
	case "sku":
		return it.Sku
	case "name":
		return it.Name
	case "short_description":
		return it.ShortDescription
	case "description":
		return it.Description
	case "default_image", "defaultimage":
		return it.DefaultImage
	case "price":
		return it.Price
	case "weight":
		return it.Weight
	case "size":
		return it.Size
	default:
		return it.CustomAttributes.Get(attribute)
	}

	return nil
}

func (it *DefaultProduct) Set(attribute string, value interface{}) error {
	lowerCaseAttribute := strings.ToLower(attribute)
	switch lowerCaseAttribute {
	case "_id", "id":
		it.id = value.(string)
	case "sku":
		it.Sku = value.(string)
	case "name":
		it.Name = value.(string)
	case "short_description":
		it.ShortDescription = value.(string)
	case "description":
		it.Description = value.(string)
	case "default_image", "defaultimage":
		it.DefaultImage = value.(string)

	case "price", "weight", "size":
		switch value := value.(type) {
		case float64:
			switch lowerCaseAttribute {
			case "price":
				it.Price = value
			case "weight":
				it.Weight = value
			case "size":
				it.Size = value
			}
		case string:
			newValue, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}

			switch lowerCaseAttribute {
			case "price":
				it.Price = newValue
			case "weight":
				it.Weight = newValue
			case "size":
				it.Size = newValue
			}
		default:
			return errors.New("wrong " + lowerCaseAttribute + " format")
		}

	default:
		err := it.CustomAttributes.Set(attribute, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (it *DefaultProduct) FromHashMap(input map[string]interface{}) error {

	if value, ok := input["_id"]; ok {
		if value, ok := value.(string); ok {
			it.id = value
		}
	}
	if value, ok := input["sku"]; ok {
		if value, ok := value.(string); ok {
			it.Sku = value
		}
	}
	if value, ok := input["name"]; ok {
		if value, ok := value.(string); ok {
			it.Name = value
		}
	}
	if value, ok := input["short_description"]; ok {
		if value, ok := value.(string); ok {
			it.ShortDescription = value
		}
	}
	if value, ok := input["description"]; ok {
		if value, ok := value.(string); ok {
			it.Description = value
		}
	}
	if value, ok := input["default_image"]; ok {
		if value, ok := value.(string); ok {
			it.DefaultImage = value
		}
	}
	if value, ok := input["price"]; ok {
		it.Price = utils.InterfaceToFloat64(value)
	}
	if value, ok := input["weight"]; ok {
		it.Weight = utils.InterfaceToFloat64(value)
	}
	if value, ok := input["size"]; ok {
		it.Size = utils.InterfaceToFloat64(value)
	}

	it.CustomAttributes.FromHashMap(input)

	return nil
}

func (it *DefaultProduct) ToHashMap() map[string]interface{} {
	result := it.CustomAttributes.ToHashMap()

	result["_id"] = it.id
	result["sku"] = it.Sku
	result["name"] = it.Name

	result["short_description"] = it.ShortDescription
	result["description"] = it.Description

	result["default_image"] = it.DefaultImage

	result["price"] = it.Price
	result["weight"] = it.Weight
	result["size"] = it.Size

	return result
}

func (it *DefaultProduct) GetAttributesInfo() []models.T_AttributeInfo {
	result := []models.T_AttributeInfo{
		models.T_AttributeInfo{
			Model:      "Product",
			Collection: "product",
			Attribute:  "_id",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "ID",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Product",
			Collection: "product",
			Attribute:  "sku",
			Type:       "text",
			IsRequired: true,
			IsStatic:   true,
			Label:      "SKU",
			Group:      "General",
			Editors:    "line_text",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Product",
			Collection: "product",
			Attribute:  "name",
			Type:       "text",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Name",
			Group:      "General",
			Editors:    "line_text",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Product",
			Collection: "product",
			Attribute:  "short_description",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Short Description",
			Group:      "General",
			Editors:    "multiline_text",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Product",
			Collection: "product",
			Attribute:  "description",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Description",
			Group:      "General",
			Editors:    "multiline_text",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Product",
			Collection: "product",
			Attribute:  "default_image",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "DefaultImage",
			Group:      "Pictures",
			Editors:    "image_selector",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Product",
			Collection: "product",
			Attribute:  "price",
			Type:       "numeric",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Price",
			Group:      "Prices",
			Editors:    "price",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Product",
			Collection: "product",
			Attribute:  "weight",
			Type:       "numeric",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Weight",
			Group:      "Measures",
			Editors:    "numeric",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Product",
			Collection: "product",
			Attribute:  "size",
			Type:       "numeric",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Size",
			Group:      "Measures",
			Editors:    "numeric",
			Options:    "",
			Default:    "",
		},
	}

	dynamicInfo := it.CustomAttributes.GetAttributesInfo()

	for _, dynamicAttribute := range dynamicInfo {
		result = append(result, dynamicAttribute)
	}

	return result
}