package sqlite

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	sqlite3 "github.com/mxk/go-sqlite/sqlite3"
	"github.com/ottemo/foundation/env"
)

// LoadByID loads record from DB by it's id
func (it *DBCollection) LoadByID(id string) (map[string]interface{}, error) {
	var result map[string]interface{}

	if !ConstUseUUIDids {
		it.AddFilter("_id", "=", id)
	} else {
		if idValue, err := strconv.ParseInt(id, 10, 64); err == nil {
			it.AddFilter("_id", "=", idValue)
		} else {
			it.AddFilter("_id", "=", id)
		}
	}

	err := it.Iterate(func(row map[string]interface{}) bool {
		result = row
		return false
	})

	return result, env.ErrorDispatch(err)
}

// Load loads records from DB for current collection and filter if it set
func (it *DBCollection) Load() ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	err := it.Iterate(func(row map[string]interface{}) bool {
		result = append(result, row)
		return true
	})

	return result, env.ErrorDispatch(err)
}

// Iterate applies [iterator] function to each record, stops on return false
func (it *DBCollection) Iterate(iteratorFunc func(record map[string]interface{}) bool) error {

	SQL := it.getSelectSQL()
	if ConstDebugSQL {
		env.Log("sqlite", env.ConstLogPrefixInfo, SQL)
	}

	stmt, err := connectionQuery(SQL)
	defer closeStatement(stmt)

	if err == nil {
		for ; err == nil; err = stmt.Next() {
			row := make(sqlite3.RowMap)
			if err := stmt.Scan(row); err == nil {
				it.modifyResultRow(row)

				if !iteratorFunc(row) {
					break
				}
			}
		}
	}

	if err == io.EOF {
		err = nil
	} else if err != nil {
		err = sqlError(SQL, err)
	}

	return env.ErrorDispatch(err)
}

// Distinct returns distinct values of specified attribute
func (it *DBCollection) Distinct(columnName string) ([]interface{}, error) {

	prevResultColumns := it.ResultColumns
	it.SetResultColumns(columnName)

	SQL := "SELECT DISTINCT " + it.getSQLResultColumns() + " FROM " + it.Name + it.getSQLFilters() + it.getSQLOrder() + it.Limit
	if ConstDebugSQL {
		env.Log("sqlite", env.ConstLogPrefixInfo, SQL)
	}

	it.ResultColumns = prevResultColumns

	stmt, err := connectionQuery(SQL)
	defer closeStatement(stmt)

	var result []interface{}
	if err == nil {
		for ; err == nil; err = stmt.Next() {
			row := make(sqlite3.RowMap)
			if err := stmt.Scan(row); err == nil {
				ignoreNull := false
				for _, columnValue := range row {
					if columnValue == nil {
						ignoreNull = true
					}
				}
				if ignoreNull {
					continue
				}

				it.modifyResultRow(row)

				for _, columnValue := range row {
					result = append(result, columnValue)
					break
				}
			}
		}
	}

	if err == io.EOF {
		err = nil
	} else if err != nil {
		err = sqlError(SQL, err)
	}

	return result, env.ErrorDispatch(err)
}

// Count returns count of rows matching current select statement
func (it *DBCollection) Count() (int, error) {
	sqlLoadFilter := it.getSQLFilters()

	SQL := "SELECT COUNT(*) AS cnt FROM " + it.Name + sqlLoadFilter

	if ConstDebugSQL {
		env.Log("sqlite", env.ConstLogPrefixInfo, SQL)
	}

	stmt, err := connectionQuery(SQL)
	defer closeStatement(stmt)

	if err == nil {
		row := make(sqlite3.RowMap)
		if err = stmt.Scan(row); err == nil {
			cnt := int(row["cnt"].(int64)) //TODO: check this assertion works
			return cnt, err
		}
	}

	if err == io.EOF {
		err = nil
	} else if err != nil {
		err = sqlError(SQL, err)
	}

	return 0, err
}

