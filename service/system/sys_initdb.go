package system

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"sort"
	"zcfil-server/global"
	"zcfil-server/model/system/request"
)

const (
	Mysql           = "mysql"
	InitSuccess     = "\n[%v] --> Initial data successful!\n"
	InitDataExist   = "\n[%v] --> %v the initial data of already exists!\n"
	InitDataFailed  = "\n[%v] --> %v initial data failure! \nerr: %+v\n"
	InitDataSuccess = "\n[%v] --> %v initial data successful!\n"
)

const (
	InitOrderSystem   = 10
	InitOrderInternal = 1000
	InitOrderExternal = 100000
)

var (
	ErrMissingDBContext        = errors.New("missing db in context")
	ErrMissingDependentContext = errors.New("missing dependent value in context")
	ErrDBTypeMismatch          = errors.New("db type mismatch")
)

// SubInitializer Provide an interface for using source/*/init(), with each initializer completing an initialization process
type SubInitializer interface {
	InitializerName() string // It does not necessarily represent a separate table, so it has been changed to a broader semantics
	MigrateTable(ctx context.Context) (next context.Context, err error)
	InitializeData(ctx context.Context) (next context.Context, err error)
	TableCreated(ctx context.Context) bool
	DataInserted(ctx context.Context) bool
}

// TypedDBInitHandler Execute the incoming initializer
type TypedDBInitHandler interface {
	EnsureDB(ctx context.Context, conf *request.InitDB) (context.Context, error) // Creating a database, failure belongs to a fatal error, so make it panic
	WriteConfig(ctx context.Context) error                                       // Write back configuration
	InitTables(ctx context.Context, inits initSlice) error                       // Creating Tables handler
	InitData(ctx context.Context, inits initSlice) error                         // Building data handler
}

// orderedInitializer Combine a sequential field for sorting
type orderedInitializer struct {
	order int
	SubInitializer
}

// initSlice 供 initializer Use when sorting dependencies
type initSlice []*orderedInitializer

var (
	initializers initSlice
	cache        map[string]*orderedInitializer
)

// RegisterInit The initialization process to be executed during registration will be called in InitDB()
func RegisterInit(order int, i SubInitializer) {
	if initializers == nil {
		initializers = initSlice{}
	}
	if cache == nil {
		cache = map[string]*orderedInitializer{}
	}
	name := i.InitializerName()
	if _, existed := cache[name]; existed {
		panic(fmt.Sprintf("Name conflict on %s", name))
	}
	ni := orderedInitializer{order, i}
	initializers = append(initializers, &ni)
	cache[name] = &ni
}

/* ---- * service * ---- */

type InitDBService struct{}

// InitDB Create a database and initialize the main entrance
func (initDBService *InitDBService) InitDB(conf request.InitDB) (err error) {
	ctx := context.TODO()
	if len(initializers) == 0 {
		return errors.New("No initialization process available, please check if initialization has been completed")
	}
	sort.Sort(&initializers) // Ensure that initializers with dependencies are executed later
	//Note: If the initializer only has a single dependency, it can be written as B=A+1, C=A+1; Since there is no dependency relationship between BCs, it does not affect initialization as who comes first and who comes later
	//If there are multiple dependencies, they can be written as C=A+B, D=A+B+C, E=A+1;
	//C must be greater than A | B, so it is executed after AB, D must be greater than A | B | C, so it is executed after ABC, while E only depends on A and the order is independent of CD, so it does not affect which E or CD executes first
	var initHandler TypedDBInitHandler
	switch conf.DBType {
	case "mysql":
		initHandler = NewMysqlInitHandler()
		ctx = context.WithValue(ctx, "dbtype", "mysql")
	default:
		initHandler = NewMysqlInitHandler()
		ctx = context.WithValue(ctx, "dbtype", "mysql")
	}
	ctx, err = initHandler.EnsureDB(ctx, &conf)
	if err != nil {
		return err
	}

	db := ctx.Value("db").(*gorm.DB)
	global.ZC_DB = db

	if err = initHandler.InitTables(ctx, initializers); err != nil {
		return err
	}
	if err = initHandler.InitData(ctx, initializers); err != nil {
		return err
	}

	if err = initHandler.WriteConfig(ctx); err != nil {
		return err
	}
	initializers = initSlice{}
	cache = map[string]*orderedInitializer{}
	return nil
}

// InitDB Create a database and initialize the main entrance
func (initDBService *InitDBService) InitData(DBType string) (err error) {
	ctx := context.TODO()
	if len(initializers) == 0 {
		return errors.New("No initialization process available, please check if initialization has been completed")
	}
	sort.Sort(&initializers) // Ensure that initializers with dependencies are executed later
	//Note: If the initializer only has a single dependency, it can be written as B=A+1, C=A+1; Since there is no dependency relationship between BCs, it does not affect initialization as who comes first and who comes later
	//If there are multiple dependencies, they can be written as C=A+B, D=A+B+C, E=A+1;
	//C must be greater than A | B, so it is executed after AB, D must be greater than A | B | C, so it is executed after ABC, while E only depends on A and the order is independent of CD, so it does not affect which E or CD executes first
	var initHandler TypedDBInitHandler
	switch DBType {
	case "mysql":
		initHandler = NewMysqlInitHandler()
		ctx = context.WithValue(ctx, "dbtype", "mysql")
	default:
		initHandler = NewMysqlInitHandler()
		ctx = context.WithValue(ctx, "dbtype", "mysql")
	}
	ctx = context.WithValue(ctx, "db", global.ZC_DB)
	if err = initHandler.InitData(ctx, initializers); err != nil {
		return err
	}

	initializers = initSlice{}
	return nil
}

// createDatabase Create database (called in EnsureDB())
func createDatabase(dsn string, driver string, createSql string) error {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(db)
	if err = db.Ping(); err != nil {
		return err
	}
	_, err = db.Exec(createSql)
	return err
}

// createTables Create Table (default dbInitHandler. initTables behavior)
func createTables(ctx context.Context, inits initSlice) error {
	next, cancel := context.WithCancel(ctx)
	defer func(c func()) { c() }(cancel)
	for _, init := range inits {
		if init.TableCreated(next) {
			continue
		}
		if n, err := init.MigrateTable(next); err != nil {
			return err
		} else {
			next = n
		}

	}
	return nil
}

/* -- sortable interface -- */

func (a initSlice) Len() int {
	return len(a)
}

func (a initSlice) Less(i, j int) bool {
	return a[i].order < a[j].order
}

func (a initSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
