package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

type CustomError struct {
	ErrorCode int
	Err       error
}

func (err *CustomError) Error() string {
	return fmt.Sprintf("Error: %v, StatusCode: %d", err.Err, err.ErrorCode)
}

type DNSRecord struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
	TTL    int    `json:"ttl"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type Configuration struct {
	Config []DNSRecord
}

type GodaddyRecordBody struct {
	Data string `json:"data"`
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
}

type GodaddyErrorField struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

type GodaddyErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Fields  []GodaddyErrorField
}

var (
	version             string        = "1.0.0-go1.17"
	config_file         string        = "config.json"
	home, _                           = os.UserHomeDir()
	config_loc          string        = home + "/.config"
	config_dir_perm     fs.FileMode   = 0700
	config_file_perm    fs.FileMode   = 0600
	godaddy_api_version string        = "v1"
	max_record_size     int           = 5
	daemon_poll_time    time.Duration = 1 * time.Minute // Time in minute
	log_file            string        = config_loc + "/godaddy-ddns/log/godaddy-ddns.log"
)

const (
	ErrorLog       string = "ERROR"
	InformationLog string = "INFO"
	WarningLog     string = "WARN"
)

func GoDaddyDDNSLogger(logType, name, domain, message string) {
	file, err := os.OpenFile(log_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var (
		WarningLogger       *log.Logger
		InfoLogger          *log.Logger
		ErrorLogger         *log.Logger
		StdoutInfoLogger    *log.Logger
		StdoutWarningLogger *log.Logger
		StdoutErrorLogger   *log.Logger
	)

	InfoLogger = log.New(file, "INFO ", log.Ldate|log.Ltime)
	WarningLogger = log.New(file, "WARN ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(file, "ERROR ", log.Ldate|log.Ltime)
	StdoutInfoLogger = log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)
	StdoutWarningLogger = log.New(os.Stdout, "WARN ", log.Ldate|log.Ltime)
	StdoutErrorLogger = log.New(os.Stdout, "ERROR ", log.Ldate|log.Ltime)

	if logType == "INFO" {
		InfoLogger.Println(name+"."+domain, message)
		StdoutInfoLogger.Println(name+"."+domain, message)
	} else if logType == "WARN" {
		WarningLogger.Println(name+"."+domain, message)
		StdoutWarningLogger.Println(name+"."+domain, message)
	} else if logType == "ERROR" {
		ErrorLogger.Println(name+"."+domain, message)
		StdoutErrorLogger.Println(name+"."+domain, message)
	} else {
		fmt.Println(name+"."+domain, message)
	}
}

func init() {
	if _, err := os.Stat(config_loc); os.IsNotExist(err) {
		err := os.Mkdir(config_loc, config_dir_perm)
		if err != nil {
			fmt.Println("Failed to create directory,", config_loc, err.Error())
			return
		}

		err = os.Mkdir(config_loc+"/godaddy-ddns", config_dir_perm)
		if err != nil {
			fmt.Println("Failed to create directory,", config_loc+"/godaddy-ddns", err.Error())
			return
		}

		file, err := os.Create(config_loc + "/godaddy-ddns/" + config_file)
		if err != nil {
			fmt.Println("Failed to create config file,", config_loc+"/godaddy-ddns/"+config_file, err.Error())
			return
		}
		file.Close()
	}

	if _, err := os.Stat(config_loc + "/godaddy-ddns/"); os.IsNotExist(err) {
		err = os.Mkdir(config_loc+"/godaddy-ddns", config_dir_perm)
		if err != nil {
			fmt.Println("Failed to create directory,", config_loc+"/godaddy-ddns", err.Error())
			return
		}

		file, err := os.Create(config_loc + "/godaddy-ddns/" + config_file)
		if err != nil {
			fmt.Println("Failed to create config file,", config_loc+"/godaddy-ddns/"+config_file, err.Error())
			return
		}
		file.Close()
	}

	if _, err := os.Stat(config_loc + "/godaddy-ddns/" + config_file); os.IsNotExist(err) {
		file, err := os.Create(config_loc + "/godaddy-ddns/" + config_file)
		if err != nil {
			fmt.Println("Failed to create config file,", config_loc+"/godaddy-ddns/"+config_file, err.Error())
			return
		}
		file.Close()
	}

	if _, err := os.Stat(config_loc + "/godaddy-ddns/log"); os.IsNotExist(err) {
		err := os.Mkdir(config_loc+"/godaddy-ddns/log", config_dir_perm)
		if err != nil {
			if err != nil {
				fmt.Println("Failed to create directory,", config_loc+"/godaddy-ddns/log", err.Error())
				return
			}
			return
		}
	}

	// executablePath()
}

// func executablePath() {
// 	ex, err := os.Executable()
// 	if err != nil {
// 		GoDaddyDDNSLogger(WarningLog, "", "", err.Error())
// 	}
// 	exPath := filepath.Dir(ex)
// 	fmt.Println(exPath)
// }

func main() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)

	domain := addCmd.String("domain", "", "Domain name e.g. example.com")
	name := addCmd.String("name", "", "Subdomain or hostname e.g. www")
	ttl := addCmd.Int("ttl", 600, "Time-to-live in seconds. Minimum 600 seconds.")
	key := addCmd.String("key", "", "Key value generated from godaddy developer console")
	secret := addCmd.String("secret", "", "Secret value generated from godaddy developer console")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteDomain := deleteCmd.String("domain", "", "Domain name e.g. example.com")
	deleteName := deleteCmd.String("name", "", "Subdomain or hostname e.g. www")

	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	updateDomain := updateCmd.String("domain", "", "Domain name e.g. example.com")
	updateName := updateCmd.String("name", "", "Subdomain or hostname e.g. www")
	updateTtl := updateCmd.Int("ttl", 600, "Time-to-live in seconds. Minimum 600 seconds.")
	updateKey := updateCmd.String("key", "", "Key value generated from godaddy developer console")
	updateSecret := updateCmd.String("secret", "", "Secret value generated from godaddy developer console")

	switch os.Args[1] {

	case "version":
		fmt.Println("GoDaddy DDNS version", version)
		return

	case "add":
		addCmd.Parse(os.Args[2:])
		if *domain == "" || *name == "" || *key == "" || *secret == "" {
			fmt.Println("ERROR domain, name, key and secret are mandatory")
			fmt.Printf("\nUsage of %s:\n", os.Args[1])
			addCmd.PrintDefaults()
			return
		}
		if *ttl < 600 {
			fmt.Println("ERROR TTL value cannot be less than 600 seconds.")
			fmt.Printf("\nUsage of %s:\n", os.Args[1])
			addCmd.PrintDefaults()
			return
		}

		err := addRecord(*domain, *name, *key, *secret, *ttl, false)
		if err != nil {
			GoDaddyDDNSLogger(ErrorLog, *name, *domain, "Failed to add record. "+err.Error())
			return
		}

	case "delete":
		deleteCmd.Parse(os.Args[2:])
		if *deleteDomain == "" || *deleteName == "" {
			fmt.Println("ERROR domain and name are mandatory")
			fmt.Printf("\nUsage of %s:\n", os.Args[1])
			deleteCmd.PrintDefaults()
			return
		}
		err := deleteRecord(*deleteDomain, *deleteName)
		if err != nil {
			GoDaddyDDNSLogger(ErrorLog, *deleteName, *deleteDomain, "Failed to delete record. "+err.Error())
			return
		} else {
			GoDaddyDDNSLogger(InformationLog, *deleteName, *deleteDomain, "Record removed from configuration. If not in use, delete the record manually from GoDaddy console.")
		}

	case "update":
		updateCmd.Parse(os.Args[2:])
		if *updateDomain == "" || *updateName == "" || *updateKey == "" || *updateSecret == "" {
			fmt.Println("ERROR domain, name, key and secret are mandatory")
			fmt.Printf("\nUsage of %s:\n", os.Args[1])
			updateCmd.PrintDefaults()
			return
		}
		if *ttl < 600 {
			fmt.Println("ERROR TTL value cannot be less than 600 seconds.")
			fmt.Printf("\nUsage of %s:\n", os.Args[1])
			updateCmd.PrintDefaults()
			return
		}
		err := addRecord(*updateDomain, *updateName, *updateKey, *updateSecret, *updateTtl, true)
		if err != nil {
			GoDaddyDDNSLogger(ErrorLog, *updateName, *updateDomain, "Failed to update record. "+err.Error())
			return
		}

	case "daemon":
		// ticker := time.NewTicker(daemon_poll_time * time.Minute)
		// quit := make(chan struct{})
		// go daemonDDNS(ticker, &quit)
		daemonDDNS()

	case "list":
		err := listRecord()
		if err != nil {
			fmt.Println("Failed to list records,", err.Error())
			return
		}

	default:
		fmt.Printf("\nUsage:\n")

		fmt.Printf("\nadd\n")
		fmt.Printf("\tAdd new record. Max 5 records\n")
		addCmd.PrintDefaults()
		fmt.Printf("\nupdate\n")
		fmt.Printf("\tUpdate existing record\n")
		updateCmd.PrintDefaults()
		fmt.Printf("\ndelete\n")
		fmt.Printf("\tDelete existing record\n")
		deleteCmd.PrintDefaults()
		fmt.Printf("\nlist\n")
		fmt.Printf("\tList all configured records\n")
		fmt.Printf("\nversion\n")
		fmt.Printf("\tCheck version\n")
		fmt.Printf("\n\nExamples\n")
		fmt.Printf("\tgoddns list\n")
		fmt.Printf("\tgoddns add --domain='example.com' --name='myweb' --ttl=1200 --key='kEyGeneratedFr0mG0DaddY' --secret='s3cRe7GeneratedFr0mG0DaddY'\n")
		fmt.Printf("\tgoddns update --domain='example.com' --name='myweb' --key='kEyGeneratedFr0mG0DaddY' --secret='s3cRe7GeneratedFr0mG0DaddY'\n")
		fmt.Printf("\tgoddns delete --domain='example.com' --name='myweb'\n")
		fmt.Printf("\tgoddns version'\n")

	}

}

func addRecord(domain, name, key, secret string, ttl int, isUpdate bool) error {
	record := DNSRecord{
		Domain: domain,
		Name:   name,
		Key:    key,
		Secret: secret,
		TTL:    ttl,
	}

	var config Configuration
	var updatedConfig Configuration
	var hasUpdated bool = false

	body, err := getDNSRecord(name, domain, key, secret)
	if err != nil {
		return err
	}

	var recordsBody []GodaddyRecordBody
	err = json.Unmarshal([]byte(body), &recordsBody)
	if err != nil {
		return err
	}

	var existingTtl int
	var existingIp string

	if len(recordsBody) != 0 {
		existingTtl = recordsBody[0].TTL
		existingIp = recordsBody[0].Data
	} else {
		existingTtl = 0
		existingIp = ""
	}

	pubIp, err := getPubIP()
	if err != nil {
		return err
	}

	configFileContent, err := ioutil.ReadFile(config_loc + "/godaddy-ddns/" + config_file)
	if err != nil {
		return err
	}

	if len(configFileContent) != 0 {
		err = json.Unmarshal(configFileContent, &config)
		if err != nil {
			return err
		}
		if len(config.Config) >= max_record_size && !isUpdate {
			return &CustomError{ErrorCode: 1, Err: errors.New("reached record limit. maximum " + fmt.Sprintf("%v", max_record_size) + " records allowed per server")}
		}
		for _, i := range config.Config {
			if i.Domain == domain && i.Name == name {
				if !isUpdate {
					return &CustomError{ErrorCode: 1, Err: errors.New("record already exist")}
				} else {
					hasUpdated = true
					continue
				}
			}
			updatedConfig.Config = append(updatedConfig.Config, i)
		}
		if isUpdate && !hasUpdated {
			return &CustomError{ErrorCode: 1, Err: errors.New("record not found")}
		}
		updatedConfig.Config = append(updatedConfig.Config, record)
	} else {
		updatedConfig.Config = append(updatedConfig.Config, record)
	}

	configFileContent, err = json.MarshalIndent(updatedConfig, "", "  ")
	if err != nil {
		return err
	}

	if ttl != existingTtl || pubIp != existingIp {
		_, err := setDNSRecord(name, domain, key, secret, pubIp, ttl)
		if err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(config_loc+"/godaddy-ddns/"+config_file, configFileContent, config_file_perm)

	if err != nil {
		return err
	}

	GoDaddyDDNSLogger(InformationLog, name, domain, "Record created/updated (ttl: "+fmt.Sprintf("%d", ttl)+", ip: "+pubIp+", key: ****, secret: ****)")

	return nil
}

func getDNSRecord(name, domain, key, secret string) (string, error) {
	// Get record details from GoDaddy

	gdURL := "https://api.godaddy.com/" + godaddy_api_version + "/domains/" + domain + "/records/A/" + name
	authorization := key + ":" + secret

	apiclient := &http.Client{}

	req, err := http.NewRequest("GET", gdURL, nil)
	if err != nil {
		// fmt.Println(err.Error())
		return "", err
	}

	req.Header.Add("Authorization", "sso-key "+authorization)
	response, err := apiclient.Do(req)
	if err != nil {
		// fmt.Println(err.Error())
		return "", err
	}

	defer response.Body.Close()

	// var bodyBytes []byte

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// fmt.Println(err.Error())
		return "", err
	}

	if response.StatusCode != 200 {
		var errorBody GodaddyErrorBody
		err := json.Unmarshal(bodyBytes, &errorBody)
		if err != nil {
			// fmt.Println(err.Error())
			return "", err
		}
		return "", &CustomError{ErrorCode: response.StatusCode, Err: errors.New(errorBody.Message)}
	}

	return string(bodyBytes), nil
}

// func addGodaddyRecord() {

// }
func getPubIP() (string, error) {

	type GetIPBody struct {
		IP string `json:"ip"`
	}

	var ipbody GetIPBody

	response, err := http.Get("https://ipinfo.io/json")
	if err != nil {
		return "", nil
	}

	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// fmt.Println(err.Error())
		return "", err
	}

	err = json.Unmarshal(bodyBytes, &ipbody)
	if err != nil {
		// fmt.Println(err.Error())
		return "", err
	}

	return ipbody.IP, nil

}

func setDNSRecord(name, domain, key, secret, pubIp string, ttl int) (string, error) {

	gdURL := "https://api.godaddy.com/" + godaddy_api_version + "/domains/" + domain + "/records/A/" + name
	authorization := key + ":" + secret

	type Data struct {
		Data string `json:"data"`
		TTL  int    `json:"ttl"`
	}

	data, _ := json.Marshal([]Data{
		{
			Data: pubIp,
			TTL:  ttl,
		},
	})

	requestBody := bytes.NewBuffer(data)
	// requestBody := bytes.NewBuffer()

	apiclient := &http.Client{}

	req, err := http.NewRequest("PUT", gdURL, requestBody)
	if err != nil {
		// fmt.Println(1, err.Error())
		return "", err
	}

	req.Header.Add("Authorization", "sso-key "+authorization)
	req.Header.Add("Content-Type", "application/json")

	response, err := apiclient.Do(req)
	if err != nil {
		// fmt.Println(2, err.Error())
		return "", err
	}

	defer response.Body.Close()

	var bodyBytes []byte

	if response.StatusCode != 200 {
		var errorBody GodaddyErrorBody
		bodyBytes, err = ioutil.ReadAll(response.Body)
		if err != nil {
			// fmt.Println(err.Error())
			return "", err
		}
		err := json.Unmarshal(bodyBytes, &errorBody)
		if err != nil {
			// fmt.Println(3, err.Error())
			return "", err
		}
		// return "", &CustomError{ErrorCode: response.StatusCode, Err: errors.New(errorBody.Message)}

		return "", &CustomError{ErrorCode: response.StatusCode, Err: errors.New(errorBody.Message)}
	}

	return string(bodyBytes), nil
}

func deleteRecord(domain, name string) error {

	var config Configuration
	var newConfig Configuration
	var done bool = false

	configFileContent, err := ioutil.ReadFile(config_loc + "/godaddy-ddns/" + config_file)
	if err != nil {
		return err
	}

	if len(configFileContent) != 0 {
		err = json.Unmarshal(configFileContent, &config)
		if err != nil {
			return err
		}

		for _, i := range config.Config {
			if i.Domain == domain && i.Name == name {
				done = true
				continue
			}
			newConfig.Config = append(newConfig.Config, i)
		}
	}

	if len(configFileContent) == 0 || !done {
		return &CustomError{ErrorCode: 1, Err: errors.New("record doesnot exist")}
	}

	configFileContent, err = json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(config_loc+"/godaddy-ddns/"+config_file, configFileContent, config_file_perm)

	if err != nil {
		return err
	}

	return nil
}

func listRecord() error {
	var config Configuration

	configFileContent, err := ioutil.ReadFile(config_loc + "/godaddy-ddns/" + config_file)
	if err != nil {
		return err
	}

	if len(configFileContent) != 0 {
		err = json.Unmarshal(configFileContent, &config)
		if err != nil {
			return err
		}

		if len(config.Config) == 0 {
			return &CustomError{ErrorCode: 1, Err: errors.New("no record exist")}
		}

		t := table.NewWriter()

		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"#", "Name", "Domain", "TTL"})

		for i, rec := range config.Config {
			t.AppendRow(table.Row{i + 1, rec.Name, rec.Domain, rec.TTL})
		}
		t.Render()

	} else {
		return &CustomError{ErrorCode: 1, Err: errors.New("no record exist")}
	}

	return nil
}

func daemonDDNS() {

	GoDaddyDDNSLogger(InformationLog, "", "", "Starting daemon process")

	ticker := time.NewTicker(daemon_poll_time)
	done := make(chan bool)

	if _, err := os.Stat(config_loc + "/godaddy-ddns/" + "daemon.lock"); !os.IsNotExist(err) {
		err = os.Remove(config_loc + "/godaddy-ddns/" + "daemon.lock")
		if err != nil {
			GoDaddyDDNSLogger(ErrorLog, "", "", "Failed to release lock")
			os.Exit(1)
		}
	}

	go func() {
		for {
			select {
			case <-done:
				return

			case <-ticker.C:

				GoDaddyDDNSLogger(InformationLog, "", "", "Polling the records")

				var config Configuration

				if _, err := os.Stat(config_loc + "/godaddy-ddns/" + "daemon.lock"); !os.IsNotExist(err) {
					GoDaddyDDNSLogger(WarningLog, "", "", "A daemon is already running. Waiting to release lock")
					continue
				}

				file, err := os.Create(config_loc + "/godaddy-ddns/" + "daemon.lock")
				if err != nil {
					GoDaddyDDNSLogger(ErrorLog, "", "", "Failed to apply lock")
					continue
				}
				file.Close()

				configFileContent, err := ioutil.ReadFile(config_loc + "/godaddy-ddns/" + config_file)
				if err != nil {
					GoDaddyDDNSLogger(ErrorLog, "", "", "Failed to read configuration file. "+err.Error())
					continue
				}

				if len(configFileContent) != 0 {
					err = json.Unmarshal(configFileContent, &config)
					if err != nil {
						GoDaddyDDNSLogger(ErrorLog, "", "", "Failed to read configuration file. "+err.Error())
						continue
					}

					if len(config.Config) == 0 {
						GoDaddyDDNSLogger(WarningLog, "", "", "No record added in configuration")
						continue
					}

					for _, i := range config.Config {

						name := i.Name
						domain := i.Domain
						key := i.Key
						secret := i.Secret
						ttl := i.TTL

						body, err := getDNSRecord(name, domain, key, secret)
						if err != nil {
							GoDaddyDDNSLogger(ErrorLog, name, domain, "Failed to get current state of record. "+err.Error())
							continue
						}

						var recordsBody []GodaddyRecordBody
						err = json.Unmarshal([]byte(body), &recordsBody)
						if err != nil {
							GoDaddyDDNSLogger(ErrorLog, name, domain, "Failed to read current state of record. "+err.Error())
							continue
						}

						var existingTtl int
						var existingIp string

						if len(recordsBody) != 0 {
							existingTtl = recordsBody[0].TTL
							existingIp = recordsBody[0].Data
						} else {
							existingTtl = 0
							existingIp = ""
						}

						pubIp, err := getPubIP()
						if err != nil {
							GoDaddyDDNSLogger(ErrorLog, name, domain, "Failed to get current Pub IP of server. "+err.Error())
							continue
						}

						if ttl != existingTtl || pubIp != existingIp {
							_, err := setDNSRecord(name, domain, key, secret, pubIp, ttl)
							if err != nil {
								GoDaddyDDNSLogger(ErrorLog, name, domain, "Failed to update record. "+err.Error())
								continue
							} else {
								GoDaddyDDNSLogger(InformationLog, name, domain, "Record updated (ttl: "+fmt.Sprintf("%d", existingTtl)+"->"+fmt.Sprintf("%d", ttl)+", ip: "+existingIp+"->"+pubIp+")")
							}
						} else {
							GoDaddyDDNSLogger(InformationLog, name, domain, "Desired state is current state")
						}

						time.Sleep(10 * time.Second) // Wait for 10 seconds before picking next record.
					}
				}

				err = os.Remove(config_loc + "/godaddy-ddns/" + "daemon.lock")
				if err != nil {
					GoDaddyDDNSLogger(ErrorLog, "", "", "Failed to release lock")
					continue
				}
			}
		}
	}()

	// Handle signal interrupt

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			GoDaddyDDNSLogger(InformationLog, "", "", "Interrupt signal received. Exiting")
			ticker.Stop()
			done <- true
			if _, err := os.Stat(config_loc + "/godaddy-ddns/" + "daemon.lock"); !os.IsNotExist(err) {
				_ = os.Remove(config_loc + "/godaddy-ddns/" + "daemon.lock")
			}
			os.Exit(0)
		}
	}()

	time.Sleep(8760 * time.Hour) // Sleep for 365 days
	ticker.Stop()
	done <- true
}
