package api

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gocraft/web"
	"net/http"
	"strings"
)

const awsRegion = "us-west-2"

var vpcFilterName = "vpc-id"

type ApiContext struct {
	*Context
	AwsCredentials *credentials.Credentials
}

var apiRouter *web.Router

func init() {
	apiRouter = router.Subrouter(ApiContext{}, "/")
	apiRouter.Middleware((*ApiContext).ValidateAwsCredentials)
	apiRouter.Get("/", (*ApiContext).Root)
	apiRouter.Get("/vpc/:id/instances", (*ApiContext).VpcListInstances)
	apiRouter.Delete("/instances/:id", (*ApiContext).TerminateInstance)
	apiRouter.Put("/instances/:id/clone", (*ApiContext).CloneInstance)
}

func (c *ApiContext) Root(rw web.ResponseWriter, r *web.Request) {
	creds, err := c.AwsCredentials.Get()
	if err != nil {
		panic(err)
	}

	writeJson(rw, map[string]string{
		"status":        "ok",
		"access_key_id": creds.AccessKeyID,
	})
}

func (c *ApiContext) VpcListInstances(rw web.ResponseWriter, r *web.Request) {
	if ok := validatePresencePathParams(r.PathParams, "id"); !ok {
		rw.WriteHeader(http.StatusBadRequest)
		writeJson(rw, map[string]string{
			"error": "no vpc id given",
		})
		return
	}

	vpcId := fmt.Sprintf("vpc-%s", strings.ToLower(r.PathParams["id"]))
	ec2service := ec2.New(&aws.Config{
		Credentials: c.AwsCredentials,
		Region:      awsRegion,
	})

	vpcParams := &ec2.DescribeVPCsInput{}
	vpcs, err := ec2service.DescribeVPCs(vpcParams)
	if err != nil {
		panic(err)
	}

	foundVpc := false
	for _, v := range vpcs.VPCs {
		if vpcId == *v.VPCID {
			foundVpc = true
		}
	}
	if !foundVpc {
		rw.WriteHeader(http.StatusNotFound)
		writeJson(rw, map[string]string{
			"error": "no vpc with that id",
		})
		return
	}

	instanceFilters := []*ec2.Filter{
		&ec2.Filter{
			Name:   &vpcFilterName,
			Values: []*string{&vpcId},
		},
	}
	instanceParams := &ec2.DescribeInstancesInput{
		Filters: instanceFilters,
	}
	instances, err := ec2service.DescribeInstances(instanceParams)
	if err != nil {
		panic(err)
	}

	writeJson(rw, map[string]string{
		"instances": awsutil.StringValue(instances),
	})
}

func (c *ApiContext) TerminateInstance(rw web.ResponseWriter, r *web.Request) {
	writeJson(rw, map[string]string{
		"status": "ok",
	})
}

func (c *ApiContext) CloneInstance(rw web.ResponseWriter, r *web.Request) {
	writeJson(rw, map[string]string{
		"status": "ok",
	})
}

func (c *ApiContext) ValidateAwsCredentials(rw web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	if ok := validatePresenceRequest(r, "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"); !ok {
		rw.WriteHeader(http.StatusBadRequest)
		writeJson(rw, map[string]string{
			"error": "missing credentials",
		})
	} else {
		creds := credentials.NewStaticCredentials(r.FormValue("AWS_ACCESS_KEY_ID"), r.FormValue("AWS_SECRET_ACCESS_KEY"), "")
		c.AwsCredentials = creds
		next(rw, r)
	}
}
