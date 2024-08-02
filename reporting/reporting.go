// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reporting

import (
	"context"
	"fmt"
	"time"

	"github.com/czcorpus/hltscl"
	"github.com/rs/zerolog/log"
)

/*
Expected tables:

create table camus_operations_stats (
  "time" timestamp with time zone NOT NULL,
  num_fetched int,
  num_errors int,
  num_merged int,
  num_inserted int
);

select create_hypertable('camus_operations_stats', 'time');

create table camus_cleanup_stats (
  "time" timestamp with time zone NOT NULL,
  num_fetched int,
  num_merged int,
  num_errors int
);

select create_hypertable('camus_cleanup_stats', 'time');

*/

type StatusWriter struct {
	tableWriterOps     *hltscl.TableWriter
	tableWriterCleanup *hltscl.TableWriter
	opsDataCh          chan<- hltscl.Entry
	cleanupDataCh      chan<- hltscl.Entry
	errCh              <-chan hltscl.WriteError
	location           *time.Location
}

func (job *StatusWriter) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("about to close ArchKeeper")
				return
			case err := <-job.errCh:
				log.Error().
					Err(err.Err).
					Str("entry", err.Entry.String()).
					Msg("error writing data to TimescaleDB")
				fmt.Println("reporting timescale write err: ", err.Err)
			}
		}
	}()
}

func (job *StatusWriter) Stop(ctx context.Context) error {
	log.Warn().Msg("stopping StatusWriter")
	return nil
}

func (ds *StatusWriter) WriteOperationsStatus(item OpStats) {
	if ds.tableWriterOps != nil {
		ds.opsDataCh <- *ds.tableWriterOps.NewEntry(time.Now().In(ds.location)).
			Int("num_merged", item.NumMerged).
			Int("num_errors", item.NumErrors).
			Int("num_fetched", item.NumFetched).
			Int("num_inserted", item.NumInserted)
	}
}

func (ds *StatusWriter) WriteCleanupStatus(item CleanupStats) {
	if ds.tableWriterCleanup != nil {
		fmt.Println("writing data")
		ds.cleanupDataCh <- *ds.tableWriterCleanup.NewEntry(time.Now().In(ds.location)).
			Int("num_errors", item.NumErrors).
			Int("num_fetched", item.NumFetched).
			Int("num_merged", item.NumMerged)
	}
}

func NewStatusWriter(conf hltscl.PgConf, tz *time.Location, onError func(err error)) (*StatusWriter, error) {

	conn, err := hltscl.CreatePool(conf)
	if err != nil {
		return nil, err
	}
	twriter1 := hltscl.NewTableWriter(conn, "camus_operations_stats", "time", tz)
	opsDataCh, errCh1 := twriter1.Activate()
	twriter2 := hltscl.NewTableWriter(conn, "camus_cleanup_stats", "time", tz)
	cleanupDataCh, errCh2 := twriter2.Activate()
	mergedErr := make(chan hltscl.WriteError)
	go func() {
		for err := range errCh1 {
			mergedErr <- err
		}
	}()
	go func() {
		for err := range errCh2 {
			mergedErr <- err
		}
	}()

	return &StatusWriter{
		tableWriterOps:     twriter1,
		tableWriterCleanup: twriter2,
		opsDataCh:          opsDataCh,
		cleanupDataCh:      cleanupDataCh,
		errCh:              mergedErr,
		location:           tz,
	}, nil
}
