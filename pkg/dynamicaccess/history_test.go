package dynamicaccess_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ethersphere/bee/v2/pkg/dynamicaccess"
	"github.com/ethersphere/bee/v2/pkg/file/loadsave"
	"github.com/ethersphere/bee/v2/pkg/file/pipeline"
	"github.com/ethersphere/bee/v2/pkg/file/pipeline/builder"
	"github.com/ethersphere/bee/v2/pkg/storage"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/stretchr/testify/assert"
)

func TestHistoryAdd(t *testing.T) {
	h, err := dynamicaccess.NewHistory(nil)
	assert.NoError(t, err)

	addr := swarm.NewAddress([]byte("addr"))

	ctx := context.Background()

	err = h.Add(ctx, addr, nil, nil)
	assert.NoError(t, err)
}

func TestSingleNodeHistoryLookup(t *testing.T) {
	storer := mockstorer.New()
	ctx := context.Background()
	ls := loadsave.New(storer.ChunkStore(), storer.Cache(), pipelineFactory(storer.Cache(), false))

	h, err := dynamicaccess.NewHistory(ls)
	assert.NoError(t, err)

	testActRef := swarm.RandAddress(t)
	err = h.Add(ctx, testActRef, nil, nil)
	assert.NoError(t, err)

	_, err = h.Store(ctx)
	assert.NoError(t, err)

	searchedTime := time.Now().Unix()
	entry, err := h.Lookup(ctx, searchedTime)
	// actRef := entry.Reference()
	actRef := entry.Entry()
	assert.NoError(t, err)
	assert.True(t, swarm.NewAddress(actRef).Equal(testActRef))
	assert.Nil(t, entry.Metadata())
}

func TestMultiNodeHistoryLookup(t *testing.T) {
	storer := mockstorer.New()
	ctx := context.Background()
	ls := loadsave.New(storer.ChunkStore(), storer.Cache(), pipelineFactory(storer.Cache(), false))
	fmt.Printf("bagoy now: %d\n", time.Now().Unix())

	h, _ := dynamicaccess.NewHistory(ls)

	testActRef1 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd891"))
	firstTime := time.Date(1994, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	mtdt1 := map[string]string{"firstTime": "1994-04-01"}
	err := h.Add(ctx, testActRef1, &firstTime, &mtdt1)
	assert.NoError(t, err)

	testActRef2 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd892"))
	secondTime := time.Date(2000, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	mtdt2 := map[string]string{"secondTime": "2000-04-01"}
	err = h.Add(ctx, testActRef2, &secondTime, &mtdt2)
	assert.NoError(t, err)

	href1, err := h.Store(ctx)
	assert.NoError(t, err)

	h2, err := dynamicaccess.NewHistoryReference(ls, href1)
	assert.NoError(t, err)
	// se2, err := h2.Lookup(ctx, secondTime)
	// assert.True(t, se2.Reference().Equal(testActRef2))

	testActRef3 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd893"))
	thirdTime := time.Date(2015, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	mtdt3 := map[string]string{"thirdTime": "2015-04-01"}
	err = h2.Add(ctx, testActRef3, &thirdTime, &mtdt3)
	assert.NoError(t, err)

	testActRef4 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd894"))
	fourthTime := time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	mtdt4 := map[string]string{"fourthTime": "2020-04-01"}
	err = h2.Add(ctx, testActRef4, &fourthTime, &mtdt4)
	assert.NoError(t, err)

	testActRef5 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd895"))
	fifthTime := time.Date(2030, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	mtdt5 := map[string]string{"fifthTime": "2030-04-01"}
	err = h2.Add(ctx, testActRef5, &fifthTime, &mtdt5)
	assert.NoError(t, err)

	// latest
	searchedTime := time.Date(1980, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	entry, err := h2.Lookup(ctx, searchedTime)
	// actRef := entry.Reference()
	actRef := entry.Entry()
	assert.NoError(t, err)
	assert.True(t, swarm.NewAddress(actRef).Equal(testActRef1))
	assert.True(t, reflect.DeepEqual(mtdt1, entry.Metadata()))

	// before first time
	searchedTime = time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	entry, err = h2.Lookup(ctx, searchedTime)
	// actRef = entry.Reference()
	actRef = entry.Entry()
	assert.NoError(t, err)
	assert.True(t, swarm.NewAddress(actRef).Equal(testActRef4))
	assert.True(t, reflect.DeepEqual(mtdt4, entry.Metadata()))

	// same time
	searchedTime = time.Date(2000, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	entry, err = h.Lookup(ctx, searchedTime)
	// actRef = entry.Reference()
	actRef = entry.Entry()
	assert.NoError(t, err)
	assert.True(t, swarm.NewAddress(actRef).Equal(testActRef2))
	assert.True(t, reflect.DeepEqual(mtdt2, entry.Metadata()))

	// after time
	searchedTime = time.Date(2045, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	entry, err = h2.Lookup(ctx, searchedTime)
	// actRef = entry.Reference()
	actRef = entry.Entry()
	assert.NoError(t, err)
	assert.True(t, swarm.NewAddress(actRef).Equal(testActRef5))
	assert.True(t, reflect.DeepEqual(mtdt5, entry.Metadata()))
}

func TestHistoryStore(t *testing.T) {
	storer := mockstorer.New()
	ctx := context.Background()
	ls := loadsave.New(storer.ChunkStore(), storer.Cache(), pipelineFactory(storer.Cache(), false))

	h1, _ := dynamicaccess.NewHistory(ls)

	testActRef1 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd891"))
	firstTime := time.Date(1994, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	mtdt1 := map[string]string{"firstTime": "1994-04-01"}
	h1.Add(ctx, testActRef1, &firstTime, &mtdt1)

	href1, err := h1.Store(ctx)
	assert.NoError(t, err)
	t.Logf("bagoy href1: %v\n", href1)

	h2, err := dynamicaccess.NewHistoryReference(ls, href1)
	assert.NoError(t, err)

	entry1, err := h2.Lookup(ctx, firstTime)
	actRef1 := entry1.Entry()
	assert.True(t, swarm.NewAddress(actRef1).Equal(testActRef1))
	assert.True(t, reflect.DeepEqual(mtdt1, entry1.Metadata()))

	testActRef3 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd893"))
	thirdTime := time.Date(2015, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	mtdt3 := map[string]string{"thirdTime": "2015-04-01"}
	err = h2.Add(ctx, testActRef3, &thirdTime, &mtdt3)
	assert.NoError(t, err)

	h2.Add(ctx, testActRef3, &thirdTime, &mtdt3)
	href3, err := h2.Store(ctx)
	assert.NoError(t, err)
	t.Logf("bagoy href3: %v\n", href3)

	// testActRef4 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd894"))
	testActRef4 := swarm.NewAddress([]byte("ffffffffffffffffffffffffffffffff"))
	fourthTime := time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	mtdt4 := map[string]string{"fourthTime": "2020-04-01"}
	err = h2.Add(ctx, testActRef4, &fourthTime, &mtdt4)
	href4, err := h2.Store(ctx)
	assert.NoError(t, err)
	t.Logf("bagoy href4: %v\n", href4)
}

func pipelineFactory(s storage.Putter, encrypt bool) func() pipeline.Interface {
	return func() pipeline.Interface {
		return builder.NewPipelineBuilder(context.Background(), s, encrypt, 0)
	}
}
