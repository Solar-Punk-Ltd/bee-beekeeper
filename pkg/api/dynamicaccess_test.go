// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethersphere/bee/v2/pkg/api"
	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/ethersphere/bee/v2/pkg/dynamicaccess"
	mockdac "github.com/ethersphere/bee/v2/pkg/dynamicaccess/mock"
	"github.com/ethersphere/bee/v2/pkg/file/loadsave"
	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	"github.com/ethersphere/bee/v2/pkg/jsonhttp"
	"github.com/ethersphere/bee/v2/pkg/jsonhttp/jsonhttptest"
	"github.com/ethersphere/bee/v2/pkg/log"
	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

func prepareHistoryFixture(storer api.Storer) (dynamicaccess.History, swarm.Address) {
	ctx := context.Background()
	ls := loadsave.New(storer.ChunkStore(), storer.Cache(), pipelineFactory(storer.Cache(), false, redundancy.NONE))

	h, _ := dynamicaccess.NewHistory(ls, nil)

	testActRef1 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd891"))
	firstTime := time.Date(1994, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	h.Add(ctx, testActRef1, &firstTime)

	testActRef2 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd892"))
	secondTime := time.Date(2000, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	h.Add(ctx, testActRef2, &secondTime)

	testActRef3 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd893"))
	thirdTime := time.Date(2015, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	h.Add(ctx, testActRef3, &thirdTime)

	testActRef4 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd894"))
	fourthTime := time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	h.Add(ctx, testActRef4, &fourthTime)

	testActRef5 := swarm.NewAddress([]byte("39a5ea87b141fe44aa609c3327ecd895"))
	fifthTime := time.Date(2030, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	h.Add(ctx, testActRef5, &fifthTime)

	ref, _ := h.Store(ctx)
	return h, ref
}

// TODO: maybe add HEAD tests
// nolint:paralleltest,tparallel
// TODO: TestDacWithoutActHeader doc. comment
func TestDacWithoutAct(t *testing.T) {
	t.Parallel()
	var (
		data, _              = hex.DecodeString("a786dd84b61485de12146fd9c4c02d87e8fd95f0542765cb7fc3d2e428c0bcfa")
		pk, _                = crypto.DecodeSecp256k1PrivateKey(data)
		publicKeyBytes       = crypto.EncodeSecp256k1PublicKey(&pk.PublicKey)
		publisher            = hex.EncodeToString(publicKeyBytes)
		fileUploadResource   = "/bzz"
		fileDownloadResource = func(addr string) string { return "/bzz/" + addr }
		storerMock           = mockstorer.New()
		h, fixtureHref       = prepareHistoryFixture(storerMock)
		logger               = log.Noop
		fileName             = "sample.html"
		now                  = time.Now().Unix()
	)

	t.Run("upload-w/-act-then-download-w/o-act", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(mockdac.WithHistory(h, fixtureHref.String())),
		})
		var (
			testfile     = "testfile1"
			encryptedRef = "a5df670544eaea29e61b19d8739faa4573b19e4426e58a173e51ed0b5e7e2ade"
		)
		jsonhttptest.Request(t, client, http.MethodPost, fileUploadResource+"?name="+fileName, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmActHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(strings.NewReader(testfile)),
			jsonhttptest.WithExpectedJSONResponse(api.BzzUploadResponse{
				Reference: swarm.MustParseHexAddress(encryptedRef),
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithNonEmptyResponseHeader(api.SwarmTagHeader),
			jsonhttptest.WithExpectedResponseHeader(api.ETagHeader, fmt.Sprintf("%q", encryptedRef)),
		)

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusNotFound,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: "address not found or incorrect",
				Code:    http.StatusNotFound,
			}),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "application/json; charset=utf-8"),
		)
	})

	t.Run("upload-w/o-act-then-download-w/-act", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(),
		})
		var (
			rootHash   = "0cb947ccbc410c43139ba4409d83bf89114cb0d79556a651c06c888cf73f4d7e"
			sampleHtml = `<!DOCTYPE html>
			<html>
			<body>

			<h1>My First Heading</h1>

			<p>My first paragraph.</p>

			</body>
			</html>`
		)

		jsonhttptest.Request(t, client, http.MethodPost, fileUploadResource+"?name="+fileName, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(strings.NewReader(sampleHtml)),
			jsonhttptest.WithExpectedJSONResponse(api.BzzUploadResponse{
				Reference: swarm.MustParseHexAddress(rootHash),
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithNonEmptyResponseHeader(api.SwarmTagHeader),
			jsonhttptest.WithExpectedResponseHeader(api.ETagHeader, fmt.Sprintf("%q", rootHash)),
		)

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(rootHash), http.StatusInternalServerError,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(now, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: api.ErrActDownload.Error(),
				Code:    http.StatusInternalServerError,
			}),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "application/json; charset=utf-8"),
		)
	})
}

