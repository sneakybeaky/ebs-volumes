package shared

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const instanceIdentityDocument = `{
  "devpayProductCodes" : null,
  "availabilityZone" : "us-east-1d",
  "privateIp" : "10.158.112.84",
  "version" : "2010-08-31",
  "region" : "us-east-1",
  "instanceId" : "i-1234567890abcdef0",
  "billingProducts" : null,
  "instanceType" : "t1.micro",
  "accountId" : "123456789012",
  "pendingTime" : "2015-11-19T16:32:11Z",
  "imageId" : "ami-5fb8c835",
  "kernelId" : "aki-919dcaf8",
  "ramdiskId" : null,
  "architecture" : "x86_64"
}`

func initTestServer(path string, resp string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != path {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Write([]byte(resp))
	}))
}

func TestHappyPath(t *testing.T) {
	server := initTestServer(
		"/latest/dynamic/instance-identity/document",
		instanceIdentityDocument,
	)
	defer server.Close()

	undertest := NewEC2Instance(unit.Session, &aws.Config{Endpoint: aws.String(server.URL + "/latest")})

	instanceid, err := undertest.InstanceID()
	assert.Nil(t, err, "Expect no error, %v", err)
	assert.Equal(t, instanceid, "i-1234567890abcdef0")
}
