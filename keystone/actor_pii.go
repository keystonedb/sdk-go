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

func (a *Actor) NewPiiToken(reference, country string, regulation PiiRegulation) (string, error) {
	return a.NewPiiTokenWithExpiry(reference, country, regulation, time.Time{})
}

func (a *Actor) NewGDPRToken(reference, country string) (string, error) {
	return a.NewPiiTokenWithExpiry(reference, country, RegulationGDPR, time.Time{})
}

func (a *Actor) NewCCPAToken(reference string) (string, error) {
	return a.NewPiiTokenWithExpiry(reference, "US:CA", RegulationCCPA, time.Time{})
}

func (a *Actor) NewPiiTokenWithExpiry(reference, country string, regulation PiiRegulation, expiry time.Time) (string, error) {
	conn := a.Connection()
	req := &proto.PiiTokenRequest{
		Authorization: a.Authorization(),
		Reference:     reference,
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
