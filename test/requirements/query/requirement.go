package query

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
)

type Requirement struct {
	runID          string
	subscriptionID keystone.ID
	renewalIDs     []string
}

func (d *Requirement) Name() string {
	return "Query Entity Index"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	d.runID = k4id.New().String()
	conn.RegisterTypes(models.FileData{}, models.Subscription{}, models.Renewal{}, models.Offer{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	report(d.create(actor))
	report(d.readPending(actor))
	report(d.readOneTwo(actor))
	report(d.readThree(actor))
	report(d.readComplex(actor))
	report(d.createOffers(actor))
	report(d.readOfferByProduct(actor))
	report(d.readOfferByGlobalOrProduct(actor))
	report(d.readOfferByProductsIn(actor))
	report(d.readOfferByProductsNotIn(actor))
	report(d.createChildEntities(actor))
	report(d.readByEntityIDs(actor))
	report(d.readByEntityIDsAndParent(actor))
}

func (d *Requirement) readOneTwo(actor *keystone.Actor) requirements.TestResult {

	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.FileData{}),
		[]string{"check_key", "line_information"}, keystone.Limit(2, 0), keystone.SortBy("modified", true),
		keystone.WhereIn("state", 1, 2))

	if err == nil && len(entities) < 2 {
		err = errors.New("not enough entities returned")
	}

	for _, entity := range entities {
		file := &models.FileData{}
		unErr := keystone.Unmarshal(entity, file)
		if unErr != nil {
			err = unErr
			break
		}

		if file.CheckKey != d.runID {
			err = errors.New("incorrect check key - " + file.CheckKey + " != " + d.runID)
			break
		}

		if file.LineInformation == "" {
			err = errors.New("line information is empty")
			break
		}
	}

	return requirements.TestResult{
		Name:  "readOneTwo",
		Error: err,
	}
}
func (d *Requirement) readThree(actor *keystone.Actor) requirements.TestResult {

	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.FileData{}),
		[]string{"check_key"}, keystone.Limit(2, 0), keystone.SortBy("modified", true),
		keystone.WhereEquals("connector_id", "tester"))
	if err == nil && len(entities) < 2 {
		err = errors.New("not enough entities returned")
	}

	for _, entity := range entities {
		file := &models.FileData{}
		unErr := keystone.Unmarshal(entity, file)
		if unErr != nil {
			err = unErr
			break
		}

		if file.CheckKey != d.runID {
			err = errors.New("incorrect check key - " + file.CheckKey + " != " + d.runID)
			break
		}
	}

	return requirements.TestResult{
		Name:  "readThree",
		Error: err,
	}
}

func (d *Requirement) readComplex(actor *keystone.Actor) requirements.TestResult {

	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.FileData{}),
		[]string{"check_key", "identifier"}, keystone.Limit(2, 0), keystone.SortBy("modified", true),
		keystone.WhereEquals("user_id", "usr1"),
		keystone.Or(keystone.WhereEquals("state", 1), keystone.WhereEquals("state", 2)))
	if err == nil && len(entities) < 2 {
		err = errors.New("not enough entities returned")
	}

	expect := map[string]bool{
		"fd0": false,
		"fd1": false,
	}

	for _, entity := range entities {
		file := &models.FileData{}
		unErr := keystone.Unmarshal(entity, file)
		if unErr != nil {
			err = unErr
			break
		}

		if file.CheckKey != d.runID {
			err = errors.New("incorrect check key - " + file.CheckKey + " != " + d.runID)
			break
		}

		expect[file.Identifier] = true
	}

	for k, v := range expect {
		if !v {
			err = errors.New("expected identifier not found - " + k)
			break
		}
	}

	return requirements.TestResult{
		Name:  "readComplex",
		Error: err,
	}
}

