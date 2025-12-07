package translations

import (
	"context"
	"errors"
	"fmt"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	createdID keystone.ID
}

// TranslatableItem is an entity that has translations
type TranslatableItem struct {
	keystone.BaseEntity
	Name        string
	Description keystone.Translations
}

func (d *Requirement) Name() string {
	return "Translations"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(TranslatableItem{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.read(actor),
		d.addTranslations(actor),
		d.readAfterAdd(actor),
		d.removeTranslation(actor),
		d.readAfterRemove(actor),
		d.replaceTranslations(actor),
		d.readAfterReplace(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	resp := requirements.TestResult{
		Name: "Create Entity with Translations",
	}

	item := &TranslatableItem{
		Name:        "Product",
		Description: keystone.Translations{},
	}

	// Set initial translations
	item.Description.Replace(map[string]*keystone.Translation{
		"en": keystone.NewTranslation("A great product"),
		"fr": keystone.NewTranslation("Un excellent produit"),
		"es": keystone.NewTranslation("Un gran producto"),
	})

	mutateErr := actor.Mutate(context.Background(), item)
	if mutateErr == nil {
		d.createdID = item.GetKeystoneID()
	} else {
		return resp.WithError(mutateErr)
	}

	return resp
}

func (d *Requirement) read(actor *keystone.Actor) requirements.TestResult {
	resp := requirements.TestResult{
		Name: "Read Translations",
	}

	item := &TranslatableItem{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(item, d.createdID), item, keystone.WithProperties("description"))

	if getErr != nil {
		return resp.WithError(getErr)
	}

	// Verify translations were stored correctly
	enText, ok := item.Description.Get("en")
	if !ok || enText.String() != "A great product" {
		return resp.WithError(fmt.Errorf("expected 'A great product' for 'en', got '%s' (ok: %v)", enText.String(), ok))
	}

	frText, ok := item.Description.Get("fr")
	if !ok || frText.String() != "Un excellent produit" {
		return resp.WithError(fmt.Errorf("expected 'Un excellent produit' for 'fr', got '%s' (ok: %v)", frText.String(), ok))
	}

	esText, ok := item.Description.Get("es")
	if !ok || esText.String() != "Un gran producto" {
		return resp.WithError(fmt.Errorf("expected 'Un gran producto' for 'es', got '%s' (ok: %v)", esText.String(), ok))
	}

	// Verify All() returns all translations
	all := item.Description.All()
	if len(all) != 3 {
		return resp.WithError(fmt.Errorf("expected 3 translations, got %d", len(all)))
	}

	return resp
}

func (d *Requirement) addTranslations(actor *keystone.Actor) requirements.TestResult {
	resp := requirements.TestResult{
		Name: "Add New Translations",
	}

	item := &TranslatableItem{}
	item.SetKeystoneID(d.createdID)

	// Add new translations
	item.Description.Add("de", "Ein großartiges Produkt")
	item.Description.Add("it", "Un ottimo prodotto")
	// Update existing translation
	item.Description.Add("en", "An amazing product")

	mutateErr := actor.Mutate(context.Background(), item, keystone.MutateProperties("description"))
	if mutateErr != nil {
		return resp.WithError(mutateErr)
	}

	return resp
}

func (d *Requirement) readAfterAdd(actor *keystone.Actor) requirements.TestResult {
	resp := requirements.TestResult{
		Name: "Read After Adding Translations",
	}

	item := &TranslatableItem{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(item, d.createdID), item, keystone.WithProperties("description"))

	if getErr != nil {
		return resp.WithError(getErr)
	}

	// Verify new translations were added
	deText, ok := item.Description.Get("de")
	if !ok || deText.String() != "Ein großartiges Produkt" {
		return resp.WithError(fmt.Errorf("expected 'Ein großartiges Produkt' for 'de', got '%s' (ok: %v)", deText.String(), ok))
	}

	itText, ok := item.Description.Get("it")
	if !ok || itText.String() != "Un ottimo prodotto" {
		return resp.WithError(fmt.Errorf("expected 'Un ottimo prodotto' for 'it', got '%s' (ok: %v)", itText.String(), ok))
	}

	// Verify existing translation was updated
	enText, ok := item.Description.Get("en")
	if !ok || enText.String() != "An amazing product" {
		return resp.WithError(fmt.Errorf("expected 'An amazing product' for 'en' (updated), got '%s' (ok: %v)", enText.String(), ok))
	}

	// Verify original translations still exist
	frText, ok := item.Description.Get("fr")
	if !ok || frText.String() != "Un excellent produit" {
		return resp.WithError(fmt.Errorf("expected 'Un excellent produit' for 'fr', got '%s' (ok: %v)", frText.String(), ok))
	}

	// Verify All() returns all 5 translations
	all := item.Description.All()
	if len(all) != 5 {
		return resp.WithError(fmt.Errorf("expected 5 translations after add, got %d: %v", len(all), all))
	}

	return resp
}

func (d *Requirement) removeTranslation(actor *keystone.Actor) requirements.TestResult {
	resp := requirements.TestResult{
		Name: "Remove Translation",
	}

	item := &TranslatableItem{}
	item.SetKeystoneID(d.createdID)

	// Remove a translation
	item.Description.Remove("es")

	mutateErr := actor.Mutate(context.Background(), item, keystone.MutateProperties("description"))
	if mutateErr != nil {
		return resp.WithError(mutateErr)
	}

	return resp
}

func (d *Requirement) readAfterRemove(actor *keystone.Actor) requirements.TestResult {
	resp := requirements.TestResult{
		Name: "Read After Removing Translation",
	}

	item := &TranslatableItem{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(item, d.createdID), item, keystone.WithProperties("description"))

	if getErr != nil {
		return resp.WithError(getErr)
	}

	// Verify removed translation is gone
	_, ok := item.Description.Get("es")
	if ok {
		return resp.WithError(errors.New("expected 'es' translation to be removed, but it still exists"))
	}

	// Verify other translations still exist
	enText, ok := item.Description.Get("en")
	if !ok || enText.String() != "An amazing product" {
		return resp.WithError(fmt.Errorf("expected 'An amazing product' for 'en', got '%s' (ok: %v)", enText.String(), ok))
	}

	// Verify All() returns 4 translations (5 - 1 removed)
	all := item.Description.All()
	if len(all) != 4 {
		return resp.WithError(fmt.Errorf("expected 4 translations after remove, got %d: %v", len(all), all))
	}

	return resp
}

func (d *Requirement) replaceTranslations(actor *keystone.Actor) requirements.TestResult {
	resp := requirements.TestResult{
		Name: "Replace All Translations",
	}

	item := &TranslatableItem{}
	item.SetKeystoneID(d.createdID)

	// Replace all translations with new ones
	item.Description.Replace(map[string]*keystone.Translation{
		"en": keystone.NewTranslation("New product description"),
		"ja": keystone.NewTranslation("新しい製品の説明"),
		"zh": keystone.NewTranslation("新产品描述"),
	})

	mutateErr := actor.Mutate(context.Background(), item, keystone.MutateProperties("description"))
	if mutateErr != nil {
		return resp.WithError(mutateErr)
	}

	return resp
}

func (d *Requirement) readAfterReplace(actor *keystone.Actor) requirements.TestResult {
	resp := requirements.TestResult{
		Name: "Read After Replacing Translations",
	}

	item := &TranslatableItem{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(item, d.createdID), item, keystone.WithProperties("description"))

	if getErr != nil {
		return resp.WithError(getErr)
	}

	// Verify old translations are gone
	_, ok := item.Description.Get("fr")
	if ok {
		return resp.WithError(errors.New("expected 'fr' translation to be replaced, but it still exists"))
	}

	_, ok = item.Description.Get("de")
	if ok {
		return resp.WithError(errors.New("expected 'de' translation to be replaced, but it still exists"))
	}

	// Verify new translations exist
	enText, ok := item.Description.Get("en")
	if !ok || enText.String() != "New product description" {
		return resp.WithError(fmt.Errorf("expected 'New product description' for 'en', got '%s' (ok: %v)", enText.String(), ok))
	}

	jaText, ok := item.Description.Get("ja")
	if !ok || jaText.String() != "新しい製品の説明" {
		return resp.WithError(fmt.Errorf("expected '新しい製品の説明' for 'ja', got '%s' (ok: %v)", jaText.String(), ok))
	}

	zhText, ok := item.Description.Get("zh")
	if !ok || zhText.String() != "新产品描述" {
		return resp.WithError(fmt.Errorf("expected '新产品描述' for 'zh', got '%s' (ok: %v)", zhText.String(), ok))
	}

	// Verify All() returns exactly 3 translations
	all := item.Description.All()
	if len(all) != 3 {
		return resp.WithError(fmt.Errorf("expected 3 translations after replace, got %d: %v", len(all), all))
	}

	return resp
}
