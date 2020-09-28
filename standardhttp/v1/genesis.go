// Copyright © 2020 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"context"
	"encoding/json"

	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/pkg/errors"
)

type genesisJSON struct {
	Data *api.Genesis `json:"data"`
}

// Genesis provides the genesis information of the chain.
func (s *Service) Genesis(ctx context.Context) (*api.Genesis, error) {
	if s.genesis == nil {
		respBodyReader, err := s.get(ctx, "/eth/v1/beacon/genesis")
		if err != nil {
			return nil, errors.Wrap(err, "failed to request genesis")
		}

		var resp genesisJSON
		if err := json.NewDecoder(respBodyReader).Decode(&resp); err != nil {
			return nil, errors.Wrap(err, "failed to parse genesis")
		}
		s.genesis = resp.Data
	}

	return &api.Genesis{
		GenesisTime:           s.genesis.GenesisTime,
		GenesisValidatorsRoot: s.genesis.GenesisValidatorsRoot,
		GenesisForkVersion:    s.genesis.GenesisForkVersion,
	}, nil
}