// Save stores record in DB for current collection
func (it *DBCollection) Save(item map[string]interface{}) (string, error) {

	// prevents saving of blank records
	if len(item) == 0 {
		return "", nil
	}

	// we should make new _id column if it was not set
	if ConstUseUUIDids {
		if idValue, present := item["_id"]; !present || idValue == nil {
			item["_id"] = it.makeUUID("")
		} else {
			if idValue, ok := idValue.(string); ok {
				item["_id"] = it.makeUUID(idValue)
			}
		}
	} else {
		// _id in SQLite supposed to be auto-incremented int but for MongoDB it forced to be string
		// collection interface also forced us to use string but we still want it ti be int in DB
		// to make that we need to convert it before save from  string to int or nil
		// and after save get auto-incremented id as convert to string
		if idValue, present := item["_id"]; present && idValue != nil {
			if idValue, ok := idValue.(string); ok {

				if intValue, err := strconv.ParseInt(idValue, 10, 64); err == nil {
					item["_id"] = intValue
				} else {
					item["_id"] = nil
				}

			} else {
				return "", env.ErrorNew("unexpected _id value '" + fmt.Sprint(item) + "'")
			}
		} else {
			item["_id"] = nil
		}
	}

	// SQL generation
	columns := make([]string, 0, len(item))
	args := make([]string, 0, len(item))
	values := make([]interface{}, 0, len(item))

	for k, v := range item {
		if item[k] != nil {
			columns = append(columns, "`"+k+"`")
			args = append(args, convertValueForSQL(v))

			//args = append(args, "$_"+k)
			//values = append(values, convertValueForSQL(v))
		}
	}

	SQL := "INSERT OR REPLACE INTO " + it.Name +
		" (" + strings.Join(columns, ",") + ") VALUES" +
		" (" + strings.Join(args, ",") + ")"

	if ConstDebugSQL {
		env.Log("sqlite", env.ConstLogPrefixInfo, SQL)
	}

	if !ConstUseUUIDids {
		newIDInt64, err := connectionExecWLastInsertID(SQL, values...)
		if err != nil {
			return "", sqlError(SQL, err)
		}

		// auto-incremented _id back to string
		newIDString := strconv.FormatInt(newIDInt64, 10)
		item["_id"] = newIDString
	} else {
		err := connectionExec(SQL, values...)
		if err != nil {
			return "", sqlError(SQL, err)
		}
	}

	return item["_id"].(string), nil
}

// Delete removes records that matches current select statement from DB
//   - returns amount of affected rows
func (it *DBCollection) Delete() (int, error) {
	sqlDeleteFilter := it.getSQLFilters()

	SQL := "DELETE FROM " + it.Name + sqlDeleteFilter

	if ConstDebugSQL {
		env.Log("sqlite", env.ConstLogPrefixInfo, SQL)
	}

	affected, err := connectionExecWAffected(SQL)

	return affected, env.ErrorDispatch(err)
}

// DeleteByID removes record from DB by is's id
func (it *DBCollection) DeleteByID(id string) error {
	SQL := "DELETE FROM " + it.Name + " WHERE _id = " + convertValueForSQL(id)

	if ConstDebugSQL {
		env.Log("sqlite", env.ConstLogPrefixInfo, SQL)
	}

	return connectionExec(SQL)
}

// SetupFilterGroup setups filter group params for collection
func (it *DBCollection) SetupFilterGroup(groupName string, orSequence bool, parentGroup string) error {
	if _, present := it.FilterGroups[parentGroup]; !present && parentGroup != "" {
		return env.ErrorNew("invalid parent group")
	}

	filterGroup := it.getFilterGroup(groupName)
	filterGroup.OrSequence = orSequence
	filterGroup.ParentGroup = parentGroup

	return nil
}

// RemoveFilterGroup removes filter group for collection
func (it *DBCollection) RemoveFilterGroup(groupName string) error {
	if _, present := it.FilterGroups[groupName]; !present {
		return env.ErrorNew("invalid group name")
	}

	delete(it.FilterGroups, groupName)
	return nil
}

// AddGroupFilter adds selection filter to specific filter group (all filter groups will be joined before db query)
func (it *DBCollection) AddGroupFilter(groupName string, columnName string, operator string, value interface{}) error {
	err := it.updateFilterGroup(groupName, columnName, operator, value)
	if err != nil {
		return err
	}

	return nil
}

// AddStaticFilter adds selection filter that will not be cleared by ClearFilters() function
func (it *DBCollection) AddStaticFilter(columnName string, operator string, value interface{}) error {

	err := it.updateFilterGroup(ConstFilterGroupStatic, columnName, operator, value)
	if err != nil {
		return err
	}

	return nil
}

// AddFilter adds selection filter to current collection(table) object
func (it *DBCollection) AddFilter(ColumnName string, Operator string, Value interface{}) error {

	err := it.updateFilterGroup(ConstFilterGroupDefault, ColumnName, Operator, Value)
	if err != nil {
		return err
	}

	return nil
}

