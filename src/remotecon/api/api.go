package api

import (
	"github.com/gocraft/web"
	"net/http"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/awsutil"
)

const awsRegion = "us-west-2"
var vpcFilterName = "vpc-id"

type ApiContext struct {
	*Context
	AwsCredentials		*credentials.Credentials
}

var apiRouter	*web.Router

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
		"status": "ok",
		"access_key_id": creds.AccessKeyID,
	})
}

func (c *ApiContext) VpcListInstances(rw web.ResponseWriter, r *web.Request) {
	if ok := validatePresencePathParams(r.PathParams, "id"); !ok {
		rw.WriteHeader(http.StatusBadRequest)
		writeJson(rw, map[string]string{
			"error": "no vpc id given",
		})
	}
	vpcId := r.PathParams["id"]

	ec2service := ec2.New(&aws.Config{
		Credentials: c.AwsCredentials,
		Region: awsRegion,
	})
	ec2filters := []*ec2.Filter{
		&ec2.Filter{
			Name: &vpcFilterName,
			Values: []*string{&vpcId},
		},
	}
	ec2params := &ec2.DescribeInstancesInput{
		Filters: ec2filters,
	}
	instances, err := ec2service.DescribeInstances(ec2params)
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
