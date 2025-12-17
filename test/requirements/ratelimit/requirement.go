package ratelimit

import (
	"context"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
)

type Requirement struct {
}

func (d *Requirement) Name() string {
	return "Rate Limiter"
}

func (d *Requirement) Register(conn *keystone.Connection) error { return nil }

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.push(actor),
		d.quantity(actor, false),
	}
}

func (d *Requirement) push(actor *keystone.Actor) requirements.TestResult {

	result := requirements.TestResult{
		Name: "Basic Test",
	}

	testID := k4id.New().String()
	rl := actor.NewRateLimit(testID, 2, 2)

	resp, err := rl.Trigger(context.Background(), "trans1")
	if err != nil {
		return result.WithError(err)
	}
	if resp.CurrentCount() != 0 {
		return result.WithError(errors.New("current count is not 0 on the first call"))
	}

	resp, err = rl.Trigger(context.Background(), "trans2")
	if err != nil {
		return result.WithError(err)
	}
	if resp.CurrentCount() != 1 {
		return result.WithError(errors.New("current count is not 1"))
	}
	if resp.HitLimit() {
		return result.WithError(errors.New("hit limit on the second call"))
	}
	if resp.Percent() != 0.5 {
		return result.WithError(errors.New("percent is not 0.5"))
	}

	resp, err = rl.Trigger(context.Background(), "trans3")
	if err != nil {
		return result.WithError(err)
	}
	if resp.CurrentCount() != 2 {
		return result.WithError(errors.New("current count is not 2"))
	}
	if !resp.HitLimit() {
		return result.WithError(errors.New("hit limit expected on the third call"))
	}

	return result
}

func (d *Requirement) quantity(actor *keystone.Actor, concurrent bool) requirements.TestResult {
	result := requirements.TestResult{
		Name: "Quantity Test",
	}

	start := time.Now()
	checkpoint := time.Now()

	testID := k4id.New().String()
	rl := actor.NewTrackedRateLimit(testID, 2000, 2)

	if concurrent {
		wg := sync.WaitGroup{}
		for i := 0; i < 2000; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				rl.Trigger(context.Background(), k4id.New().String())
			}()
		}
		wg.Wait()
	} else {
		for i := 0; i < 2000; i++ {
			resp, err := rl.Trigger(context.Background(), k4id.New().String())
			if err != nil {
				return result.WithError(err)
			}
			if resp.HitLimit() {
				return result.WithError(errors.New("hit limit on call " + strconv.Itoa(i)))
			}
			if i%100 == 0 {
				log.Println("Per 100: ", time.Since(checkpoint))
				checkpoint = time.Now()
			}
		}
	}
	log.Println("Total: ", time.Since(start))

	resp, err := rl.Trigger(context.Background(), k4id.New().String())
	if err != nil {
		return result.WithError(err)
	}
	if !resp.HitLimit() {
		return result.WithError(errors.New("hit limit expected on call 2001"))
	}

	return result
}