// ClearFilters removes all filters that were set for current collection, except static
func (it *DBCollection) ClearFilters() error {
	for filterGroup := range it.FilterGroups {
		if filterGroup != ConstFilterGroupStatic {
			delete(it.FilterGroups, filterGroup)
		}
	}

	return nil
}

// AddSort adds sorting for current collection
func (it *DBCollection) AddSort(ColumnName string, Desc bool) error {
	if it.HasColumn(ColumnName) {
		if Desc {
			it.Order = append(it.Order, ColumnName+" DESC")
		} else {
			it.Order = append(it.Order, ColumnName)
		}
	} else {
		return env.ErrorNew("can't find column '" + ColumnName + "'")
	}

	return nil
}

// ClearSort removes any sorting that was set for current collection
func (it *DBCollection) ClearSort() error {
	it.Order = make([]string, 0)
	return nil
}

// SetResultColumns limits column selection for Load() and LoadByID()function
func (it *DBCollection) SetResultColumns(columns ...string) error {
	for _, columnName := range columns {
		if !it.HasColumn(columnName) {
			return env.ErrorNew("there is no column " + columnName + " found")
		}

		it.ResultColumns = append(it.ResultColumns, columnName)
	}

	return nil
}

// SetLimit results pagination
func (it *DBCollection) SetLimit(Offset int, Limit int) error {
	if Limit == 0 {
		it.Limit = ""
	} else {
		it.Limit = " LIMIT " + strconv.Itoa(Limit) + " OFFSET " + strconv.Itoa(Offset)
	}

	return nil
}

// ListColumns returns attributes(columns) available for current collection(table)
func (it *DBCollection) ListColumns() map[string]string {

	result := make(map[string]string)

	if ConstUseUUIDids {
		result["_id"] = "int"
	} else {
		result["_id"] = "varchar"
	}

	// updating column into collection
	SQL := "SELECT column, type FROM " + ConstCollectionNameColumnInfo + " WHERE collection = '" + it.Name + "'"
	stmt, err := connectionQuery(SQL)
	defer closeStatement(stmt)

	row := make(sqlite3.RowMap)
	for ; err == nil; err = stmt.Next() {
		stmt.Scan(row)

		key := row["column"].(string)
		value := row["type"].(string)

		result[key] = value
	}

	// updating cached attribute types information
	if _, present := dbEngine.attributeTypes[it.Name]; !present {
		dbEngine.attributeTypes[it.Name] = make(map[string]string)
	}

	dbEngine.attributeTypesMutex.Lock()
	for attributeName, attributeType := range result {
		dbEngine.attributeTypes[it.Name][attributeName] = attributeType
	}
	dbEngine.attributeTypesMutex.Unlock()

	return result
}

// GetColumnType returns SQL like type of attribute in current collection, or if not present ""
func (it *DBCollection) GetColumnType(columnName string) string {
	if columnName == "_id" {
		return "string"
	}

	// looking in cache first
	attributeType, present := dbEngine.attributeTypes[it.Name][columnName]
	if !present {
		// updating cache, and looking again
		it.ListColumns()
		attributeType, present = dbEngine.attributeTypes[it.Name][columnName]
	}

	return attributeType
}

// HasColumn checks attribute(column) presence in current collection
func (it *DBCollection) HasColumn(columnName string) bool {
	// looking in cache first
	_, present := dbEngine.attributeTypes[it.Name][columnName]
	if !present {
		// updating cache, and looking again
		it.ListColumns()
		_, present = dbEngine.attributeTypes[it.Name][columnName]
	}

	return present
}

// AddColumn adds new attribute(column) to current collection(table)
func (it *DBCollection) AddColumn(columnName string, columnType string, indexed bool) error {

	// checking column name
	if !ConstSQLNameValidator.MatchString(columnName) {
		return env.ErrorNew("not valid column name for DB engine: " + columnName)
	}

	// checking if column already present
	if it.HasColumn(columnName) {
		if currentType := it.GetColumnType(columnName); currentType != columnType {
			return env.ErrorNew("column '" + columnName + "' already exists with type '" + currentType + "' for '" + it.Name + "' collection. Requested type '" + columnType + "'")
		}
		return nil
	}

	// updating physical table
	//-------------------------
	ColumnType, err := GetDBType(columnType)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	SQL := "ALTER TABLE " + it.Name + " ADD COLUMN \"" + columnName + "\" " + ColumnType

	if ConstDebugSQL {
		env.Log("sqlite", env.ConstLogPrefixInfo, SQL)
	}

	err = connectionExec(SQL)
	if err != nil {
		return sqlError(SQL, err)
	}

	// updating collection info table
	//--------------------------------
	SQL = "INSERT INTO " + ConstCollectionNameColumnInfo + "(collection, column, type, indexed) VALUES (" +
		"'" + it.Name + "', " +
		"'" + columnName + "', " +
		"'" + columnType + "', "
	if indexed {
		SQL += "1)"
	} else {
		SQL += "0)"
	}

	if ConstDebugSQL {
		env.Log("sqlite", env.ConstLogPrefixInfo, SQL)
	}

	err = connectionExec(SQL)
	if err != nil {
		return sqlError(SQL, err)
	}

	it.ListColumns()

	return nil
}

