package awscost

import (
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
	"github.com/tmax-cloud/hypercloud-api-server/util"
	"k8s.io/klog"
)

func Get(res http.ResponseWriter, req *http.Request) {

	/*** DETERMIN HOW TO SORT THE RESULT ***/
	// POSSIBLE VALUE : "account", "dimension"
	sort := req.URL.Query().Get(util.QUERY_PARAMETER_SORT)
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
	size := len(lines)/3
	var sess [size]session.Session
	var svc [size]costexplorer.CostExplorer
	idx := 0
	result := make(map[string]awscostModel.Awscost)
	var output *costexplorer.GetCostAndUsageOutput
	
	for _, account := range lines {
		if reg.MatchString(account) {
			account = strings.TrimLeft(account, "[")
			account = strings.TrimRight(account, "]")
			klog.Infoln("Account Name : ", account)

			/*** GET CREDENTIALS BY READING /root/.aws/credentials ***/
			sess[idx], err = session.NewSessionWithOptions(session.Options{
				Profile: account,
				//SharedConfigState: session.SharedConfigEnable,
			})
			if err != nil {
				klog.Errorln(err)
			}
			svc[idx] := costexplorer.New(sess)
			idx++

			output, err = makeCost(req)
			if err != nil {
				klog.Errorln(err)
				return
			}
			result = insert(result, output, req, account, sort)
		}
	}

	/*** MOVE IN TO STRUCT ARRAY FOR SIMPLIFICATION OUTPUT ***/
	result_struct := make([]awscostModel.Awscost, len(result))
	i := 0
	for k, e := range result {
		result_struct[i].Keys = k
		result_struct[i].Metrics = e.Metrics
		i++
	}

	util.SetResponse(res, "", result_struct, http.StatusOK)

	klog.Infoln("=== RESULT ===")
	klog.Infoln(result_struct)
}

func makeCost(req *http.Request) (*costexplorer.GetCostAndUsageOutput, error) {

	queryParams := req.URL.Query()

	/*** GET QUERY PARAMS ***/
	//Must be in YYYY-MM-DD Format
	var startTime int64
	var endTime int64
	startUnix := queryParams.Get(util.QUERY_PARAMETER_STARTTIME)
	endUnix := queryParams.Get(util.QUERY_PARAMETER_ENDTIME)
	if startUnix == "" || endUnix == "" {
		klog.Errorln("Must pass both of startTime and endTime")
		return nil, errors.New("Time parameter error")
	}
	startTime, _ = strconv.ParseInt(startUnix, 10, 64)
	endTime, _ = strconv.ParseInt(endUnix, 10, 64)

	granularity := queryParams.Get(util.QUERY_PARAMETER_GRANULARITY) // "MONTHLY"
	if granularity == "" {
		granularity = "MONTHLY"
	}
	granularity = strings.ToUpper(granularity)

	// "AmortizedCost", "NetAmortizedCost", "BlendedCost", "UnblendedCost", "NetUnblendedCost", "UsageQuantity", "NormalizedUsageAmount",
	metrics := queryParams[util.QUERY_PARAMETER_METRICS]
	if len(metrics) == 0 {
		metrics = []string{
			"BlendedCost",
		}
	}

	// "AZ", "INSTANCE_TYPE", "OPERATING_SYSTEM", "SERVICE", "REGION", ...
	dimension := queryParams.Get(util.QUERY_PARAMETER_DIMENSION)
	if dimension == "" {
		dimension = "INSTANCE_TYPE"
	}
	dimension = strings.ToUpper(dimension)

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
