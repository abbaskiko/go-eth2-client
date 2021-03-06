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

package prysmgrpc

import (
	"context"

	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

// BeaconBlockRootBySlot fetches a block's root given its slot.
func (s *Service) BeaconBlockRootBySlot(ctx context.Context, slot uint64) ([]byte, error) {
	conn := ethpb.NewBeaconChainClient(s.conn)

	req := &ethpb.ListBlocksRequest{}
	if slot == 0 {
		req.QueryFilter = &ethpb.ListBlocksRequest_Genesis{Genesis: true}
	} else {
		req.QueryFilter = &ethpb.ListBlocksRequest_Slot{Slot: slot}
	}
	opCtx, cancel := context.WithTimeout(ctx, s.timeout)
	resp, err := conn.ListBlocks(opCtx, req)
	cancel()
	if err != nil {
		return nil, errors.Wrap(err, "call to ListBlocks() failed")
	}
	if len(resp.BlockContainers) == 0 {
		return nil, nil
	}

	root, err := resp.BlockContainers[0].Block.Block.Body.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate root of block")
	}

	return root[:], nil
}
