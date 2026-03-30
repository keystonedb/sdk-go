package find

import (
	"context"
	"errors"
	"fmt"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	createdID    keystone.ID
	controlValue string
	secretValue  keystone.SecureString
}

func (d *Requirement) Name() string {
	return "Find with SecureString"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.DataTypes{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	d.controlValue = "controlled"
	d.secretValue = keystone.NewSecureString("secretval", "secre***")

	return []requirements.TestResult{
		d.create(actor),
		d.findWithoutDecryption(actor),
		d.findWithDecryption(actor),
		d.findWithDecryptionNamed(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create"}

	entity := &models.DataTypes{
		String: d.controlValue,
		Secret: d.secretValue,
	}

	err := actor.Mutate(context.Background(), entity, keystone.WithMutationComment("Create entity for find secure test"))
	if err == nil {
		d.createdID = entity.GetKeystoneID()
	}
	return res.WithError(err)
}

func (d *Requirement) findWithDecryption(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Find with Decrypted SecureString"}

	results, err := actor.Find(
		context.Background(),
		keystone.Type(models.DataTypes{}),
		keystone.WithDecryptedProperties(),
		keystone.WhereIn("_entity_id", d.createdID.String()),
	)
	if err != nil {
		return res.WithError(err)
	}

	if len(results) == 0 {
		return res.WithError(errors.New("no results found"))
	}

	var found []models.DataTypes
	if err = keystone.UnmarshalToSlice(&found, results...); err != nil {
		return res.WithError(fmt.Errorf("unmarshal error: %w", err))
	}

	for _, dt := range found {
		if dt.GetKeystoneID() == d.createdID {
			if dt.String != d.controlValue {
				return res.WithError(fmt.Errorf("control string mismatch: got %q, want %q", dt.String, d.controlValue))
			}
			if dt.Secret.Original != d.secretValue.Original {
				return res.WithError(fmt.Errorf("decrypted original mismatch: got %q, want %q", dt.Secret.Original, d.secretValue.Original))
			}
			if dt.Secret.Masked != d.secretValue.Masked {
				return res.WithError(fmt.Errorf("masked mismatch: got %q, want %q", dt.Secret.Masked, d.secretValue.Masked))
			}
			return res
		}
	}

	return res.WithError(errors.New("created entity not found in results"))
}

func (d *Requirement) findWithDecryptionNamed(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Find with Named Decrypted SecureString"}

	results, err := actor.Find(
		context.Background(),
		keystone.Type(models.DataTypes{}),
		keystone.RetrieveOptions(
			keystone.WithProperties(),
			keystone.WithDecryptedProperties("secret"),
		),
		keystone.WhereIn("_entity_id", d.createdID.String()),
	)
	if err != nil {
		return res.WithError(err)
	}

	if len(results) == 0 {
		return res.WithError(errors.New("no results found"))
	}

	var found []models.DataTypes
	if err = keystone.UnmarshalToSlice(&found, results...); err != nil {
		return res.WithError(fmt.Errorf("unmarshal error: %w", err))
	}

	for _, dt := range found {
		if dt.GetKeystoneID() == d.createdID {
			if dt.String != d.controlValue {
				return res.WithError(fmt.Errorf("control string mismatch: got %q, want %q", dt.String, d.controlValue))
			}
			if dt.Secret.Original != d.secretValue.Original {
				return res.WithError(fmt.Errorf("decrypted original mismatch: got %q, want %q", dt.Secret.Original, d.secretValue.Original))
			}
			if dt.Secret.Masked != d.secretValue.Masked {
				return res.WithError(fmt.Errorf("masked mismatch: got %q, want %q", dt.Secret.Masked, d.secretValue.Masked))
			}
			return res
		}
	}

	return res.WithError(errors.New("created entity not found in results"))
}

func (d *Requirement) findWithoutDecryption(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Find without Decryption has no Original"}

	results, err := actor.Find(
		context.Background(),
		keystone.Type(models.DataTypes{}),
		keystone.WithProperties(),
		keystone.WhereIn("_entity_id", d.createdID.String()),
	)
	if err != nil {
		return res.WithError(err)
	}

	if len(results) == 0 {
		return res.WithError(errors.New("no results found"))
	}

	var found []models.DataTypes
	if err = keystone.UnmarshalToSlice(&found, results...); err != nil {
		return res.WithError(fmt.Errorf("unmarshal error: %w", err))
	}

	for _, dt := range found {
		if dt.GetKeystoneID() == d.createdID {
			if dt.String != d.controlValue {
				return res.WithError(fmt.Errorf("control string mismatch: got %q, want %q", dt.String, d.controlValue))
			}
			if dt.Secret.Original != "" {
				return res.WithError(fmt.Errorf("expected empty original without decryption, got %q", dt.Secret.Original))
			}
			if dt.Secret.Masked != d.secretValue.Masked {
				return res.WithError(fmt.Errorf("masked mismatch: got %q, want %q", dt.Secret.Masked, d.secretValue.Masked))
			}
			return res
		}
	}

	return res.WithError(errors.New("created entity not found in results"))
}
