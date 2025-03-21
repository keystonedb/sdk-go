package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

type PiiRegulation string

func (p PiiRegulation) String() string {
	return strings.ToUpper(strings.TrimSpace(string(p)))
}

const (
	RegulationGDPR PiiRegulation = "GDPR"
	RegulationCCPA PiiRegulation = "CCPA"
)

func (a *Actor) NewPiiToken(country string, regulation PiiRegulation) (string, error) {
	return a.NewPiiTokenWithExpiry(country, regulation, time.Time{})
}

func (a *Actor) NewGDPRToken(country string) (string, error) {
	return a.NewPiiTokenWithExpiry(country, RegulationGDPR, time.Time{})
}

func (a *Actor) NewCCPAToken() (string, error) {
	return a.NewPiiTokenWithExpiry("US:CA", RegulationCCPA, time.Time{})
}

func (a *Actor) NewPiiTokenWithExpiry(country string, regulation PiiRegulation, expiry time.Time) (string, error) {
	conn := a.Connection()
	req := &proto.PiiTokenRequest{
		Authorization: a.Authorization(),
		Country:       country, // COUNTRY[:STATE[:PROVINCE]]
		Regulation:    regulation.String(),
	}
	if !expiry.IsZero() {
		req.AutoExpire = timestamppb.New(expiry)
	}

	res, err := conn.PiiToken(context.Background(), req)
	if err != nil {
		return "", err
	}

	return res.GetToken(), err
}

func (a *Actor) Anonymize(piiToken string) (*proto.PiiAnonymizeResponse, error) {
	conn := a.Connection()
	req := &proto.PiiAnonymizeRequest{
		Authorization: a.Authorization(),
		Token:         piiToken,
		Rollback:      false,
	}

	return conn.PiiAnonymize(context.Background(), req)
}

func (a *Actor) AnonymizeRollback(piiToken string) (*proto.PiiAnonymizeResponse, error) {
	conn := a.Connection()
	req := &proto.PiiAnonymizeRequest{
		Authorization: a.Authorization(),
		Token:         piiToken,
		Rollback:      true,
	}

	return conn.PiiAnonymize(context.Background(), req)
}
