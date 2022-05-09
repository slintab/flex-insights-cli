package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func NewRequest(client *http.Client, url string, method string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func NewRequestWithPayload(client *http.Client, url string, method string, headers map[string]string, payload interface{}) (*http.Response, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func NewRequestWithRetry(client *http.Client, url string, method string, headers map[string]string, payload interface{}) (*http.Response, error) {
	var resp *http.Response
	var err error

	retryCount, allowedRetries := 0, 8

	for {
		if payload != nil {
			resp, err = NewRequestWithPayload(client, url, method, headers, payload)
		} else {
			resp, err = NewRequest(client, url, method, headers)
		}

		if err != nil {
			break
		}

		if resp.StatusCode != http.StatusAccepted {
			break
		}

		if allowedRetries < retryCount {
			break
		}

		fmt.Println("Downloading report - this may take some time... ")
		retryCount += 1
		time.Sleep(time.Duration(retryCount) * 10 * time.Second)

	}

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusAccepted {
		return nil, fmt.Errorf("Report is taking too long to generate. You can download when it is ready here: %s", url)
	}

	return resp, nil

}

func ParseResponse(resp *http.Response, v interface{}) {
	json.NewDecoder(resp.Body).Decode(v)
}

type PostUserLogin struct {
	Login        string `json:"login"`
	Password     string `json:"password"`
	Remember     int    `json:"remember"`
	Verify_level int    `json:"verify_level"`
}

type SstPayload struct {
	PostUserLogin PostUserLogin `json:"postUserLogin"`
}

type UserLogin struct {
	Profile string `json:"profile"`
	State   string `json:"state"`
	Token   string `json:"token"`
}

type SstResponse struct {
	UserLogin UserLogin `json:"userLogin"`
}

type UserToken struct {
	Token string `json:"token"`
}

type TtResponse struct {
	UserToken UserToken `json:"userToken"`
}

type ReportReq struct {
	Report string `json:"report"`
}

type ReportPayload struct {
	ReportReq ReportReq `json:"report_req"`
}

type ReportResponse struct {
	Uri string `json:"uri"`
}

var (
	user      string
	password  string
	object    string
	workspace string
	filename  string

	resp *http.Response
	err  error

	exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export reports.",
		Long:  `You can use the 'export' command to retrieve reports from Flex Insights. Rather than having to make three separate API calls, with this utility you can export reports by using a single command. Example use: flex-insights-cli export --user me@email.com --password 123456 --workspace abcd --objectid 9999 --output myreport.csv`,
		Run: func(cmd *cobra.Command, args []string) {
			//Retrieval of reports based on process outline in https://www.twilio.com/docs/flex/developer/insights/api/export-data

			client := &http.Client{}

			if len(user) < 1 || len(password) < 1 {
				fmt.Println("Credentials not found. See --help for instructions.")
				return
			}

			//Retrieving the super secure token
			sstUrl := "https://analytics.ytica.com/gdc/account/login"
			sstpayload := &SstPayload{
				PostUserLogin: PostUserLogin{
					Login:        user,
					Password:     password,
					Remember:     0,
					Verify_level: 2,
				},
			}
			sstHeaders := map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			}

			fmt.Println("Retrieving Super Secure Token...")
			fmt.Println()

			resp, err = NewRequestWithPayload(client, sstUrl, "POST", sstHeaders, sstpayload)
			if err != nil {
				fmt.Println("Error retrieving SuperSecure Token. Check your credentials.")
				return
			}

			sst := SstResponse{}
			ParseResponse(resp, &sst)

			//Retrieving a temporary token
			ttUrl := "https://analytics.ytica.com/gdc/account/token"
			ttHeaders := map[string]string{
				"Accept":        "application/json",
				"Content-Type":  "application/json",
				"X-GDC-AuthSST": sst.UserLogin.Token,
			}

			fmt.Println("Retrieving Temporary Token...")
			fmt.Println()

			resp, err = NewRequest(client, ttUrl, "GET", ttHeaders)
			if err != nil {
				fmt.Println("Error retrieving Temporary Token.")
				return
			}

			tt := TtResponse{}
			ParseResponse(resp, &tt)

			//Exporting the report
			reportUrl := fmt.Sprintf("%s%s%s", "https://analytics.ytica.com/gdc/app/projects/", workspace, "/execute/raw")
			reportPayload := &ReportPayload{
				ReportReq: ReportReq{
					Report: fmt.Sprintf("%s%s%s%s", "/gdc/md/", workspace, "/obj/", object),
				},
			}
			reportHeaders := map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
				"Cookie":       "GDCAuthTT=" + tt.UserToken.Token,
			}

			fmt.Println("Exporting report...")
			fmt.Println()

			resp, err = NewRequestWithPayload(client, reportUrl, "POST", reportHeaders, reportPayload)
			if err != nil {
				fmt.Println("Error exporting report.")
				return
			}

			reportUri := ReportResponse{}
			ParseResponse(resp, &reportUri)

			//Downloading the report
			downloadUrl := "https://analytics.ytica.com" + reportUri.Uri

			downloadHeaders := map[string]string{
				"Cookie": "GDCAuthTT=" + tt.UserToken.Token,
			}

			fmt.Println("Downloading report...")
			fmt.Println()

			resp, err = NewRequestWithRetry(client, downloadUrl, "GET", downloadHeaders, nil)
			if err != nil {
				fmt.Println("Error downloading report.")
				return
			}

			//Logging out
			profile := sst.UserLogin.Profile
			profileId := profile[strings.LastIndex(profile, "/")+1:]

			logoutUrl := "https://analytics.ytica.com/gdc/account/login/" + profileId
			logoutHeaders := map[string]string{
				"Accept":        "application/json",
				"Content-Type":  "application/json",
				"Cookie":        "GDCAuthTT=" + tt.UserToken.Token,
				"X-GDC-AuthSST": sst.UserLogin.Token,
			}

			fmt.Println("Logging out...")
			fmt.Println()

			_, err = NewRequest(client, logoutUrl, "DELETE", logoutHeaders)
			if err != nil {
				fmt.Println("Error logging out.")
				fmt.Println()
			}

			//Saving report
			fmt.Println("Saving report...")
			fmt.Println()

			file, err := os.Create(filename)
			defer file.Close()
			if err != nil {
				fmt.Println("Error opening file: " + filename)
				return
			}

			_, err = io.Copy(file, resp.Body)
			if err != nil {
				fmt.Println("Error saving report.")
				return
			}

			fmt.Println("Report successfully saved to: " + filename)

		},
	}
)

func init() {
	exportCmd.Flags().StringVarP(&user, "user", "u", os.Getenv("FLEX_INSIGHTS_USER"), "login username (required) - can also be set using the FLEX_INSIGHTS_USER env var")

	exportCmd.Flags().StringVarP(&password, "password", "p", os.Getenv("FLEX_INSIGHTS_PASSWORD"), "login password (required) - can also be set using the FLEX_INSIGHTS_PASSWORD env var")

	exportCmd.Flags().StringVarP(&object, "objectid", "o", "", "report object ID (required)")
	exportCmd.MarkFlagRequired("objectid")

	exportCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "workspace ID (required)")
	exportCmd.MarkFlagRequired("workspace")

	exportCmd.Flags().StringVarP(&filename, "output", "f", "", "output file (required)")
	exportCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(exportCmd)

}
