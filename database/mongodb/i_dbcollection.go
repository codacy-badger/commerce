package mongodb

import (
	"strings"
	"errors"

	"labix.org/v2/mgo/bson"
)

const (
	COLUMN_INFO_COLLECTION = "collection_column_info"
)

func sqlError(SQL string, err error) error {
	return errors.New("SQL \"" + SQL + "\" error: " + err.Error())
}

//func getDBType(ColumnType string) (string, error) {
//	ColumnType = strings.ToLower(ColumnType)
//	switch ColumnType {
//	case ColumnType == "int" || ColumnType == "integer":
//		return "INTEGER", nil
//	case ColumnType == "real" || ColumnType == "float":
//		return "REAL", nil
//	case ColumnType == "string" || ColumnType == "text" || strings.Contains(ColumnType, "char"):
//		return "TEXT", nil
//	case ColumnType == "blob" || ColumnType == "struct" || ColumnType == "data":
//		return "BLOB", nil
//	case strings.Contains(ColumnType, "numeric") || strings.Contains(ColumnType, "decimal") || ColumnType == "money":
//		return "NUMERIC", nil
//	}
//
//	return "?", errors.New("Unknown type '" + ColumnType + "'")
//}

func getMongoOperator(Operator string, Value string) (string, string, error) {
	Operator = strings.ToLower(Operator)

	switch Operator {
	case "=":
		return "", Value, nil
	case ">":
		return "$gt", Value, nil
	case ">=":
		return "$gte", Value, nil
	case "<":
		return "$lt", Value, nil
	case "<=":
		return "$lte", Value, nil
	case "like":
		Value = strings.Replace("%", ".*", Value, -1)
		return "$regex", Value, nil
	}

	return "?", "?", errors.New("Unknown operator '" + Operator + "'")
}


func (it *MongoDBCollection) LoadById(id string) (map[string]interface{}, error) {
	result := make( map[string]interface{} )

	err := it.collection.FindId( id ).One(&result)

	return result, err
}

func (it *MongoDBCollection) Load() ([]map[string]interface{}, error) {
	result := make([] map[string]interface{}, 0)

	query := it.collection.Find( it.Selector )

	if len(it.Sort) > 0 {
		query.Sort(it.Sort...)
	}

	if it.Offset > 0 { query = query.Skip(it.Offset) }
	if it.Limit  > 0 { query = query.Limit(it.Limit) }

	err := query.All(&result)

	return result, err
}



func (it *MongoDBCollection) Save(Item map[string]interface{}) (string, error) {

	id := bson.NewObjectId().Hex()

	if _id, present := Item["_id"]; present {
		if _id, ok := _id.(string); ok && _id != "" {
			if bson.IsObjectIdHex(_id) {
				id = _id
			}
		}
	}

	Item["_id"] = id

	changeInfo, err := it.collection.UpsertId(id, Item)

	if changeInfo != nil && changeInfo.UpsertedId != nil {
		//id = changeInfo.UpsertedId
	}

	return id, err
}


func (it *MongoDBCollection) Delete() (int, error) {
	changeInfo, err := it.collection.RemoveAll(it.Selector)

	return changeInfo.Removed, err
}

func (it *MongoDBCollection) DeleteById(id string) error {

	return it.collection.RemoveId(id)
}


func (it *MongoDBCollection) AddFilter(ColumnName string, Operator string, Value string) error {

	newOperator, newValue, err := getMongoOperator(Operator, Value)
	if err != nil { return err }

	var filterValue interface{} = newValue
	if newOperator != "" {
		filterValue = map[string]interface{}{newOperator: newValue}
	} else {
		filterValue = newValue
	}

	it.Selector[ColumnName] = filterValue

	return nil
}

func (it *MongoDBCollection) ClearFilters() error {
	it.Selector = make( map[string]interface{} )
	return nil
}


func (it *MongoDBCollection) AddSort(ColumnName string, Desc bool) error {
	if Desc {
		it.Sort = append(it.Sort, "-" + ColumnName)
	} else {
		it.Sort = append(it.Sort, ColumnName)
	}
	return nil
}

func (it *MongoDBCollection) ClearSort() error {
	it.Sort = make( []string, 0 )
	return nil
}


func (it *MongoDBCollection) SetLimit(Offset int, Limit int) error {
	it.Limit = Limit
	it.Offset = Offset

	return nil
}

// Collection columns stuff
//--------------------------
func (it *MongoDBCollection) ListColumns() map[string]string {

	result := map[string]string{}
	
	infoCollection := it.database.C(COLUMN_INFO_COLLECTION)
	selector :=  map[string]string{"collection": it.Name}
	iter := infoCollection.Find(selector).Iter()
	
	row := map[string]string{}
	for iter.Next(&row) {
		colName, okColumn := row["column"]
		colType, okType := row["type"]
		
		if okColumn && okType {
			result[colName] = colType
		}
	}
	
	return result
}

func (it *MongoDBCollection) HasColumn(ColumnName string) bool {

	infoCollection := it.database.C(COLUMN_INFO_COLLECTION)
	selector :=  map[string]interface{} {"collection": it.Name, "column": ColumnName}
	count, _ := infoCollection.Find(selector).Count()
	
	return count > 0
}

func (it *MongoDBCollection) AddColumn(ColumnName string, ColumnType string, indexed bool) error {

	infoCollection := it.database.C(COLUMN_INFO_COLLECTION)

	selector := map[string]interface{} {"collection": it.Name, "column": ColumnName}
	data := map[string]interface{} {"collection": it.Name, "column": ColumnName, "type": ColumnType, "indexed": indexed}

	_, err := infoCollection.Upsert(selector, data)

	return err
}

func (it *MongoDBCollection) RemoveColumn(ColumnName string) error {

	infoCollection := it.database.C(COLUMN_INFO_COLLECTION)
	removeSelector := map[string]string{"collection": it.Name, "column": ColumnName}

	err := infoCollection.Remove(removeSelector)
	if err != nil { return err }

	updateSelector := map[string]interface{} { ColumnName: map[string]interface{} {"$exists": true} }
	data := map[string]interface{} { "$unset": map[string]interface{} {ColumnName: ""} }

	_, err = it.collection.UpdateAll(updateSelector, data)

	if err != nil { return err }

	return nil
}
