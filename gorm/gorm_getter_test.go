package gormLoader

import (
	"sync"
	"testing"

	"github.com/nuvi/go-dockerdb"
	"github.com/preston-wagner/go-dataloader"
	"github.com/preston-wagner/unicycle"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TestTable struct {
	ID   string `gorm:"primaryKey"`
	Col1 int
	Col2 bool
	Col3 string
}

func getId(tst TestTable) string {
	return tst.ID
}

func getCol1(tst TestTable) int {
	return tst.Col1
}

func deduplicationTester[KEY_TYPE comparable, VALUE_TYPE any](t *testing.T, getter dataloader.Getter[KEY_TYPE, VALUE_TYPE]) dataloader.Getter[KEY_TYPE, VALUE_TYPE] {
	calledKeys := unicycle.Set[KEY_TYPE]{}
	lock := &sync.RWMutex{}
	return func(keys []KEY_TYPE) (map[KEY_TYPE]VALUE_TYPE, map[KEY_TYPE]error) {
		lock.Lock()
		for _, key := range keys {
			if calledKeys.Has(key) {
				t.Error("deduplicationTester found that key", key, "was passed to a getter more than once!")
			} else {
				calledKeys.Add(key)
			}
		}
		lock.Unlock()
		return getter(keys)
	}
}

func TestGormGetter(t *testing.T) {
	container, connectURL := dockerdb.SetupSuite()
	defer dockerdb.StopContainer(container)

	db, err := gorm.Open(postgres.Open(connectURL))
	if err != nil {
		t.Fatal(err)
	}

	err = db.AutoMigrate(
		&TestTable{},
	)
	if err != nil {
		t.Fatal(err)
	}

	item1 := TestTable{
		ID:   "lorem",
		Col1: 7,
		Col2: true,
		Col3: "ipsum",
	}
	db.Create(&item1)
	item2 := TestTable{
		ID:   "dolor",
		Col1: 7,
		Col2: false,
		Col3: "sit",
	}
	db.Create(&item2)
	item3 := TestTable{
		ID:   "amet",
		Col1: 9,
		Col2: true,
		Col3: "lol",
	}
	db.Create(&item3)

	// test single item lookups, including multiple at the same time
	gormLoader := dataloader.NewDataLoader(deduplicationTester(t, GormGetter(db, "id", getId)), 2, 10)

	_, err = gormLoader.Load("n/a")
	assert.ErrorIs(t, err, dataloader.ErrMissingResponse)

	unicycle.ForEachMultithread([]TestTable{
		item1,
		item2,
		item3,
		item2,
		item1,
	}, func(item TestTable) {
		itemCopy, err := gormLoader.Load(item.ID)
		if err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, item, itemCopy)
		}
	}, 5)

	// test list lookups
	gormListLoader := dataloader.NewDataLoader(deduplicationTester(t, GormListGetter(db, "col1", getCol1)), 2, 10)

	unicycle.AwaitAll(
		unicycle.WrapInPromise(func() (bool, error) {
			items, err := gormListLoader.Load(7)
			if err != nil {
				t.Fatal(err)
			} else {
				assert.Contains(t, items, item1)
				assert.Contains(t, items, item2)
				assert.Len(t, items, 2)
			}
			return true, nil
		}),
		unicycle.WrapInPromise(func() (bool, error) {
			items, err := gormListLoader.Load(8)
			if err != nil {
				t.Fatal(err)
			} else {
				assert.Empty(t, items)
			}
			return true, nil
		}),
	)
}