func (d *Requirement) readPending(actor *keystone.Actor) requirements.TestResult {

	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.FileData{}),
		[]string{"check_key"}, keystone.Limit(3, 0), keystone.SortBy("modified", true),
		keystone.WhereEquals("is_pending", true))

	if err == nil && len(entities) < 3 {
		err = errors.New("not enough entities returned")
	}

	for _, entity := range entities {
		file := &models.FileData{}
		unErr := keystone.Unmarshal(entity, file)
		if unErr != nil {
			err = unErr
			break
		}

		if file.CheckKey != d.runID {
			err = errors.New("incorrect check key - " + file.CheckKey + " != " + d.runID)
			break
		}
	}

	return requirements.TestResult{
		Name:  "readPending",
		Error: err,
	}
}
func (d *Requirement) createChildEntities(actor *keystone.Actor) requirements.TestResult {
	sub := &models.Subscription{
		StartDate: time.Now(),
	}

	createErr := actor.Mutate(context.Background(), sub, keystone.WithMutationComment("Create parent subscription for query test"))
	if createErr != nil {
		return requirements.TestResult{
			Name:  "createChildEntities",
			Error: createErr,
		}
	}
	d.subscriptionID = sub.GetKeystoneID()

	for i := 0; i < 3; i++ {
		renewal := &models.Renewal{
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 1, 0),
			CreationDate: time.Now(),
			PaymentDate:  time.Now(),
		}
		renewal.SetKeystoneID(d.subscriptionID)

		createErr = actor.Mutate(context.Background(), renewal, keystone.WithMutationComment("Create renewal "+strconv.Itoa(i)))
		if createErr != nil {
			return requirements.TestResult{
				Name:  "createChildEntities",
				Error: createErr,
			}
		}
		d.renewalIDs = append(d.renewalIDs, renewal.GetKeystoneID().ChildID())
	}

	return requirements.TestResult{
		Name: "createChildEntities",
	}
}

func (d *Requirement) readByEntityIDs(actor *keystone.Actor) requirements.TestResult {
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.Renewal{}),
		[]string{}, keystone.Limit(10, 0),
		keystone.WithEntityIDs(d.renewalIDs))

	if err == nil && len(entities) != 3 {
		err = errors.New("expected 3 renewals, got " + strconv.Itoa(len(entities)))
	}

	return requirements.TestResult{
		Name:  "readByEntityIDs",
		Error: err,
	}
}

func (d *Requirement) readByEntityIDsAndParent(actor *keystone.Actor) requirements.TestResult {
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.Renewal{}),
		[]string{}, keystone.Limit(10, 0),
		keystone.WithEntityIDs(d.renewalIDs),
		keystone.ChildOf(d.subscriptionID.String()))

	if err == nil && len(entities) != 3 {
		err = errors.New("expected 3 renewals, got " + strconv.Itoa(len(entities)))
	}

	return requirements.TestResult{
		Name:  "readByEntityIDsAndParent",
		Error: err,
	}
}

func (d *Requirement) createOffers(actor *keystone.Actor) requirements.TestResult {
	offers := []*models.Offer{
		{
			DisplayName: "offer-simple-" + d.runID,
			TestRun:     d.runID,
			ProductIDs:  []string{d.offerProductID("101")},
		},
		{
			DisplayName: "offer-global-" + d.runID,
			TestRun:     d.runID,
			IsGlobal:    true,
		},
		{
			DisplayName: "offer-targeted-" + d.runID,
			TestRun:     d.runID,
			ProductIDs:  []string{d.offerProductID("102")},
		},
		{
			DisplayName: "offer-unrelated-" + d.runID,
			TestRun:     d.runID,
			ProductIDs:  []string{d.offerProductID("999")},
		},
	}

	for _, offer := range offers {
		if err := actor.Mutate(context.Background(), offer, keystone.WithMutationComment("Create offer for query contains test")); err != nil {
			return requirements.TestResult{
				Name:  "createOffers",
				Error: err,
			}
		}
	}

	return requirements.TestResult{Name: "createOffers"}
}

func (d *Requirement) readOfferByProduct(actor *keystone.Actor) requirements.TestResult {
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.Offer{}),
		[]string{"display_name", "test_run", "is_global", "product_ids"},
		keystone.WhereEquals("test_run", d.runID),
		keystone.WhereContains("product_ids", d.offerProductID("101")))
	if err == nil && len(entities) != 1 {
		err = fmt.Errorf("expected 1 offer, got %d", len(entities))
	}

	if err == nil {
		offer := &models.Offer{}
		if unmarshalErr := keystone.Unmarshal(entities[0], offer); unmarshalErr != nil {
			err = unmarshalErr
		} else if offer.DisplayName != "offer-simple-"+d.runID {
			err = fmt.Errorf("unexpected offer returned: %s", offer.DisplayName)
		} else if offer.IsGlobal {
			err = errors.New("expected non-global offer for product contains query")
		} else if !containsString(offer.ProductIDs, d.offerProductID("101")) {
			err = fmt.Errorf("expected product id %s in result", d.offerProductID("101"))
		}
	}

	return requirements.TestResult{
		Name:  "readOfferByProduct",
		Error: err,
	}
}

