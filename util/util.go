package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"errors"

	"regexp"

	gomail "gopkg.in/gomail.v2"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

type ClusterMemberInfo struct {
	Id          int64
	Namespace   string
	Cluster     string
	MemberId    string
	Groups      []string
	MemberName  string
	Attribute   string
	Role        string
	Status      string
	CreatedTime time.Time
	UpdatedTime time.Time
}

var (
	SMTPUsernamePath       string
	SMTPPasswordPath       string
	SMTPHost               string
	SMTPPort               int
	AccessSecretPath       string
	accessSecret           string
	username               string
	password               string
	inviteMail             string
	HtmlHomePath           string
	TokenExpiredDate       string
	ParsedTokenExpiredDate time.Duration
	ValidTime              string
)

//Jsonpatch를 담을 수 있는 구조체
type PatchOps struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func ReadFile() {
	content, err := ioutil.ReadFile(AccessSecretPath)
	if err != nil {
		klog.Errorln(err)
		return
	}
	accessSecret = string(content)
	// klog.Infoln(accessSecret)

	content, err = ioutil.ReadFile(SMTPUsernamePath)
	if err != nil {
		klog.Errorln(err)
		return
	}
	username = string(content)

	content, err = ioutil.ReadFile(SMTPPasswordPath)
	if err != nil {
		klog.Errorln(err)
		return
	}
	password = string(content)

	ParsedTokenExpiredDate = parseDate(TokenExpiredDate)
}

func parseDate(tokenExpiredDate string) time.Duration {
	regex := regexp.MustCompile("[0-9]+")
	num := regex.FindAllString(tokenExpiredDate, -1)[0]
	parsedNum, err := strconv.Atoi(num)
	if err != nil {
		panic(err)
	}
	regex = regexp.MustCompile("[a-z]+")
	unit := regex.FindAllString(tokenExpiredDate, -1)[0]

	switch unit {
	case "minutes":
		ValidTime = strconv.Itoa(parsedNum) + "분"
		return time.Minute * time.Duration(parsedNum)
	case "hours":
		ValidTime = strconv.Itoa(parsedNum) + "시"
		return time.Hour * time.Duration(parsedNum)
	case "days":
		ValidTime = strconv.Itoa(parsedNum) + "일"
		return time.Hour * time.Duration(24) * time.Duration(parsedNum)
	case "weeks":
		ValidTime = strconv.Itoa(parsedNum) + "주"
		return time.Hour * time.Duration(24) * time.Duration(7) * time.Duration(parsedNum)
	default:
		return time.Hour * time.Duration(24) * time.Duration(7) //1days
	}
}

// Jsonpatch를 하나 만들어서 slice에 추가하는 함수
func CreatePatch(po *[]PatchOps, o, p string, v interface{}) {
	*po = append(*po, PatchOps{
		Op:    o,
		Path:  p,
		Value: v,
	})
}

// Response.result.message에 err 메시지 넣고 반환
func ToAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func SetResponse(res http.ResponseWriter, outString string, outJson interface{}, status int) http.ResponseWriter {

	//set Cors
	// res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	res.Header().Set("Access-Control-Max-Age", "3628800")
	res.Header().Set("Access-Control-Expose-Headers", "Content-Type, X-Requested-With, Accept, Authorization, Referer, User-Agent")

	//set Out
	if outJson != nil {
		res.Header().Set("Content-Type", "application/json")
		js, err := json.Marshal(outJson)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		//set StatusCode
		res.WriteHeader(status)
		res.Write(js)
		return res

	} else {
		//set StatusCode
		res.WriteHeader(status)
		res.Write([]byte(outString))
		return res

	}
}

func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

func Remove(slice []string, item string) []string {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	// for _, item := range items {
	if _, ok := set[item]; ok {
		delete(set, item)
	}
	// }

	var newSlice []string
	for k, _ := range set {
		newSlice = append(newSlice, k)
	}
	return newSlice
}

