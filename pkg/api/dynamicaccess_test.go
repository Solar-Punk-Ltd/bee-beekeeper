// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api_test

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/ethersphere/bee/v2/pkg/log"
	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"

	"github.com/ethersphere/bee/v2/pkg/api"
	"github.com/ethersphere/bee/v2/pkg/jsonhttp/jsonhttptest"
)

// nolint:paralleltest,tparallel
// TODO: TestDacUploadDownload doc. comment
func TestDacUploadDownload(t *testing.T) {
	t.Parallel()
	var (
		pk, _                = crypto.GenerateSecp256k1Key()
		publicKeyBytes       = crypto.EncodeSecp256k1PublicKey(&pk.PublicKey)
		publisher            = hex.EncodeToString(publicKeyBytes)
		fileUploadResource   = "/bzz"
		fileDownloadResource = func(addr string) string { return "/bzz/" + addr }
		storerMock           = mockstorer.New()
		logger               = log.Noop
		client, _, _, _      = newTestServer(t, testServerOptions{
			Storer:    storerMock,
			Logger:    logger,
			Post:      mockpost.New(mockpost.WithAcceptAll()),
			PublicKey: pk.PublicKey,
		})
	)

	t.Run("upload-then-download-and-check-data", func(t *testing.T) {
		fileName := "sample.html"
		rootHash := "36e6c1bbdfee6ac21485d5f970479fd1df458d36df9ef4e8179708ed46da557f"
		timestamp := time.Now().Unix()
		sampleHtml := `<!DOCTYPE html>
		<html>
		<body>

		<h1>My First Heading</h1>

		<p>My first paragraph.</p>

		</body>
		</html>`

		// TODO: proper mocked act headers
		// TODO: expect different root hash because of encryption
		header := jsonhttptest.Request(t, client, http.MethodPost, fileUploadResource+"?name="+fileName, http.StatusCreated,
			jsonhttptest.WithRequestHeader(api.SwarmActHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(strings.NewReader(sampleHtml)),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithNonEmptyResponseHeader(api.SwarmTagHeader),
		)

		// try to fetch the same file and check the data
		historyRef := header.Get(api.SwarmActHistoryAddressHeader)
		jsonhttptest.Request(t, client, http.MethodGet, fileDownloadResource(rootHash), http.StatusOK,
			jsonhttptest.WithRequestHeader(api.SwarmActHeader, "true"),
			jsonhttptest.WithRequestHeader(api.SwarmActTimestampHeader, strconv.FormatInt(timestamp, 10)), // Convert timestamp to int
			jsonhttptest.WithRequestHeader(api.SwarmActHistoryAddressHeader, historyRef),
			jsonhttptest.WithRequestHeader(api.SwarmActPublisherHeader, publisher),
			jsonhttptest.WithExpectedResponse([]byte(sampleHtml)),
			jsonhttptest.WithExpectedContentLength(len(sampleHtml)),
			jsonhttptest.WithExpectedResponseHeader(api.ContentTypeHeader, "text/html; charset=utf-8"),
			jsonhttptest.WithExpectedResponseHeader(api.ContentDispositionHeader, fmt.Sprintf(`inline; filename="%s"`, fileName)),
		)
	})

}
