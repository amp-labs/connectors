package main

import (
	"context"
	"fmt"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/intercom"
	msTest "github.com/amp-labs/connectors/test/intercom"
	"github.com/amp-labs/connectors/test/utils"
)

type ArticlePayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorId    string `json:"author_id"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := msTest.GetIntercomConnector(ctx)

	fmt.Println("> TEST Create/Update/Delete Article")
	fmt.Println("Prepare by getting first admin user id")

	authorID := getAdminID(ctx, conn)

	fmt.Println("Creating Article")

	article := createArticle(ctx, conn, &ArticlePayload{
		Title:       "Famous quotes",
		Description: "To be, or not to be, that is the question. – William Shakespeare",
		AuthorId:    authorID,
	})

	fmt.Println("Updating description of an Article")
	updateArticle(ctx, conn, article.RecordId, &ArticlePayload{
		Title:       "Famous quotes",
		Description: "I think, therefore I am. – Rene Descartes",
		AuthorId:    authorID,
	})

	fmt.Println("View that article has changed accordingly")

	res := readArticles(ctx, conn)

	updatedArticle := searchArticles(res, "id", article.RecordId)
	for k, v := range map[string]string{
		"title":       "Famous quotes",
		"description": "I think, therefore I am. – Rene Descartes",
		"author_id":   authorID,
	} {
		if !compare(updatedArticle[k], v) {
			utils.Fail("error updated properties do not match", k, v, updatedArticle[k])
		}
	}

	fmt.Println("Removing this Article")
	removeArticle(ctx, conn, article.RecordId)
	fmt.Println("> Successful test completion")
}

func getAdminID(ctx context.Context, conn *intercom.Connector) string {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "admins",
		Fields:     connectors.Fields("id"),
	})
	if err != nil {
		utils.Fail("error reading from Intercom", "error", err)
	}

	if res.Rows < 1 {
		utils.Fail("expected to have at least one admin user")
	}

	userID, ok := res.Data[0].Fields["id"]
	if !ok {
		utils.Fail("user/admin response has no id field")
	}

	result, ok := userID.(string)
	if !ok {
		utils.Fail("user id is of an unexpected type")
	}

	return result
}

func searchArticles(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if compare(data.Fields[key], value) {
			return data.Raw
		}
	}

	utils.Fail("error finding article")

	return nil
}

func readArticles(ctx context.Context, conn *intercom.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "articles",
		Fields:     connectors.Fields("id", "title", "description", "author_id"),
	})
	if err != nil {
		utils.Fail("error reading from Intercom", "error", err)
	}

	return res
}

func createArticle(ctx context.Context, conn *intercom.Connector, payload *ArticlePayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "articles",
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Intercom", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a Article")
	}

	return res
}

func updateArticle(ctx context.Context, conn *intercom.Connector, articleID string, payload *ArticlePayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "articles",
		RecordId:   articleID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Intercom", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a Article")
	}

	return res
}

func removeArticle(ctx context.Context, conn *intercom.Connector, articleID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "articles",
		RecordId:   articleID,
	})
	if err != nil {
		utils.Fail("error deleting for Intercom", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a Article")
	}
}

func compare(field any, value string) bool {
	if len(value) == 0 && field == nil {
		return true
	}

	switch field.(type) {
	case float64:
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return false
		}

		return num == field.(float64)
	}

	return fmt.Sprintf("%v", field) == value
}