// nolint:paralleltest,tparallel
// TODO: TestDacInvalidPath doc. comment
func TestDacInvalidPath(t *testing.T) {
	t.Parallel()
	var (
		data, _              = hex.DecodeString("a786dd84b61485de12146fd9c4c02d87e8fd95f0542765cb7fc3d2e428c0bcfa")
		pk, _                = crypto.DecodeSecp256k1PrivateKey(data)
		publicKeyBytes       = crypto.EncodeSecp256k1PublicKey(&pk.PublicKey)
		publisher            = hex.EncodeToString(publicKeyBytes)
		fileDownloadResource = func(addr string) string { return "/bzz/" + addr }
		storerMock           = mockstorer.New()
		_, fixtureHref       = prepareHistoryFixture(storerMock)
		logger               = log.Noop
		now                  = time.Now().Unix()
	)

	t.Run("invalid-path-params", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(),
		})
		var (
			encryptedRef = "asd"
		)

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusBadRequest,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(now, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Code:    http.StatusBadRequest,
				Message: "invalid path params",
				Reasons: []jsonhttp.Reason{
					{
						Field: "address",
						Error: api.HexInvalidByteError('s').Error(),
					},
				}}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
		)
	})
}

// nolint:paralleltest,tparallel
// TODO: TestDacHistory doc. comment
func TestDacHistory(t *testing.T) {
	t.Parallel()
	var (
		data, _              = hex.DecodeString("a786dd84b61485de12146fd9c4c02d87e8fd95f0542765cb7fc3d2e428c0bcfa")
		pk, _                = crypto.DecodeSecp256k1PrivateKey(data)
		publicKeyBytes       = crypto.EncodeSecp256k1PublicKey(&pk.PublicKey)
		publisher            = hex.EncodeToString(publicKeyBytes)
		fileUploadResource   = "/bzz"
		fileDownloadResource = func(addr string) string { return "/bzz/" + addr }
		storerMock           = mockstorer.New()
		h, fixtureHref       = prepareHistoryFixture(storerMock)
		logger               = log.Noop
		fileName             = "sample.html"
		now                  = time.Now().Unix()
	)

	t.Run("empty-history-upload-then-download-and-check-data", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(),
		})
		var (
			testfile     = "testfile1"
			encryptedRef = "a5df670544eaea29e61b19d8739faa4573b19e4426e58a173e51ed0b5e7e2ade"
		)
		header := jsonhttptest.Request(t, client, http.MethodPost, fileUploadResource+"?name="+fileName, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmActHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(strings.NewReader(testfile)),
			jsonhttptest.WithExpectedJSONResponse(api.BzzUploadResponse{
				Reference: swarm.MustParseHexAddress(encryptedRef),
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithNonEmptyResponseHeader(api.SwarmTagHeader),
			jsonhttptest.WithExpectedResponseHeader(api.ETagHeader, fmt.Sprintf("%q", encryptedRef)),
		)

		historyRef := header.Get(api.SwarmActHistoryAddressHeader)
		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusOK,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(now, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, historyRef),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedResponse([]byte(testfile)),
			jsonhttptest.WithExpectedContentLength(len(testfile)),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithExpectedResponseHeader(api.ContentDispositionHeader, fmt.Sprintf(`inline; filename="%s"`, fileName)),
		)
	})

	t.Run("with-history-upload-then-download-and-check-data", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(mockdac.WithHistory(h, fixtureHref.String())),
		})
		var (
			encryptedRef = "c611199e1b3674d6bf89a83e518bd16896bf5315109b4a23dcb4682a02d17b97"
			testfile     = `<!DOCTYPE html>
			<html>
			<body>

			<h1>My First Heading</h1>

			<p>My first paragraph.</p>

			</body>
			</html>`
		)

		jsonhttptest.Request(t, client, http.MethodPost, fileUploadResource+"?name="+fileName, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmActHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestBody(strings.NewReader(testfile)),
			jsonhttptest.WithExpectedJSONResponse(api.BzzUploadResponse{
				Reference: swarm.MustParseHexAddress(encryptedRef),
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithNonEmptyResponseHeader(api.SwarmTagHeader),
			jsonhttptest.WithExpectedResponseHeader(api.ETagHeader, fmt.Sprintf("%q", encryptedRef)),
		)

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusOK,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(now, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedResponse([]byte(testfile)),
			jsonhttptest.WithExpectedContentLength(len(testfile)),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithExpectedResponseHeader(api.ContentDispositionHeader, fmt.Sprintf(`inline; filename="%s"`, fileName)),
		)
	})

	t.Run("upload-then-download-wrong-history", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(mockdac.WithHistory(h, fixtureHref.String())),
		})
		var (
			testfile     = "testfile1"
			encryptedRef = "a5df670544eaea29e61b19d8739faa4573b19e4426e58a173e51ed0b5e7e2ade"
		)
		jsonhttptest.Request(t, client, http.MethodPost, fileUploadResource+"?name="+fileName, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmActHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(strings.NewReader(testfile)),
			jsonhttptest.WithExpectedJSONResponse(api.BzzUploadResponse{
				Reference: swarm.MustParseHexAddress(encryptedRef),
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithNonEmptyResponseHeader(api.SwarmTagHeader),
			jsonhttptest.WithExpectedResponseHeader(api.ETagHeader, fmt.Sprintf("%q", encryptedRef)),
		)

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusInternalServerError,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(now, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, "fc4e9fe978991257b897d987bc4ff13058b66ef45a53189a0b4fe84bb3346396"),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: api.ErrActDownload.Error(),
				Code:    http.StatusInternalServerError,
			}),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "application/json; charset=utf-8"),
		)
	})

	t.Run("upload-wrong-history", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(),
		})
		var (
			testfile = "testfile1"
		)

		jsonhttptest.Request(t, client, http.MethodPost, fileUploadResource+"?name="+fileName, http.StatusInternalServerError,
			jsonhttptest.WithRequestHeader(api.SwarmActHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestBody(strings.NewReader(testfile)),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: api.ErrActUpload.Error(),
				Code:    http.StatusInternalServerError,
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
		)
	})

	t.Run("download-w/o-history", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(mockdac.WithHistory(h, fixtureHref.String())),
		})
		var (
			encryptedRef = "a5df670544eaea29e61b19d8739faa4573b19e4426e58a173e51ed0b5e7e2ade"
		)

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusNotFound,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(now, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "application/json; charset=utf-8"),
		)
	})
}

