package product

import (
	"errors"
	"strings"
	"net/http"
	"mime"

	"github.com/ottemo/foundation/api"

	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/product"
)

func (it *DefaultProduct) setupAPI() error {

	var err error = nil

	err = api.GetRestService().RegisterAPI("product", "GET", "list", it.ListProductsRestAPI )
	if err != nil { return err }
	err = api.GetRestService().RegisterAPI("product", "GET", "get/:id", it.GetProductRestAPI )
	if err != nil { return err }
	err = api.GetRestService().RegisterAPI("product", "POST", "create", it.CreateProductRestAPI )
	if err != nil { return err }
	err = api.GetRestService().RegisterAPI("product", "PUT", "update/:id", it.UpdateProductRestAPI )
	if err != nil { return err }
	err = api.GetRestService().RegisterAPI("product", "DELETE", "delete/:id", it.DeleteProductRestAPI )
	if err != nil { return err }


	err = api.GetRestService().RegisterAPI("product", "GET", "attribute/list", it.ListProductAttributesRestAPI )
	if err != nil { return err }
	err = api.GetRestService().RegisterAPI("product", "DELETE", "attribute/remove/:attribute", it.RemoveProductAttributeRestAPI )
	if err != nil { return err }
	err = api.GetRestService().RegisterAPI("product", "POST", "attribute/add", it.AddProductAttributeRestAPI )
	if err != nil { return err }


	err = api.GetRestService().RegisterAPI("product", "GET", "media/get/:productId/:mediaType/:mediaName", it.MediaGetRestAPI )
	if err != nil { return err }
	err = api.GetRestService().RegisterAPI("product", "GET", "media/list/:productId/:mediaType", it.MediaListRestAPI )
	if err != nil { return err }
	err = api.GetRestService().RegisterAPI("product", "GET", "media/path/:productId/:mediaType", it.MediaPathRestAPI )
	if err != nil { return err }
	err = api.GetRestService().RegisterAPI("product", "POST", "media/add/:productId/:mediaType/:mediaName", it.MediaAddRestAPI )
	if err != nil { return err }
	err = api.GetRestService().RegisterAPI("product", "DELETE", "media/remove/:productId/:mediaType/:mediaName", it.MediaRemoveRestAPI )
	if err != nil { return err }


	return nil
}


// WEB REST API function used to obtain product attributes information
func (it *DefaultProduct) ListProductAttributesRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {
	model, err := models.GetModel("Product")
	if err != nil { return nil, err }

	prod, isObject := model.(models.I_Object)
	if !isObject { return nil, errors.New("product model is not I_Object compatible")  }

	attrInfo := prod.GetAttributesInfo()
	return attrInfo, nil
}



// WEB REST API function used to add new one custom attribute
func (it *DefaultProduct) AddProductAttributeRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	reqData, ok := reqContent.(map[string]interface{})
	if !ok { return nil, errors.New("unexpected request content") }

	// check request params
	//---------------------
	attributeName, isSpecified := reqParams["attribute"]
	if !isSpecified { return nil, errors.New("attribute name was not specified") }

	attributeLabel, isSpecified := reqParams["label"]
	if !isSpecified { return nil, errors.New("attribute name was not specified") }

	// processing
	//-----------
	model, err := models.GetModel("Product")
	if err != nil { return nil, err }

	attribute := models.T_AttributeInfo {
		Model:      "product",
		Collection: "product",
		Attribute:  attributeName,
		Type:       "text",
		IsRequired: false,
		IsStatic:   false,
		Label:      attributeLabel,
		Group:      "General",
		Editors:    "text",
		Options:    "",
		Default:    "",
	}

	for key, value := range reqData {
		switch value := value.(type) {
		case string:
			switch strings.ToLower(key) {
			case "type":
				attribute.Type = value
			case "required":
				if strings.ToLower(value) == "true" {
					attribute.IsRequired = true
				}
			case "group":
				attribute.Group = value
			case "editors":
				attribute.Editors = value
			case "options":
				attribute.Options = value
			case "default":
				attribute.Default = value
			}
		case bool:
			switch key {
			case "required":
				attribute.IsRequired = value
			}
		}
	}

	if prod, ok := model.(models.I_CustomAttributes); ok {
		if err := prod.AddNewAttribute(attribute); err != nil {
			return nil, errors.New("Product new attribute error: " + err.Error())
		}
	} else {
		return nil, errors.New("product model is not I_CustomAttributes")
	}

	return attribute, nil
}



// WEB REST API function used to remove custom attribute of product
func (it *DefaultProduct) RemoveProductAttributeRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	// check request params
	//--------------------
	attributeName, isSpecified := reqParams["attribute"]
	if !isSpecified { return nil, errors.New("attribute name was not specified") }

	// remove attribute actions
	//-------------------------
	model, err := models.GetModel("Product")
	if err != nil { return nil, err }

	customable, ok := model.(models.I_CustomAttributes)
	if !ok { return nil, errors.New("product model is not I_CustomAttributes compatible") }

	err = customable.RemoveAttribute(attributeName)
	if err != nil { return nil, err }

	return "ok", nil
}



