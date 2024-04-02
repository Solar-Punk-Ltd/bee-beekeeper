package dynamicaccess_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/file/loadsave"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/stretchr/testify/assert"
)

func TestHistoryAdd(t *testing.T) {
	h, err := dynamicaccess.NewHistory(nil)
	assert.NoError(t, err)

	addr := swarm.NewAddress([]byte("addr"))

	ctx := context.Background()

	err = h.Add(ctx, addr, nil)
	assert.NoError(t, err)
}

func TestSingleNodeHistoryLookup(t *testing.T) {
	storer := mockstorer.New()
	ctx := context.Background()
	ls := loadsave.New(storer.ChunkStore(), storer.Cache(), pipelineFactory(storer.Cache(), false, 0))

	h, err := dynamicaccess.NewHistory(ls)
	assert.NoError(t, err)

	testActRef := swarm.RandAddress(t)
	err = h.Add(ctx, testActRef, nil)
	assert.NoError(t, err)

	_, err = h.Store(ctx)
	assert.NoError(t, err)

	searchedTime := time.Now().Unix()
	actRef, err := h.Lookup(ctx, searchedTime, ls)
	assert.NoError(t, err)
	assert.True(t, actRef.Equal(testActRef))
}

func TestMultiNodeHistoryLookup(t *testing.T) {
	storer := mockstorer.New()
	ctx := context.Background()
	ls := loadsave.New(storer.ChunkStore(), storer.Cache(), pipelineFactory(storer.Cache(), false, 0))

	h, _ := dynamicaccess.NewHistory(ls)

	testActRef1 := swarm.RandAddress(t)
	firstTime := time.Date(1994, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	h.Add(ctx, testActRef1, &firstTime)

	testActRef2 := swarm.RandAddress(t)
	secondTime := time.Date(2000, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	h.Add(ctx, testActRef2, &secondTime)

	testActRef3 := swarm.RandAddress(t)
	thirdTime := time.Date(2015, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	h.Add(ctx, testActRef3, &thirdTime)

	testActRef4 := swarm.RandAddress(t)
	fourthTime := time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	h.Add(ctx, testActRef4, &fourthTime)

	testActRef5 := swarm.RandAddress(t)
	fifthTime := time.Date(2030, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	h.Add(ctx, testActRef5, &fifthTime)

	_, err := h.Store(ctx)
	assert.NoError(t, err)

	searchedTime := time.Date(2016, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	actRef, err := h.Lookup(ctx, searchedTime, ls)
	assert.NoError(t, err)
	assert.True(t, actRef.Equal(testActRef3))
}