// nolint:paralleltest,tparallel
// TODO: TestDacTimestamp doc. comment
func TestDacTimestamp(t *testing.T) {
	t.Parallel()
	var (
		data, _              = hex.DecodeString("a786dd84b61485de12146fd9c4c02d87e8fd95f0542765cb7fc3d2e428c0bcfa")
		pk, _                = crypto.DecodeSecp256k1PrivateKey(data)
		publicKeyBytes       = crypto.EncodeSecp256k1PublicKey(&pk.PublicKey)
		publisher            = hex.EncodeToString(publicKeyBytes)
		fileUploadResource   = "/bzz"
		fileDownloadResource = func(addr string) string { return "/bzz/" + addr }
		storerMock           = mockstorer.New()
		h, fixtureHref       = prepareHistoryFixture(storerMock)
		logger               = log.Noop
		fileName             = "sample.html"
	)
	t.Run("upload-then-download-with-timestamp-and-check-data", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(mockdac.WithHistory(h, fixtureHref.String())),
		})
		var (
			thirdTime    = time.Date(2015, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
			encryptedRef = "c611199e1b3674d6bf89a83e518bd16896bf5315109b4a23dcb4682a02d17b97"
			testfile     = `<!DOCTYPE html>
			<html>
			<body>

			<h1>My First Heading</h1>

			<p>My first paragraph.</p>

			</body>
			</html>`
		)

		jsonhttptest.Request(t, client, http.MethodPost, fileUploadResource+"?name="+fileName, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmActHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestBody(strings.NewReader(testfile)),
			jsonhttptest.WithExpectedJSONResponse(api.BzzUploadResponse{
				Reference: swarm.MustParseHexAddress(encryptedRef),
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithNonEmptyResponseHeader(api.SwarmTagHeader),
			jsonhttptest.WithExpectedResponseHeader(api.ETagHeader, fmt.Sprintf("%q", encryptedRef)),
		)

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusOK,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(thirdTime, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedResponse([]byte(testfile)),
			jsonhttptest.WithExpectedContentLength(len(testfile)),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithExpectedResponseHeader(api.ContentDispositionHeader, fmt.Sprintf(`inline; filename="%s"`, fileName)),
		)
	})

	t.Run("download-w/o-timestamp", func(t *testing.T) {
		var (
			encryptedRef = "a5df670544eaea29e61b19d8739faa4573b19e4426e58a173e51ed0b5e7e2ade"
		)
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(mockdac.WithHistory(h, fixtureHref.String())),
		})

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusNotFound,
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "application/json; charset=utf-8"),
		)
	})
}

