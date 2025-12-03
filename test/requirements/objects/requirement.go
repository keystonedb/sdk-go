package objects

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	createdID keystone.ID
}

func (d *Requirement) Name() string {
	return "Entity Objects"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.upload(actor),
		d.list(actor),
		d.byPath(actor),
	}
}

func (d *Requirement) upload(actor *keystone.Actor) requirements.TestResult {

	psn := &models.Person{
		BaseEntity: keystone.BaseEntity{},
		Name:       "Upload",
	}

	fileOne := keystone.NewUpload("profile.png", proto.ObjectType_Standard)
	fileOne.SetExpiry(time.Now().Add(time.Second * 60))
	fileTwo := keystone.NewUpload("policy.txt", proto.ObjectType_NearLine)
	fileThree := keystone.NewUpload("public.pdf", proto.ObjectType_Standard)
	fileThree.SetData([]byte("file contents here"))
	fileRemote, err := keystone.NewUploadFromURL("README.md", "https://raw.githubusercontent.com/keystonedb/sdk-go/refs/heads/main/README.md", proto.ObjectType_Standard)
	if err != nil {
		return requirements.TestResult{
			Name:  "Upload",
			Error: errors.New("unable to upload from URL"),
		}
	}
	fileRemote.SetPublic(true)

	createErr := actor.Mutate(context.Background(), psn, keystone.PrepareUploads(fileOne, fileTwo, fileThree, fileRemote))
	if createErr == nil {
		d.createdID = psn.GetKeystoneID()

		if !fileOne.ReadyForUpload() {
			return requirements.TestResult{
				Name:  "Upload",
				Error: errors.New("no signed url was created for the upload"),
			}
		} else {
			resp, err := fileOne.Upload(bytes.NewBuffer([]byte("file contents")))
			if err != nil {
				createErr = err
			} else {
				if resp.StatusCode != 200 {
					createErr = errors.New("upload failed, status code: " + string(rune(resp.StatusCode)))
					bdy, _ := io.ReadAll(resp.Body)
					fmt.Println(string(bdy))
				}
			}
		}

		if createErr == nil && !fileTwo.ReadyForUpload() {
			return requirements.TestResult{
				Name:  "Upload",
				Error: errors.New("no signed url was created for the upload"),
			}
		} else if createErr == nil {
			resp, err := fileTwo.Upload(bytes.NewBuffer([]byte("policy document")))
			if err != nil {
				createErr = err
			} else {
				if resp.StatusCode != 200 {
					createErr = errors.New("upload failed, status code: " + string(rune(resp.StatusCode)))
					bdy, _ := io.ReadAll(resp.Body)
					fmt.Println(string(bdy))
				}
			}
		}
	}

	return requirements.TestResult{
		Name:  "Upload",
		Error: createErr,
	}
}

func (d *Requirement) list(actor *keystone.Actor) requirements.TestResult {

	psn := &models.Person{}
	listErr := actor.Get(context.Background(), keystone.ByEntityID(psn, d.createdID), psn, keystone.WithObjects())

	if listErr == nil {
		o := psn.GetObject("README.md")
		log.Println(o.GetUrl())
		if len(psn.GetObjects()) != 4 {
			listErr = errors.New("object count is not 3, got " + string(len(psn.GetObjects())))
		} else if obj := psn.GetObject("profile.png"); obj == nil {
			listErr = errors.New("object not found")
		} else {
			if obj.GetUrl() == "" {
				listErr = errors.New("object url is empty")
			}
		}
	}

	return requirements.TestResult{
		Name:  "List",
		Error: listErr,
	}
}
func (d *Requirement) byPath(actor *keystone.Actor) requirements.TestResult {
	psn := &models.Person{}
	byPathErr := actor.Get(context.Background(), keystone.ByEntityID(psn, d.createdID), psn, keystone.WithObjects("profile.png"))

	if byPathErr == nil {
		if obj := psn.GetObject("profile.png"); obj == nil {
			byPathErr = errors.New("object not found")
		} else {
			if obj.GetUrl() == "" {
				byPathErr = errors.New("object url is empty")
			}
		}
	}

	return requirements.TestResult{
		Name:  "Object By Path",
		Error: byPathErr,
	}
}
