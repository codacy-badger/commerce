package default_category

import (
	"errors"
	"github.com/ottemo/foundation/models"
	"github.com/ottemo/foundation/database"

	"github.com/ottemo/foundation/api"
)

func init(){
	instance := new(DefaultCategory)

	models.RegisterModel("Visitor", instance )
	database.RegisterOnDatabaseStart( instance.setupModel )

	api.RegisterOnRestServiceStart( instance.setupAPI )
}


func (it *DefaultCategory) setupModel() error {

	if dbEngine := database.GetDBEngine(); dbEngine != nil {
		collection, err := dbEngine.GetCollection( CATEGORY_COLLECTION_NAME )
		if err != nil { return err }

		collection.AddColumn("parent_id", "id", true)
		collection.AddColumn("name", "text", true)


		collection, err = dbEngine.GetCollection( CATEGORY_PRODUCT_JUNCTION_COLLECTION_NAME )
		if err != nil { return err }

		collection.AddColumn("category_id", "id", true)
		collection.AddColumn("product_id",  "id", true)

	} else {
		return errors.New("Can't get database engine")
	}

	return nil
}


func (it *DefaultCategory) setupAPI() error {

	return nil
}