// nolint:paralleltest,tparallel
// TODO: TestDacPublisher doc. comment
func TestDacPublisher(t *testing.T) {
	t.Parallel()
	var (
		data, _              = hex.DecodeString("a786dd84b61485de12146fd9c4c02d87e8fd95f0542765cb7fc3d2e428c0bcfa")
		pk, _                = crypto.DecodeSecp256k1PrivateKey(data)
		publicKeyBytes       = crypto.EncodeSecp256k1PublicKey(&pk.PublicKey)
		publisher            = hex.EncodeToString(publicKeyBytes)
		fileUploadResource   = "/bzz"
		fileDownloadResource = func(addr string) string { return "/bzz/" + addr }
		storerMock           = mockstorer.New()
		h, fixtureHref       = prepareHistoryFixture(storerMock)
		logger               = log.Noop
		fileName             = "sample.html"
		now                  = time.Now().Unix()
	)

	t.Run("upload-then-download-w/-publisher-and-check-data", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(mockdac.WithHistory(h, fixtureHref.String()), mockdac.WithPublisher(publisher)),
		})
		var (
			encryptedRef = "a5a26b4915d7ce1622f9ca52252092cf2445f98d359dabaf52588c05911aaf4f"
			testfile     = `<!DOCTYPE html>
			<html>
			<body>

			<h1>My First Heading</h1>

			<p>My first paragraph.</p>

			</body>
			</html>`
		)

		jsonhttptest.Request(t, client, http.MethodPost, fileUploadResource+"?name="+fileName, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmActHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestBody(strings.NewReader(testfile)),
			jsonhttptest.WithExpectedJSONResponse(api.BzzUploadResponse{
				Reference: swarm.MustParseHexAddress(encryptedRef),
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithNonEmptyResponseHeader(api.SwarmTagHeader),
			jsonhttptest.WithExpectedResponseHeader(api.ETagHeader, fmt.Sprintf("%q", encryptedRef)),
		)

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusOK,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(now, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedResponse([]byte(testfile)),
			jsonhttptest.WithExpectedContentLength(len(testfile)),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithExpectedResponseHeader(api.ContentDispositionHeader, fmt.Sprintf(`inline; filename="%s"`, fileName)),
		)
	})

	t.Run("upload-then-invalid-publickey", func(t *testing.T) {
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(mockdac.WithPublisher(publisher)),
		})
		var (
			publickey    = "b786dd84b61485de12146fd9c4c02d87e8fd95f0542765cb7fc3d2e428c0bcfb"
			encryptedRef = "a5a26b4915d7ce1622f9ca52252092cf2445f98d359dabaf52588c05911aaf4f"
			testfile     = `<!DOCTYPE html>
			<html>
			<body>

			<h1>My First Heading</h1>

			<p>My first paragraph.</p>

			</body>
			</html>`
		)

		header := jsonhttptest.Request(t, client, http.MethodPost, fileUploadResource+"?name="+fileName, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmActHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(strings.NewReader(testfile)),
			jsonhttptest.WithExpectedJSONResponse(api.BzzUploadResponse{
				Reference: swarm.MustParseHexAddress(encryptedRef),
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithNonEmptyResponseHeader(api.SwarmTagHeader),
			jsonhttptest.WithExpectedResponseHeader(api.ETagHeader, fmt.Sprintf("%q", encryptedRef)),
		)

		historyRef := header.Get(api.SwarmActHistoryAddressHeader)
		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusBadRequest,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(now, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, historyRef),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publickey),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Code:    http.StatusBadRequest,
				Message: "invalid header params",
				Reasons: []jsonhttp.Reason{
					{
						Field: "Swarm-Act-Publisher",
						Error: "malformed public key: invalid length: 32",
					},
				}}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
		)
	})

	t.Run("download-w/-wrong-publisher", func(t *testing.T) {
		var (
			downloader   = "03c712a7e29bc792ac8d8ae49793d28d5bda27ed70f0d90697b2fb456c0a168bd2"
			encryptedRef = "a5df670544eaea29e61b19d8739faa4573b19e4426e58a173e51ed0b5e7e2ade"
		)
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(mockdac.WithHistory(h, fixtureHref.String()), mockdac.WithPublisher(publisher)),
		})

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusInternalServerError,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(now, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, downloader),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: api.ErrActDownload.Error(),
				Code:    http.StatusInternalServerError,
			}),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "application/json; charset=utf-8"),
		)
	})

	t.Run("download-w/o-publisher", func(t *testing.T) {
		var (
			encryptedRef = "a5df670544eaea29e61b19d8739faa4573b19e4426e58a173e51ed0b5e7e2ade"
		)
		client, _, _, _ := newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
			Dac:       mockdac.New(mockdac.WithHistory(h, fixtureHref.String()), mockdac.WithPublisher(publisher)),
		})

		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(encryptedRef), http.StatusNotFound,
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(now, 10)),
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, fixtureHref.String()),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "application/json; charset=utf-8"),
		)
	})
}
