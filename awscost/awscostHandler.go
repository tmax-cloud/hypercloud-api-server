package awscost

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	awscostModel "github.com/tmax-cloud/hypercloud-api-server/awscost/model"
	"k8s.io/klog"
)

func Get(res http.ResponseWriter, req *http.Request) {

	/*** DETERMIN HOW TO SORT THE RESULT ***/
	// POSSIBLE VALUE : "account", "dimension"
	sort := req.URL.Query().Get("sort")
	if sort == "" {
		sort = "account"
	}

	/*** READ AND PARSING FROM CREDENTIAL FILE ***/
	fileName := "/root/.aws/credentials"
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		klog.Errorln(err)
	}
	lines := strings.Split(string(dat), "\n")
	reg := regexp.MustCompile("\\[.*\\]")

	/*** GET COST INFO FOR EACH ACCOUNT ***/
	result := make(map[string]awscostModel.Awscost)
	var output *costexplorer.GetCostAndUsageOutput
	for _, account := range lines {
		if reg.MatchString(account) {
			account = strings.TrimLeft(account, "[")
			account = strings.TrimRight(account, "]")
			klog.Infoln("Account Name : ", account)

			output, err = makeCost(req, account)
			if err != nil {
				klog.Errorln(err)
				return
			}
			result = insert(result, output, req, account, sort)
		}
	}

	res.Header().Set("Content-Type", "application/json")
	js, err := json.Marshal(result)
	if err != nil {
		klog.Errorln(err)
	}
	res.Write(js)

	klog.Infoln(result)
}

func makeCost(req *http.Request, account string) (*costexplorer.GetCostAndUsageOutput, error) {

	queryParams := req.URL.Query()

	/*** GET QUERY PARAMS ***/
	//Must be in YYYY-MM-DD Format
	var startTime int64
	var endTime int64
	startUnix := queryParams.Get("startTime")
	endUnix := queryParams.Get("endTime")
	if startUnix == "" || endUnix == "" {
		klog.Errorln("Must pass both of startTime and endTime")
		return nil, errors.New("Time parameter error")
	}
	startTime, _ = strconv.ParseInt(startUnix, 10, 64)
	endTime, _ = strconv.ParseInt(endUnix, 10, 64)

	granularity := queryParams.Get("granularity") // "MONTHLY"
	if granularity == "" {
		granularity = "MONTHLY"
	}

	// "AmortizedCost", "NetAmortizedCost", "BlendedCost", "UnblendedCost", "NetUnblendedCost", "UsageQuantity", "NormalizedUsageAmount",
	metrics := queryParams["metrics"]
	if len(metrics) == 0 {
		metrics = []string{
			"BlendedCost",
		}
	}

	// "AZ", "INSTANCE_TYPE", "OPERATING_SYSTEM", "SERVICE", "REGION", ...
	dimension := queryParams.Get("dimension")
	if dimension == "" {
		dimension = "INSTANCE_TYPE"
	}

	/*** GET CREDENTIALS BY READING /root/.aws/credentials ***/
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: account,
		//SharedConfigState: session.SharedConfigEnable,
	})
	svc := costexplorer.New(sess)

	/*** GET COST FROM AWS ***/
	result, err := svc.GetCostAndUsage(&costexplorer.GetCostAndUsageInput{
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(time.Unix(startTime, 0).Format("2006-01-02")),
			End:   aws.String(time.Unix(endTime, 0).Format("2006-01-02")),
		},
		Granularity: aws.String(granularity),
		GroupBy: []*costexplorer.GroupDefinition{
			&costexplorer.GroupDefinition{
				Type: aws.String("DIMENSION"),
				Key:  aws.String(dimension),
			},
		},
		Metrics: aws.StringSlice(metrics),
	})
	if err != nil {
		klog.Errorln(err)
	}

	return result, err
}

func insert(result map[string]awscostModel.Awscost, output *costexplorer.GetCostAndUsageOutput, req *http.Request, account string, sort string) map[string]awscostModel.Awscost {

	metrics := req.URL.Query()["metrics"]
	if len(metrics) == 0 {
		metrics = []string{
			"BlendedCost",
		}
	}

	/*** append to result based on sorting criteria ***/
	for i := range output.ResultsByTime {
		for _, g := range output.ResultsByTime[i].Groups {
			temp := awscostModel.NewAwscost()
			for _, metric := range metrics {
				tAmount, _ := strconv.ParseFloat(*g.Metrics[metric].Amount, 64)
				tUnit := *g.Metrics[metric].Unit
				temp.Metrics[metric] = &awscostModel.Metric{Amount: tAmount, Unit: tUnit}
				result = add(result, *temp, account, metric, *g.Keys[0], sort)
			}
		}
	}

	// t1, _ := strconv.ParseFloat(*output.ResultsByTime[0].Groups[0].Metrics["BlendedCost"].Amount, 64)
	// t2 := *output.ResultsByTime[0].Groups[0].Metrics["BlendedCost"].Unit
	// temp := awscostModel.NewAwscost()
	// temp.Metrics["BlendedCost"] = &awscostModel.Metric{Amount: t1, Unit: t2}

	//result = append(result, *temp)

	return result
}

func add(result map[string]awscostModel.Awscost, sub awscostModel.Awscost, account string, metric string, key string, sort string) map[string]awscostModel.Awscost {

	// If there is already value in map, combine them
	// If not, just insert
	if sort == "account" {
		if _, exist := result[account]; exist {
			if _, existMetric := result[account].Metrics[metric]; existMetric {
				result[account].Metrics[metric].Amount += sub.Metrics[metric].Amount
			} else {
				result[account].Metrics[metric] = sub.Metrics[metric]
			}
		} else {
			result[account] = sub
		}
	} else if sort == "dimension" {
		if _, exist := result[key]; exist {
			if _, existAccount := result[key].Metrics[account]; existAccount {
				result[key].Metrics[account].Amount += sub.Metrics[metric].Amount
			} else {
				result[key].Metrics[account] = sub.Metrics[metric]
			}
		} else {
			result[key] = *awscostModel.NewAwscost()
			result[key].Metrics[account] = sub.Metrics[metric]
		}
	}

	return result
}
