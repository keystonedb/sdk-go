package main

import (
	"log"

	"github.com/packaged/environment/environment"
	"github.com/packaged/logger/v3/logger"
)

const (
	Check = "✅ "
	Cross = "❌ "
)

func init() {
	logger.Setup(environment.Development)
}

func main() {
	totalRun := 0
	totalFailures := 0
	var failedRequirements []string

	accessToken := "test-access-token"

	kHost := "127.0.0.1"
	kPort := "50051"
	InitKeystone(kHost, kPort, accessToken)
	defer CloseKeystone()

	actor := Actor()

	log.Println("Running Requirements Test - Against " + kHost + ":" + kPort)
	for _, req := range reqs {
		totalRun++
		log.Println("")
		log.Println("Verifying", req.Name())
		regErr := req.Register(keystoneConnection)
		if regErr != nil {
			failedRequirements = append(failedRequirements, req.Name()+" : Registration")
			log.Println(Cross, " Registration Failed", regErr.Error())
			totalFailures++
			continue
		}
		keystoneConnection.SyncSchema().Wait()

		results := req.Verify(actor)
		for _, result := range results {
			if result.Error != nil {
				log.Println(Cross, result.Name, "-", result.Error.Error())
				failedRequirements = append(failedRequirements, req.Name()+" - "+result.Name)
				totalFailures++
			} else {
				log.Println(Check, result.Name)
			}
		}
	}

	log.Println("")
	log.Printf("Total requirements run: %d", totalRun)
	log.Printf("Total requirements failed: %d", totalFailures)
	if totalFailures > 0 {
		log.Println("Failed requirements:")
		for _, req := range failedRequirements {
			log.Println("-", req)
		}
	}
	log.Println("")

}
