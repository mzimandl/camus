// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
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

package indexer

import (
	"fmt"
)

type Conf struct {
	IndexFilePath    string `json:"indexFilePath"`
	DocRemoveChannel string `json:"docRemoveChannel"`
}

func (conf *Conf) ValidateAndDefaults() error {
	if conf == nil {
		return fmt.Errorf("missing `indexer` section")
	}
	if conf.IndexFilePath == "" {
		return fmt.Errorf("missing path to index file (indexFilePath)")
	}
	return nil
}