// WEB REST API function used to obtain all product attributes
//   - product id must be specified in request URI "http://[site:port]/product/get/:id"
func (it *DefaultProduct) GetProductRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	// check request params
	//---------------------
	productId, isSpecifiedId := reqParams["id"]
	if !isSpecifiedId {
		return nil,  errors.New("product 'id' was not specified")
	}

	// load product operation
	//-----------------------
	if model, err := models.GetModel("Product"); err == nil {
		if model, ok := model.(product.I_Product); ok {

			err = model.Load( productId )
			if err != nil { return nil, err }

			return model.ToHashMap(), nil
		}
	}

	return nil, errors.New("Something went wrong...")
}



// WEB REST API function used to obtain product list we have in database
//   - only [_id, sku, name] attributes returns by default
func (it *DefaultProduct) ListProductsRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	result := make( []map[string]interface{}, 0 )
	if model, err := models.GetModel("Product"); err == nil {
		if model, ok := model.(product.I_Product); ok {

			productsList, err := model.List()
			if err != nil { return nil, err }

			for _, listValue := range productsList {
				if productItem, ok := listValue.(product.I_Product); ok {

					imagePath, _ := productItem.GetMediaPath("image")

					resultItem := map[string]interface{} {
						"_id":   productItem.GetId(),
						"sku":   productItem.GetSku(),
						"name":  productItem.GetName(),
						"description":   productItem.GetDescription(),
						"default_image": productItem.GetDefaultImage(),
						"image_path":    imagePath,
					}

					defaultImage := productItem.GetDefaultImage()
					if defaultImage != "" {
						path, err := productItem.GetMediaPath("image")
						if err != nil { return nil, err }

						resultItem["image"] = path + defaultImage
					}

					result = append(result, resultItem)
				}
			}

			return result, nil
		}
	}

	return nil, errors.New("Something went wrong...")
}



// WEB REST API used to create new one product
//   - product attributes must be included in POST form
//   - sku and name attributes required
func (it *DefaultProduct) CreateProductRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	// check request params
	//---------------------
	reqData, ok := reqContent.(map[string]interface{})
	if !ok { return nil, errors.New("unexpected request content") }

	if reqData["sku"] == "" || reqData["name"] == "" {
		return nil,  errors.New("product 'name' and/or 'sku' was not specified")
	}

	// create product operation
	//-------------------------
	if model, err := models.GetModel("Product"); err == nil {
		if model, ok := model.(product.I_Product); ok {

			for attribute, value := range reqData {
				err := model.Set(attribute, value)
				if err != nil { return nil, err }
			}

			err := model.Save()
			if err != nil { return nil, err }

			return model.ToHashMap(), nil
		}
	}

	return nil, errors.New("Something went wrong...")
}



// WEB REST API used to delete product
//   - product attributes must be included in POST form
func (it *DefaultProduct) DeleteProductRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	// check request params
	//--------------------
	productId, isSpecifiedId := reqParams["id"]
	if !isSpecifiedId { return nil, errors.New("product 'id' was not specified") }

	model, err := models.GetModel("Product")
	if err != nil { return nil, err }

	productModel, ok := model.(product.I_Product)
	if !ok { return nil, errors.New("product type is not I_Product campatible") }

	// delete operation
	//-----------------
	err = productModel.Delete( productId )
	if err != nil { return nil, err }

	return "ok", nil
}



// WEB REST API used to update existing product
//   - product id must be specified in request URI
//   - product attributes must be included in POST form
func (it *DefaultProduct) UpdateProductRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	// check request params
	//---------------------
	productId, isSpecifiedId := reqParams["id"]
	if !isSpecifiedId { return nil, errors.New("product 'id' was not specified") }

	reqData, ok := reqContent.(map[string]interface{})
	if !ok { return nil, errors.New("unexpected request content") }

	// update operations
	//------------------
	model, err := models.GetModel("Product")
	if err != nil { return nil, err }

	productModel, ok := model.(product.I_Product)
	if !ok { return nil, errors.New("product type is not I_Product campatible") }

	err = productModel.Load( productId )
	if err != nil { return nil, err }

	for attrName, attrVal := range reqData {
		err = productModel.Set(attrName, attrVal)
		if err != nil { return nil, err }
	}

	err = productModel.Save()
	if err != nil { return nil, err }

	return productModel.ToHashMap(), nil
}



