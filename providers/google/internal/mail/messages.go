package mail

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/spyzhov/ajson"
)

// fetchMessages retrieves the full message payloads for all message IDs found
// in the provided collection response.
//
// It fans out concurrent fetch jobs using simultaneously.DoCtx, then fans in
// results through a buffered channel. Only this function goroutine builds the
// result map, so no mutex is required.
//
// Returns a map keyed by message ID containing the fully populated MessageRecords.
func (a *Adapter) fetchMessages(
	ctx context.Context, objectName string, collectionResp *common.JSONHTTPResponse,
) (MessageRecords, error) {
	messageIDs, err := extractMessageIdentifiers(objectName, collectionResp)
	if err != nil {
		return nil, err
	}

	if len(messageIDs) == 0 {
		return MessageRecords{}, nil
	}

	messagesChannel := make(chan MessageRecord, len(messageIDs))
	callbacks := make([]simultaneously.Job, 0, len(messageIDs))

	// Fan-out: create one job per message ID.
	for _, messageID := range messageIDs {
		callbacks = append(callbacks, a.fetchMessage(messagesChannel, messageID))
	}

	// Run all jobs concurrently. If any job fails, context is expected to cancel.
	if err = simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		return nil, err
	}

	// All jobs are done, so no more sends will occur.
	close(messagesChannel)

	// Fan-in: single goroutine owns the map.
	messageRegistry := make(MessageRecords, len(messageIDs))

	for message := range messagesChannel {
		id, ok := message["id"].(string)
		if !ok {
			return nil, errors.New("missing field 'id' in response for object 'messages'") // nolint:err113
		}

		messageRegistry[id] = message
	}

	return messageRegistry, nil
}

// fetchMessage returns a concurrently executable job that fetches a single message
// by ID and sends the fully decoded MessageRecord to messagesChannel.
func (a *Adapter) fetchMessage(messagesChannel chan MessageRecord, messageId string,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		url, err := a.getMessageURL(messageId)
		if err != nil {
			return err
		}

		resp, err := a.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			return err
		}

		message, err := common.UnmarshalJSON[MessageRecord](resp)
		if err != nil {
			return err
		}

		select {
		case messagesChannel <- *message:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func messagesEmbedMessageRaw(messages MessageRecords) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		root, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		identifier, err := jsonquery.New(node).StringRequired("id")
		if err != nil {
			return nil, err
		}

		message, err := messages.findByID(identifier)
		if err != nil {
			return nil, err
		}

		maps.Copy(root, message)

		return root, nil
	}
}

func messagesEmbedMessageFields(params common.ReadParams, messages MessageRecords) readhelper.SelectedFieldsFunc {
	return func(node *ajson.Node) (map[string]any, string, error) {
		root, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, "", err
		}

		identifier, err := jsonquery.New(node).StringRequired("id")
		if err != nil {
			return nil, "", err
		}

		message, err := messages.findByID(identifier)
		if err != nil {
			return nil, "", err
		}

		filteredRoot := readhelper.SelectFields(root, params.Fields)
		selected := readhelper.SelectFields(message, params.Fields)

		// Combine fields of Message object from the Collection representation and from singular Item.
		maps.Copy(filteredRoot, selected)

		return filteredRoot, identifier, nil
	}
}

func draftsEmbedMessageRaw(messages MessageRecords) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		root, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		messageID, err := jsonquery.New(node, "message").StringRequired("id")
		if err != nil {
			return nil, err
		}

		message, err := messages.findByID(messageID)
		if err != nil {
			return nil, err
		}

		// Draft has a 'message' field which we populate.
		originalMessage, _ := root["message"].(map[string]any)
		combinedMessage := make(map[string]any)
		maps.Copy(combinedMessage, originalMessage)
		maps.Copy(combinedMessage, message)
		root["message"] = combinedMessage

		return root, nil
	}
}

func draftsEmbedMessageFields(params common.ReadParams, messages MessageRecords) readhelper.SelectedFieldsFunc {
	return func(node *ajson.Node) (map[string]any, string, error) {
		root, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, "", err
		}

		identifier, err := jsonquery.New(node).StringRequired("id")
		if err != nil {
			return nil, "", err
		}

		messageID, err := jsonquery.New(node, "message").StringRequired("id")
		if err != nil {
			return nil, "", err
		}

		message, err := messages.findByID(messageID)
		if err != nil {
			return nil, "", err
		}

		// Draft has a 'message' field which we populate.
		originalMessage, _ := root["message"].(map[string]any)
		combinedMessage := make(map[string]any)
		maps.Copy(combinedMessage, originalMessage)
		maps.Copy(combinedMessage, message)
		root["message"] = combinedMessage

		filtered := readhelper.SelectFields(root, params.Fields)

		return filtered, identifier, nil
	}
}

// MessagesCollection https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/list
type MessagesCollection struct {
	Messages []messageSchema `json:"messages"`
}

type messageSchema struct {
	ID string `json:"id"`
}

func (c MessagesCollection) getMessageIdentifiers() []string {
	return datautils.ForEach(c.Messages, func(message messageSchema) string {
		return message.ID
	})
}

// DraftCollection https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.drafts/list
type DraftCollection struct {
	Drafts []draftSchema `json:"drafts"`
}

type draftSchema struct {
	ID      string `json:"id"`
	Message struct {
		ID string `json:"id"`
	} `json:"message"`
}

func (c DraftCollection) getMessageIdentifiers() []string {
	return datautils.ForEach(c.Drafts, func(draft draftSchema) string {
		return draft.Message.ID
	})
}

// MessageRecord contains MessagePart and MessagePartBody.
// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages#resource:-message
// nolint:lll
// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages.attachments#resource:-messagepartbody
type MessageRecord map[string]any

type MessageRecords map[string]MessageRecord

func (m MessageRecords) findByID(identifier string) (MessageRecord, error) {
	message, ok := m[identifier]
	if !ok || message == nil {
		return nil, fmt.Errorf( // nolint:err113
			"message with identifier %v was not found in the fetched registry", identifier)
	}

	return message, nil
}

type messageIdentifiersHolder interface {
	getMessageIdentifiers() []string
}

// extractMessageIdentifiers normalizes message IDs from Gmail collection response.
//
// Gmail returns message identifiers in different JSON shapes depending on the object:
//   - messages 	-> IDs at messages[].id
//   - drafts 	-> IDs at drafts[].message.id
func extractMessageIdentifiers(objectName string, collectionResp *common.JSONHTTPResponse) ([]string, error) {
	var (
		collection messageIdentifiersHolder
		err        error
	)

	switch objectName {
	case objectNameMessages:
		collection, err = common.UnmarshalJSON[MessagesCollection](collectionResp)
	case objectNameDrafts:
		collection, err = common.UnmarshalJSON[DraftCollection](collectionResp)
	default:
		return nil, fmt.Errorf( // nolint:err113
			"message identifier extraction is not implemented for object %v", objectName)
	}

	if err != nil {
		return nil, err
	}

	return collection.getMessageIdentifiers(), nil
}
