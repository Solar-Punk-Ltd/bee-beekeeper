// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mantaray_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/ethersphere/bee/v2/pkg/manifest/mantaray"
)

func TestWalkNodePathSequence(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		toAdd    [][]byte
		expected [][]byte
	}{
		{
			name: "simple",
			toAdd: [][]byte{
				[]byte("111111"),    // 1994
				[]byte("111111435"), // 2000
				[]byte("111111257"), // 2015
				[]byte("111111256"), // 2020
				[]byte("111111258"),
				[]byte("111111334"), // 2030
			},
			expected: [][]byte{
				[]byte(""),
				[]byte("111111"),
				[]byte("111111435"),
				[]byte("11111125"),
				[]byte("111111257"),
				[]byte("111111256"),
				[]byte("111111258"),
				[]byte("111111334"),
			},
		},
	} {
		ctx := context.Background()
		tc := tc

		createTree := func(t *testing.T, toAdd [][]byte) *mantaray.Node {
			t.Helper()

			n := mantaray.New()

			for i := 0; i < len(toAdd); i++ {
				c := toAdd[i]
				e := append(make([]byte, 32-len(c)), c...)
				err := n.Add(ctx, c, e, nil, nil)
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
			return n
		}

		pathExistsInRightSequence := func(found []byte, expected [][]byte, walkedCount int) bool {
			rightPathInSequence := false

			c := expected[walkedCount]
			if bytes.Equal(found, c) {
				rightPathInSequence = true
			}

			return rightPathInSequence
		}

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n := createTree(t, tc.toAdd)

			walkedCount := 0

			walker := func(path []byte, node *mantaray.Node, err error) error {

				if !pathExistsInRightSequence(path, tc.expected, walkedCount) {
					return fmt.Errorf("walkFn returned unknown path: %s", path)
				}
				walkedCount++
				return nil
			}
			// Expect no errors.
			err := n.WalkNode(ctx, []byte{}, nil, walker)
			if err != nil {
				t.Fatalf("no error expected, found: %s", err)
			}

			if len(tc.expected) != walkedCount {
				t.Errorf("expected %d nodes, got %d", len(tc.expected), walkedCount)
			}
		})
	}
}

func TestWalkNode(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		toAdd    [][]byte
		expected [][]byte
	}{
		{
			name: "simple",
			toAdd: [][]byte{
				[]byte("9223372036089617407"), // 1994
				[]byte("9223372035900228607"), // 2000
				[]byte("9223372035426929407"), // 2015
				[]byte("9223372035269076607"), // 2020
				[]byte("9223372034953543807"), // 2030
			},
			expected: [][]byte{
				[]byte(""),
				[]byte("922337203"),
				[]byte("9223372034953543807"),
				[]byte("9223372035"),
				[]byte("9223372035269076607"),
				[]byte("9223372035426929407"),
				[]byte("9223372035900228607"),
			},
		},
	} {
		ctx := context.Background()
		tc := tc

		createTree := func(t *testing.T, toAdd [][]byte) *mantaray.Node {
			t.Helper()

			n := mantaray.New()

			for i := 0; i < len(toAdd); i++ {
				c := toAdd[i]
				e := append(make([]byte, 32-len(c)), c...)
				err := n.Add(ctx, c, e, nil, nil)
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
			return n
		}

		pathExists := func(found []byte, expected [][]byte) bool {
			pathFound := false

			for i := 0; i < len(tc.expected); i++ {
				c := tc.expected[i]
				if bytes.Equal(found, c) {
					pathFound = true
					break
				}
			}
			return pathFound
		}

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n := createTree(t, tc.toAdd)

			walkedCount := 0

			walker := func(path []byte, node *mantaray.Node, err error) error {
				walkedCount++

				if !pathExists(path, tc.expected) {
					return fmt.Errorf("walkFn returned unknown path: %s", path)
				}
				return nil
			}
			// Expect no errors.
			err := n.WalkNode(ctx, []byte{}, nil, walker)
			if err != nil {
				t.Fatalf("no error expected, found: %s", err)
			}

			if len(tc.expected) != walkedCount {
				t.Errorf("expected %d nodes, got %d", len(tc.expected), walkedCount)
			}
		})

		t.Run(tc.name+"/with load save", func(t *testing.T) {
			t.Parallel()

			n := createTree(t, tc.toAdd)

			ls := newMockLoadSaver()

			err := n.Save(ctx, ls)
			if err != nil {
				t.Fatal(err)
			}

			n2 := mantaray.NewNodeRef(n.Reference())

			walkedCount := 0

			walker := func(path []byte, node *mantaray.Node, err error) error {
				walkedCount++

				if !pathExists(path, tc.expected) {
					return fmt.Errorf("walkFn returned unknown path: %s", path)
				}

				return nil
			}
			// Expect no errors.
			err = n2.WalkNode(ctx, []byte{}, ls, walker)
			if err != nil {
				t.Fatalf("no error expected, found: %s", err)
			}

			if len(tc.expected) != walkedCount {
				t.Errorf("expected %d nodes, got %d", len(tc.expected), walkedCount)
			}
		})
	}
}