// RemoveColumn removes attribute(column) to current collection(table)
//   - sqlite do not have alter DROP COLUMN statements so it is hard task...
func (it *DBCollection) RemoveColumn(columnName string) error {

	// checking column in table
	//-------------------------
	if columnName == "_id" {
		return env.ErrorNew("you can't remove _id column")
	}

	if !it.HasColumn(columnName) {
		return env.ErrorNew("column '" + columnName + "' not exists in '" + it.Name + "' collection")
	}

	// getting table create SQL to take columns from
	//----------------------------------------------
	var tableCreateSQL string

	SQL := "SELECT sql FROM sqlite_master WHERE tbl_name='" + it.Name + "' AND type='table'"
	stmt, err := connectionQuery(SQL)
	if err != nil {
		closeStatement(stmt)
		return sqlError(SQL, err)
	}

	err = stmt.Scan(&tableCreateSQL)
	if err != nil {
		closeStatement(stmt)
		return err
	}
	closeStatement(stmt)

	SQL = "DELETE FROM " + ConstCollectionNameColumnInfo + " WHERE collection='" + it.Name + "' AND column='" + columnName + "'"
	if err := connectionExec(SQL); err != nil {
		return sqlError(SQL, err)
	}

	// parsing create SQL, making same but w/o deleting column
	//--------------------------------------------------------
	tableColumnsWTypes := ""
	tableColumnsWoTypes := ""

	re := regexp.MustCompile("CREATE TABLE [^(]*\\((.*)\\)$")
	if regexMatch := re.FindStringSubmatch(tableCreateSQL); len(regexMatch) >= 2 {
		tableColumnsList := strings.Split(regexMatch[1], ", ")

		for _, tableColumn := range tableColumnsList {
			tableColumn = strings.Trim(tableColumn, "\n\t ")

			if !strings.HasPrefix(tableColumn, columnName) && !strings.HasPrefix(tableColumn, "\""+columnName+"\"") {
				if tableColumnsWTypes != "" {
					tableColumnsWTypes += ", "
					tableColumnsWoTypes += ", "
				}
				tableColumnsWTypes += tableColumn
				tableColumnsWoTypes += tableColumn[0:strings.Index(tableColumn, " ")]
			}

		}
	} else {
		return env.ErrorNew("can't find table create columns in '" + tableCreateSQL + "', found [" + strings.Join(regexMatch, ", ") + "]")
	}

	// making new table without removing column, and filling with values from old table
	//---------------------------------------------------------------------------------
	SQL = "CREATE TABLE " + it.Name + "_removecolumn (" + tableColumnsWTypes + ") "
	if err := connectionExec(SQL); err != nil {
		return sqlError(SQL, err)
	}

	SQL = "INSERT INTO " + it.Name + "_removecolumn (" + tableColumnsWoTypes + ") SELECT " + tableColumnsWoTypes + " FROM " + it.Name
	if err := connectionExec(SQL); err != nil {
		return sqlError(SQL, err)
	}

	// switching newly created table, deleting old table
	//---------------------------------------------------
	SQL = "ALTER TABLE " + it.Name + " RENAME TO " + it.Name + "_fordelete"
	if err := connectionExec(SQL); err != nil {
		return sqlError(SQL, err)
	}

	SQL = "ALTER TABLE " + it.Name + "_removecolumn RENAME TO " + it.Name
	if err := connectionExec(SQL); err != nil {
		return sqlError(SQL, err)
	}

	SQL = "DROP TABLE " + it.Name + "_fordelete"
	if err := connectionExec(SQL); err != nil {
		return sqlError(SQL, err)
	}

	it.ListColumns()

	return nil
}
