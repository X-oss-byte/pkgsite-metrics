// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vulndb

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/pkgsite-metrics/internal/bigquery"
	"golang.org/x/pkgsite-metrics/internal/osv"
	test "golang.org/x/pkgsite-metrics/internal/testing"
)

func TestConvert(t *testing.T) {
	oe := &osv.Entry{
		ID: "a",
		Affected: []osv.Affected{
			{Module: osv.Module{Path: "example.mod/a"}, Ranges: []osv.Range{{Events: []osv.RangeEvent{{Introduced: "0"}, {Fixed: "0.9.0"}}}}},
			{Module: osv.Module{Path: "a.example.mod/a"}, Ranges: []osv.Range{{Events: []osv.RangeEvent{{Introduced: "1.0.0"}, {Fixed: "2.0.0"}}}}},
		}}
	want := &Entry{
		ID: "a",
		Modules: []Module{
			{
				Path:   "example.mod/a",
				Ranges: []Range{{Introduced: "0"}, {Fixed: "0.9.0"}},
			},
			{
				Path:   "a.example.mod/a",
				Ranges: []Range{{Introduced: "1.0.0"}, {Fixed: "2.0.0"}},
			},
		},
	}
	got := Convert(oe)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch (-want, +got):\n%s", diff)
	}
}

func TestReadMostRecentDB(t *testing.T) {
	test.NeedsIntegrationEnv(t)

	ctx := context.Background()
	const projectID = "go-ecosystem"

	client, err := bigquery.NewClientForTesting(ctx, projectID, "read_recent_db")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	writeToBigQuery := func(es []*Entry) {
		if err := client.CreateTable(ctx, TableName); err != nil {
			t.Fatal(err)
		}
		if err := bigquery.UploadMany(ctx, client, TableName, es, 0); err != nil {
			t.Fatal(err)
		}
	}

	lmt := time.Now()
	es := []*Entry{
		{ID: "A"},
		{ID: "A", ModifiedTime: lmt},
		{ID: "B", ModifiedTime: lmt},
	}
	writeToBigQuery(es)

	got, err := ReadMostRecentDB(ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 rows; got %d", len(got))
	}
	for _, e := range got {
		// Ideally, we would check lmt != e.ModifiedTime but
		// unmarshaling time.Time introduces some nanosecond
		// level imprecision. Instead, we just check that it
		// is actually set.
		if e.ModifiedTime.IsZero() {
			t.Fatalf("want last modified time %v; got %v", lmt, e.ModifiedTime)
		}
	}
}
