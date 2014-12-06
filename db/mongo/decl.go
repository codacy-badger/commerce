// Package mongo is a DBEngine implementation of interfaces declared in
// "github.com/ottemo/foundation/db" package
package mongo

import (
	"sync"

	"github.com/ottemo/foundation/env"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// Package global variables
var (
	attributeTypes      = make(map[string]map[string]string) // cached values of collection attribute types
	attributeTypesMutex sync.RWMutex                         // syncronization for attributeTypes modification
)

// Package global constants
const (
	ConstMongoDebug = false // flag which indicates to perform log on each operation

	ConstFilterGroupStatic  = "static"  // name for static filter, ref. to AddStaticFilter(...)
	ConstFilterGroupDefault = "default" // name for default filter, ref. to by AddFilter(...)

	ConstCollectionNameColumnInfo = "collection_column_info" // collection name to hold Ottemo types of attributes

	ConstErrorModule = "db/mongo"
	ConstErrorLevel  = env.ConstErrorLevelService
)

// StructDBFilterGroup is a structure to hold information of named collection filter
type StructDBFilterGroup struct {
	Name         string
	FilterValues []bson.D
	ParentGroup  string
	OrSequence   bool
}

// DBCollection is a implementer of InterfaceDBCollection
type DBCollection struct {
	database   *mgo.Database
	collection *mgo.Collection

	subcollections []*DBCollection
	subresults     []*bson.Raw

	Name string

	FilterGroups map[string]*StructDBFilterGroup

	Sort []string

	ResultAttributes []string

	Limit  int
	Offset int
}

// DBEngine is a implementer of InterfaceDBEngine
type DBEngine struct {
	database *mgo.Database
	session  *mgo.Session

	DBName      string
	collections map[string]bool
}
