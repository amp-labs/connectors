package granola

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/spyzhov/ajson"
)

const objectNotes = "notes"
const maxConcurrentNotesFetches = 4

// NotesCollection is the response shape of the List Notes endpoint.
type NotesCollection struct {
	Notes []struct {
		ID string `json:"id"`
	} `json:"notes"`
}

// NoteRecord holds the full note payload returned by Get Note.
type NoteRecord map[string]any

// NoteRecords maps note ID to the full note payload.
type NoteRecords map[string]NoteRecord

// fetchNotes retrieves the full note payloads for all note IDs found in the
// provided list response.
func (c *Connector) fetchNotes(
	ctx context.Context, collectionResp *common.JSONHTTPResponse,
) (NoteRecords, error) {
	collection, err := common.UnmarshalJSON[NotesCollection](collectionResp)
	if err != nil {
		return nil, err
	}

	notesChannel := make(chan NoteRecord, len(collection.Notes))
	callbacks := make([]simultaneously.Job, 0, len(collection.Notes))

	// create one job per note ID.
	for _, note := range collection.Notes {
		callbacks = append(callbacks, c.fetchNote(notesChannel, note.ID))
	}

	// Run all jobs concurrently. If any job fails, context is expected to cancel.
	if err = simultaneously.DoCtx(ctx, maxConcurrentNotesFetches, callbacks...); err != nil {
		return nil, err
	}

	close(notesChannel)

	noteRegistry := make(NoteRecords, len(collection.Notes))

	for note := range notesChannel {

		id, ok := note["id"].(string)
		if !ok {
			return nil, errors.New("missing field 'id' in response for object 'notes'") // nolint:err113
		}

		noteRegistry[id] = note
	}

	return noteRegistry, nil
}

// fetchNote returns a concurrently executable job that fetches a single note
// by ID and sends the fully decoded NoteRecord to notesChannel.
func (c *Connector) fetchNote(notesChannel chan NoteRecord, noteID string,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		note, err := c.getNote(ctx, noteID)
		if err != nil {
			return err
		}

		select {
		case notesChannel <- *note:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Connector) getNote(ctx context.Context, noteID string) (*NoteRecord, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectNotes, noteID)
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	note, err := common.UnmarshalJSON[NoteRecord](resp)
	if err != nil {
		return nil, err
	}

	return note, nil
}

func lookupNote(notes NoteRecords, node *ajson.Node) (NoteRecord, string, error) {
	identifier, err := jsonquery.New(node).StringRequired("id")
	if err != nil {
		return nil, "", err
	}

	note, err := notes.findByID(identifier)
	if err != nil {
		return nil, "", err
	}

	return note, identifier, nil
}

func embedRaw(notes NoteRecords) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		note, _, err := lookupNote(notes, node)
		if err != nil {
			return nil, err
		}

		out := make(map[string]any, len(note))
		maps.Copy(out, note)

		return out, nil
	}
}

// embedNoteFields selects requested fields from the fetched full note.
func embedFields(notes NoteRecords) readhelper.SelectedFieldsFunc {
	return func(node *ajson.Node, fields []string) (map[string]any, string, error) {
		note, identifier, err := lookupNote(notes, node)
		if err != nil {
			return nil, "", err
		}

		selected := readhelper.SelectFields(note, datautils.NewSetFromList(fields))

		return selected, identifier, nil
	}
}

func (n NoteRecords) findByID(identifier string) (NoteRecord, error) {
	note, ok := n[identifier]
	if !ok || note == nil {
		return nil, fmt.Errorf( // nolint:err113
			"note with identifier %v was not found in the fetched registry", identifier)
	}

	return note, nil
}