func TestWalk(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		toAdd    [][]byte
		expected [][]byte
	}{
		{
			name: "simple",
			toAdd: [][]byte{
				[]byte("index.html"),
				[]byte("img/test/"),
				[]byte("img/test/oho.png"),
				[]byte("img/test/old/test.png"),
				// file with same prefix but not a directory prefix
				[]byte("img/test/old/test.png.backup"),
				[]byte("robots.txt"),
			},
			expected: [][]byte{
				[]byte("index.html"),
				[]byte("img"),
				[]byte("img/test"),
				[]byte("img/test/oho.png"),
				[]byte("img/test/old"),
				[]byte("img/test/old/test.png"),
				[]byte("img/test/old/test.png.backup"),
				[]byte("robots.txt"),
			},
		},
	} {
		ctx := context.Background()

		createTree := func(t *testing.T, toAdd [][]byte) *mantaray.Node {
			t.Helper()

			n := mantaray.New()

			for i := 0; i < len(toAdd); i++ {
				c := toAdd[i]
				e := append(make([]byte, 32-len(c)), c...)
				err := n.Add(ctx, c, e, nil, nil)
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
			return n
		}

		pathExists := func(found []byte, expected [][]byte) bool {
			pathFound := false

			for i := 0; i < len(tc.expected); i++ {
				c := tc.expected[i]
				if bytes.Equal(found, c) {
					pathFound = true
					break
				}
			}
			return pathFound
		}

		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n := createTree(t, tc.toAdd)

			walkedCount := 0

			walker := func(path []byte, isDir bool, err error) error {
				walkedCount++

				if !pathExists(path, tc.expected) {
					return fmt.Errorf("walkFn returned unknown path: %s", path)
				}

				return nil
			}
			// Expect no errors.
			err := n.Walk(ctx, []byte{}, nil, walker)
			if err != nil {
				t.Fatalf("no error expected, found: %s", err)
			}

			if len(tc.expected) != walkedCount {
				t.Errorf("expected %d nodes, got %d", len(tc.expected), walkedCount)
			}

		})

		t.Run(tc.name+"/with load save", func(t *testing.T) {
			t.Parallel()

			n := createTree(t, tc.toAdd)

			ls := newMockLoadSaver()

			err := n.Save(ctx, ls)
			if err != nil {
				t.Fatal(err)
			}

			n2 := mantaray.NewNodeRef(n.Reference())

			walkedCount := 0

			walker := func(path []byte, isDir bool, err error) error {
				walkedCount++

				if !pathExists(path, tc.expected) {
					return fmt.Errorf("walkFn returned unknown path: %s", path)
				}

				return nil
			}
			// Expect no errors.
			err = n2.Walk(ctx, []byte{}, ls, walker)
			if err != nil {
				t.Fatalf("no error expected, found: %s", err)
			}

			if len(tc.expected) != walkedCount {
				t.Errorf("expected %d nodes, got %d", len(tc.expected), walkedCount)
			}

		})
	}
}