// func Remove(slice []string, items []string) []string {
// 	set := make(map[string]struct{}, len(slice))
// 	for _, s := range slice {
// 		set[s] = struct{}{}
// 	}

// 	for _, item := range items {
// 		_, ok := set[item]
// 		if ok {
// 			delete(set, item)
// 		}
// 	}

// 	var newSlice []string
// 	for k, _ := range set {
// 		newSlice = append(newSlice, k)
// 	}
// 	return newSlice
// }

func MonthToInt(month time.Month) int {
	switch month {
	case time.January:
		return 1
	case time.February:
		return 2
	case time.March:
		return 3
	case time.April:
		return 4
	case time.May:
		return 5
	case time.June:
		return 6
	case time.July:
		return 7
	case time.August:
		return 8
	case time.September:
		return 9
	case time.October:
		return 10
	case time.November:
		return 11
	case time.December:
		return 12
	default:
		return 0
	}
}

func SendEmail(from string, to []string, subject string, bodyParameter map[string]string) error {
	// func SendEmail(from string, to []string, subject string, body string, imgPath string, imgCid string) error {
	content, err := ioutil.ReadFile(HtmlHomePath + "cluster-invitation.html")
	if err != nil {
		klog.Errorln(err)
		return err
	}
	inviteMail = string(content)

	inviteMail = strings.Replace(inviteMail, "@@LINK@@", bodyParameter["@@LINK@@"], -1)
	for k, v := range bodyParameter {
		inviteMail = strings.Replace(inviteMail, k, v, -1)
	}

	klog.Infoln(inviteMail)

	m := gomail.NewMessage()
	m.SetHeader("From", username)
	m.SetHeader("To", strings.Join(to[:], ","))
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", inviteMail)
	// m.Embed(imgPath)
	d := gomail.NewDialer(SMTPHost, SMTPPort, username, password)

	if err := d.DialAndSend(m); err != nil {
		klog.Errorln(err)
		return err
	}
	return nil
}

func CreateToken(clusterMember ClusterMemberInfo) (string, error) {
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["namespace"] = clusterMember.Namespace
	atClaims["cluster"] = clusterMember.Cluster
	atClaims["user_id"] = clusterMember.MemberId
	atClaims["user_name"] = clusterMember.MemberName
	atClaims["user_groups"] = clusterMember.Groups
	atClaims["exp"] = time.Now().Add(ParsedTokenExpiredDate).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(accessSecret))
	if err != nil {
		return "", err
	}
	return token, nil
}

func StringParameterException(userGroups []string, args ...string) error {
	if userGroups == nil {
		msg := "UserGroups is empty."
		klog.Infoln(msg)
		return errors.New(msg)
	}

	for _, arg := range args {
		if arg == "" {
			msg := arg + "Something is empty."
			klog.Infoln(msg)
			return errors.New(msg)
		}
	}
	return nil
}

func ExtractTokenFromHeader(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func ExtractTokenFromQuery(r *http.Request) string {
	bearToken := r.URL.Query().Get("token")
	if bearToken == "" {
		return ""
	}
	return bearToken
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractTokenFromQuery(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}
		return []byte(accessSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// func TokenValid(r *http.Request) error {
// 	token, err := VerifyToken(r)
// 	if err != nil {
// 		return err
// 	}
// 	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
// 		return err
// 	}
// 	return nil
// }

func TokenValid(r *http.Request, clusterMember ClusterMemberInfo) ([]string, error) {
	var memberId string
	var cluster string
	var namespace string
	var groups []string
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		memberId, ok = claims["user_id"].(string)
		cluster, ok = claims["cluster"].(string)
		namespace, ok = claims["namespace"].(string)
		tmp := claims["user_groups"].([]interface{})
		groups = make([]string, len(tmp))
		for i, v := range tmp {
			groups[i] = fmt.Sprint(v)
		}
	}

	if clusterMember.MemberId == memberId && clusterMember.Cluster == cluster && clusterMember.Namespace == namespace {
		return groups, nil
	}
	return nil, errors.New("Request user or target cluster does not match with token payload")
}