// WEB REST API used to add media for a product
//   - product id, media type must be specified in request URI
func (it *DefaultProduct) MediaPathRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	// check request params
	//---------------------
	productId, isIdSpecified := reqParams["productId"]
	if !isIdSpecified { return nil, errors.New("product id was not specified") }

	mediaType, isTypeSpecified := reqParams["mediaType"]
	if !isTypeSpecified { return nil, errors.New("media type was not specified") }

	// list media operation
	//---------------------
	model, err := models.GetModel("Product")
	if err != nil { return nil, err }

	productModel, ok := model.(product.I_Product)
	if !ok { return nil, errors.New("product type is not I_Product campatible") }

	productModel.SetId(productId)
	mediaList, err := productModel.GetMediaPath(mediaType)
	if err != nil { return nil, err }

	return mediaList, nil
}


// WEB REST API used to add media for a product
//   - product id, media type must be specified in request URI
func (it *DefaultProduct) MediaListRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	// check request params
	//---------------------
	productId, isIdSpecified := reqParams["productId"]
	if !isIdSpecified { return nil, errors.New("product id was not specified") }

	mediaType, isTypeSpecified := reqParams["mediaType"]
	if !isTypeSpecified { return nil, errors.New("media type was not specified") }

	// list media operation
	//---------------------
	model, err := models.GetModel("Product")
	if err != nil { return nil, err }

	productModel, ok := model.(product.I_Product)
	if !ok { return nil, errors.New("product type is not I_Product campatible") }

	productModel.SetId(productId)
	mediaList, err := productModel.ListMedia(mediaType)
	if err != nil { return nil, err }

	return mediaList, nil
}



// WEB REST API used to add media for a product
//   - product id, media type and media name must be specified in request URI
//   - media contents must be included as file in POST form
func (it *DefaultProduct) MediaAddRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	// check request params
	//---------------------
	productId, isIdSpecified := reqParams["productId"]
	if !isIdSpecified { return nil, errors.New("product id was not specified") }

	mediaType, isTypeSpecified := reqParams["mediaType"]
	if !isTypeSpecified { return nil, errors.New("media type was not specified") }

	mediaName, isNameSpecified := reqParams["mediaName"]
	if !isNameSpecified { return nil, errors.New("media name was not specified") }

	// income file processing
	//-----------------------
	file, _, err := req.FormFile("file")
	if err != nil { return nil, err }

	fileSize, _ := file.Seek(0, 2)
	fileContents := make([]byte, fileSize)

	file.Seek(0, 0)
	file.Read(fileContents)

	// add media operation
	//--------------------
	model, err := models.GetModel("Product")
	if err != nil { return nil, err }

	productModel, ok := model.(product.I_Product)
	if !ok { return nil, errors.New("product type is not I_Product campatible") }

	productModel.SetId(productId)
	err = productModel.AddMedia(mediaType, mediaName, fileContents)
	if err != nil { return nil, err }

	return "ok", nil
}



// WEB REST API used to add media for a product
//   - product id, media type and media name must be specified in request URI
func (it *DefaultProduct) MediaRemoveRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	// check request params
	//---------------------
	productId, isIdSpecified := reqParams["productId"]
	if !isIdSpecified { return nil, errors.New("product id was not specified") }

	mediaType, isTypeSpecified := reqParams["mediaType"]
	if !isTypeSpecified { return nil, errors.New("media type was not specified") }

	mediaName, isNameSpecified := reqParams["mediaName"]
	if !isNameSpecified { return nil, errors.New("media name was not specified") }

	// list media operation
	//---------------------
	model, err := models.GetModel("Product")
	if err != nil { return nil, err }

	productModel, ok := model.(product.I_Product)
	if !ok { return nil, errors.New("product type is not I_Product campatible") }

	productModel.SetId(productId)
	err = productModel.RemoveMedia(mediaType, mediaName)
	if err != nil { return nil, err }

	return "ok", nil
}



// WEB REST API used to get media contents for a product
//   - product id, media type and media name must be specified in request URI
func (it *DefaultProduct) MediaGetRestAPI(resp http.ResponseWriter, req *http.Request, reqParams map[string]string, reqContent interface{}) (interface{}, error) {

	// check request params
	//---------------------
	productId, isIdSpecified := reqParams["productId"]
	if !isIdSpecified { return nil, errors.New("product id was not specified") }

	mediaType, isTypeSpecified := reqParams["mediaType"]
	if !isTypeSpecified { return nil, errors.New("media type was not specified") }

	mediaName, isNameSpecified := reqParams["mediaName"]
	if !isNameSpecified { return nil, errors.New("media name was not specified") }


	resp.Header().Set("Content-Type", mime.TypeByExtension(mediaName) )

	// list media operation
	//---------------------
	model, err := models.GetModel("Product")
	if err != nil { return nil, err }

	productModel, ok := model.(product.I_Product)
	if !ok { return nil, errors.New("product type is not I_Product campatible") }

	productModel.SetId(productId)

	return productModel.GetMedia(mediaType, mediaName)
}

