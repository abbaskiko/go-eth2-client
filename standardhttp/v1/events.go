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
	"fmt"
	"net/url"
	"strings"

	client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"github.com/r3labs/sse/v2"
)

// Events feeds requested events with the given topics to the supplied handler.
func (s *Service) Events(ctx context.Context, topics []string, handler client.EventHandlerFunc) error {
	if len(topics) == 0 {
		return errors.New("no topics supplied")
	}

	// Ensure we support the requested topic(s).
	for i := range topics {
		if _, exists := api.SupportedEventTopics[topics[i]]; !exists {
			return fmt.Errorf("unsupported event topic %s", topics[i])
		}
	}

	reference, err := url.Parse(fmt.Sprintf("/eth/v1/events?topics=%s", strings.Join(topics, "&topics=")))
	if err != nil {
		return errors.Wrap(err, "invalid endpoint")
	}
	url := s.base.ResolveReference(reference).String()
	log.Trace().Str("url", url).Msg("GET request to events stream")

	client := sse.NewClient(url)
	go func() {
		if err := client.SubscribeRawWithContext(ctx, func(msg *sse.Event) {
			s.handleEvent(msg, handler)
		}); err != nil {
			log.Error().Err(err).Msg("Failed to subscribe to event stream")
		}
	}()

	return nil
}

// handleEvent parses an event and passes it on to the handler.
func (s *Service) handleEvent(msg *sse.Event, handler client.EventHandlerFunc) {
	event := &api.Event{
		Topic: string(msg.Event),
	}
	switch string(msg.Event) {
	case "head":
		headEvent := &api.HeadEvent{}
		err := json.Unmarshal(msg.Data, headEvent)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse head event")
		}
		event.Data = headEvent
	case "block":
		blockEvent := &api.BlockEvent{}
		err := json.Unmarshal(msg.Data, blockEvent)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse block event")
		}
		event.Data = blockEvent
	case "attestation":
		attestation := &spec.Attestation{}
		err := json.Unmarshal(msg.Data, attestation)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse attestation")
		}
		event.Data = attestation
	case "voluntary_exit":
		voluntaryExit := &spec.SignedVoluntaryExit{}
		err := json.Unmarshal(msg.Data, voluntaryExit)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse voluntary exit")
		}
		event.Data = voluntaryExit
	case "finalized_checkpoint":
		finalizedCheckpointEvent := &api.FinalizedCheckpointEvent{}
		err := json.Unmarshal(msg.Data, finalizedCheckpointEvent)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse finalized checkpoint event")
		}
		event.Data = finalizedCheckpointEvent
	case "chain_reorg":
		chainReorgEvent := &api.ChainReorgEvent{}
		err := json.Unmarshal(msg.Data, chainReorgEvent)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse chain reorg event")
		}
		event.Data = chainReorgEvent
	case "":
		// A message with a blank event comes when the event stream shuts down.  Ignore it.
	default:
		log.Warn().Str("topic", string(msg.Event)).Msg("Received message with unhandled topic")
	}
	handler(event)
}
