package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/ppanyukov/azure-sdk-for-go-hacks/sdk/azcore/policy/memo"
	"log"
	"os"
	"time"
)

// This example demonstrates how to use [memo.Memo] as a policy to cache
// responses from Azure SDK for Go. The output is like this:
//
//	2022/08/03 17:40:48 listWebSites: siteCount: 554 (elapsed: 12680ms)
//	2022/08/03 17:40:48 listWebSites: siteCount: 554 (elapsed: 170ms)
//	2022/08/03 17:40:49 listWebSites: siteCount: 554 (elapsed: 169ms)
//	2022/08/03 17:40:49 listWebSites: siteCount: 554 (elapsed: 173ms)
//
// Note that the unmarshalling of the response body is not cached, and it still
// happens every time.
func main() {
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		log.Fatal("Env var AZURE_SUBSCRIPTION_ID is not defined")
	}

	// Standard credentials
	var cred = func() azcore.TokenCredential {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		checkErr(err, "can't get credentials")
		return cred
	}()

	var clientOptionsWithCache = &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			PerCallPolicies: []policy.Policy{
				memo.NewMemo(memo.NoExpiration, memo.NoCleanup, nil),
			},
		},
	}

	client, err := armappservice.NewWebAppsClient(subscriptionID, cred, clientOptionsWithCache)
	checkErr(err, "armappservice.NewWebAppsClient")

	// First run is about 13-17s on my machine
	listWebSites(client)
	// Subsequent (cached) runs are about 170ms
	listWebSites(client)
	listWebSites(client)
	listWebSites(client)
}

func checkErr(err error, s string) {
	if err != nil {
		log.Fatalf("error: %s: %v", s, err)
	}
}

func listWebSites(client *armappservice.WebAppsClient) {
	start := time.Now()
	ctx := context.TODO()
	siteCount := 0

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		checkErr(err, "listWebSites.pager.NextPage")
		siteCount = siteCount + len(page.Value)
	}

	elapsed := time.Since(start).Milliseconds()
	log.Printf("listWebSites: siteCount: %d (elapsed: %dms)\n", siteCount, elapsed)
}
