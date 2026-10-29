package main

import (
	"flag"
	"log"
	"sync"
	"time"

	"github.com/keystonedb/sdk-go/test/requirements"
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
	parallelism := flag.Int("parallel", 4, "Maximum concurrently active opted-in requirements (1 disables overlap)")
	flag.Parse()
	if *parallelism < 1 {
		log.Fatal("-parallel must be at least 1")
	}
	var reportMu sync.Mutex
	totalRun := 0
	totalFailures := 0
	var failedRequirements []string

	accessToken := "test-access-token"

	kHost := "127.0.0.1"
	kPort := "50051"
	InitKeystone(kHost, kPort, accessToken)
	defer CloseKeystone()

	scheduler := requirements.NewScheduler()
	scheduler.SetParallelism(*parallelism)
	started := time.Now()
	var registrationTime time.Duration
	registrationByName := map[string]time.Duration{}
	log.Println("Running Requirements Test - Against " + kHost + ":" + kPort)
	for _, req := range reqs {
		totalRun++
		registrationStarted := time.Now()
		regErr := req.Register(keystoneConnection)
		if regErr == nil {
			keystoneConnection.SyncSchema().Wait()
		}
		duration := time.Since(registrationStarted)
		registrationTime += duration
		registrationByName[req.Name()] += duration
		if regErr != nil {
			failedRequirements = append(failedRequirements, req.Name()+" : Registration")
			log.Println(Cross, "Registration Failed", req.Name(), regErr)
			totalFailures++
			continue
		}
		actor := Actor()
		reporter := func(result requirements.TestResult) {
			reportMu.Lock()
			defer reportMu.Unlock()
			if result.Error != nil {
				log.Println(Cross, req.Name()+" / "+result.Name, "-", result.Error)
				failedRequirements = append(failedRequirements, req.Name()+" - "+result.Name)
				totalFailures++
			} else {
				log.Println(Check, req.Name()+" / "+result.Name)
			}
		}
		verify := func(scoped *requirements.Scheduler) {
			log.Println("Verifying", req.Name())
			if phased, ok := req.(requirements.ScheduledRequirement); ok {
				phased.VerifyScheduled(actor, reporter, scoped)
			} else {
				req.Verify(actor, reporter)
			}
		}
		_, early := req.(requirements.ScheduledRequirement)
		parallel, ok := req.(requirements.ParallelRequirement)
		if ok && parallel.Parallel() {
			if early {
				scheduler.StartEarlyParallel(req.Name(), verify)
			} else {
				scheduler.StartParallel(req.Name(), verify)
			}
		} else {
			if early {
				scheduler.StartEarly(req.Name(), func() { verify(scheduler) })
			} else {
				scheduler.Start(req.Name(), func() { verify(scheduler) })
			}
		}
	}

	timing := scheduler.Run()

	log.Println("")
	for _, req := range reqs {
		log.Printf("%s effective time: %v", req.Name(), registrationByName[req.Name()]+scheduler.Active(req.Name()))
	}
	log.Println("")
	log.Printf("Total requirements run: %d", totalRun)
	log.Printf("Total requirements failed: %d", totalFailures)
	log.Printf("Total Duration: %v", time.Since(started))
	log.Printf("Effective test time (including RPC latency): %v", registrationTime+timing.Effective)
	log.Printf("Aggregate active work: %v", registrationTime+timing.Work)
	log.Printf("Idle waiting time: %v", timing.Idle)
	if totalFailures > 0 {
		log.Println("Failed requirements:")
		for _, req := range failedRequirements {
			log.Println("-", req)
		}
	}
	log.Println("")

}