func (d *Requirement) readOfferByGlobalOrProduct(actor *keystone.Actor) requirements.TestResult {
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.Offer{}),
		[]string{"display_name", "test_run", "is_global", "product_ids"},
		keystone.WhereEquals("test_run", d.runID),
		keystone.Or(
			keystone.WhereEquals("is_global", true),
			keystone.WhereContains("product_ids", d.offerProductID("102")),
		))
	if err == nil && len(entities) != 2 {
		err = fmt.Errorf("expected 2 offers, got %d", len(entities))
	}

	expected := map[string]bool{
		"offer-global-" + d.runID:   false,
		"offer-targeted-" + d.runID: false,
	}

	if err == nil {
		for _, entity := range entities {
			offer := &models.Offer{}
			if unmarshalErr := keystone.Unmarshal(entity, offer); unmarshalErr != nil {
				err = unmarshalErr
				break
			}

			matched, ok := expected[offer.DisplayName]
			if !ok {
				err = fmt.Errorf("unexpected offer returned: %s", offer.DisplayName)
				break
			}
			if matched {
				err = fmt.Errorf("duplicate offer returned: %s", offer.DisplayName)
				break
			}
			expected[offer.DisplayName] = true
		}
	}

	if err == nil {
		for name, found := range expected {
			if !found {
				err = fmt.Errorf("expected offer not returned: %s", name)
				break
			}
		}
	}

	return requirements.TestResult{
		Name:  "readOfferByGlobalOrProduct",
		Error: err,
	}
}

func (d *Requirement) readOfferByProductsIn(actor *keystone.Actor) requirements.TestResult {
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.Offer{}),
		[]string{"display_name", "test_run", "is_global", "product_ids"},
		keystone.WhereEquals("test_run", d.runID),
		keystone.WhereIn("product_ids", d.offerProductID("102"), d.offerProductID("201")))

	if err == nil && len(entities) != 1 {
		err = fmt.Errorf("expected 1 offer overlapping product_ids, got %d", len(entities))
	}

	if err == nil {
		offer := &models.Offer{}
		if unmarshalErr := keystone.Unmarshal(entities[0], offer); unmarshalErr != nil {
			err = unmarshalErr
		} else if offer.DisplayName != "offer-targeted-"+d.runID {
			err = fmt.Errorf("unexpected offer returned: %s", offer.DisplayName)
		} else if !containsString(offer.ProductIDs, d.offerProductID("102")) {
			err = fmt.Errorf("expected product id %s in result", d.offerProductID("102"))
		}
	}

	return requirements.TestResult{
		Name:  "readOfferByProductsIn",
		Error: err,
	}
}

func (d *Requirement) readOfferByProductsNotIn(actor *keystone.Actor) requirements.TestResult {
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.Offer{}),
		[]string{"display_name", "test_run", "is_global", "product_ids"},
		keystone.WhereEquals("test_run", d.runID),
		keystone.WhereNotIn("product_ids", d.offerProductID("101"), d.offerProductID("999")))

	excluded := map[string]bool{
		"offer-simple-" + d.runID:    true,
		"offer-unrelated-" + d.runID: true,
	}
	requiredFound := false

	if err == nil {
		for _, entity := range entities {
			offer := &models.Offer{}
			if unmarshalErr := keystone.Unmarshal(entity, offer); unmarshalErr != nil {
				err = unmarshalErr
				break
			}

			if excluded[offer.DisplayName] {
				err = fmt.Errorf("offer %s should have been excluded by NOT IN overlap", offer.DisplayName)
				break
			}

			if offer.DisplayName == "offer-targeted-"+d.runID {
				requiredFound = true
			}
		}
	}

	if err == nil && !requiredFound {
		err = errors.New("expected offer-targeted in NOT IN overlap result")
	}

	return requirements.TestResult{
		Name:  "readOfferByProductsNotIn",
		Error: err,
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	files := []*models.FileData{
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           1,
			ConnectorID:     "connector-1",
			IsPending:       false,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd0",
		},
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           2,
			ConnectorID:     "connector-1",
			IsPending:       false,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd1",
		},
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           3,
			ConnectorID:     "connector-1",
			IsPending:       false,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd2",
		},
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           3,
			ConnectorID:     "connector-1",
			IsPending:       true,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd3",
		},
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           3,
			ConnectorID:     "tester",
			IsPending:       true,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd4",
		},
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           3,
			ConnectorID:     "tester",
			IsPending:       true,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd5",
		},
	}

	var createErr error
	for _, file := range files {
		createErr = actor.Mutate(context.Background(), file, keystone.WithMutationComment("Create a file"))
		if createErr != nil {
			return requirements.TestResult{
				Name:  "Create",
				Error: createErr,
			}
		}
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (d *Requirement) offerProductID(suffix string) string {
	return "prod-" + suffix + "-" + d.runID
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
