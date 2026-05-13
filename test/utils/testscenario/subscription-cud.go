package testscenario

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/test/utils"
)

type ConnectorSubscriptionManager interface {
	components.SubscriptionCreator
	components.SubscriptionUpdater
	components.SubscriptionRemover
}

// SubscriptionCreateUpdateDelete is a test scenario utilizing
// Subscribe/UpdateSubscription/DeleteSubscription connector operations.
// Each step will be displayed on the screen and to be analyzed by developer.
func SubscriptionCreateUpdateDelete(
	ctx context.Context, conn ConnectorSubscriptionManager,
	createParams, updateParams SubscribeParamBuilder,
) {
	fmt.Println("> TEST Subscription Create/Update/Delete")
	publicURL, ok := getPublicWebhookURL(ctx)
	if !ok {
		failOnError(errors.New("webhook URL is needed"))
	}

	fmt.Println("============= Create =============")
	result, err := conn.Subscribe(ctx, *createParams(publicURL))
	if err != nil {
		fmt.Println("conn.Subscribe() -> failed")
		failOnError(err)
	}
	validateSubscriptionResult(result)

	fmt.Println("============= Update =============")
	result, err = conn.UpdateSubscription(ctx, *updateParams(publicURL), result)
	if err != nil {
		fmt.Println("conn.UpdateSubscription() -> failed")
		failOnError(err)
	}
	validateSubscriptionResult(result)

	fmt.Println("============= Delete =============")
	err = conn.DeleteSubscription(ctx, *result)
	if err != nil {
		fmt.Println("conn.DeleteSubscription() -> failed")
		failOnError(err)
	}

	fmt.Println("> Successful test completion")
}

func validateSubscriptionResult(result *common.SubscriptionResult) {
	fmt.Println("(1) Result:")
	utils.DumpJSON(result.Result, os.Stdout)
	fmt.Println("(2) ObjectEvents:")
	utils.DumpJSON(result.ObjectEvents, os.Stdout)
	fmt.Printf("(3) Status: \"%v\"\n", result.Status)
	if result.Status != common.SubscriptionStatusSuccess {
		failOnError(errors.New("subscription has not succeeded"))
	}
}
