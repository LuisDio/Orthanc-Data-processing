/* Make sure you have the Download directory available in user directory
*/

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type BasicAuthTransport struct {
	Username string
	Password string
}

// Transsport not really implemented here
func (bat BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
			bat.Username, bat.Password)))))
	return http.DefaultTransport.RoundTrip(req)
}

func (bat *BasicAuthTransport) Client() *http.Client {
	return &http.Client{Transport: bat}
}

func accessor() *http.Client {
	identification := BasicAuthTransport{"username", "password"}
	connector := identification.Client()
	return connector

}

func contains(metadataList []string, metadataName string) (bool, string) {
	for _, element := range metadataList {
		if metadataName == element {
			return true, element
		}
	}
	return false, ""
}

func execDownload(hash string) {
	fmt.Println("Swarm Download execution")
	swarmCmd := exec.Command("bash", "-c", "./execDown.sh"+" "+hash)

	_, err := swarmCmd.Output()
	if err != nil {
		//panic(err)
		log.Fatal(err)
	}
	//fmt.Println(out1)
}

func execExtract() {
	fmt.Println("Os execute unzip file")
	unzipCmd := exec.Command("bash", "-c", "unzip -j /home/$USER/Download/data.zip -d /home/$USER/Download")

	_, err := unzipCmd.Output()
	if err != nil {
		panic(err)
	}
}

func execTransfer() {
	fmt.Println("Os execute Send file")
	transfCmd := exec.Command("bash", "-c", "storescu -aec ORTHANC localhost 4242 /home/$USER/Download/*.dcm")

	_, err := transfCmd.Output()
	if err != nil {
		panic(err)
	}
}

func execFileRemove() {
	fmt.Println("Os execute Remove dicom file")
	rmDicomCmd := exec.Command("bash", "-c", "rm /home/$USER/Download/*.dcm; rm /home/$USER/Download/*.zip")

	_, err := rmDicomCmd.Output()
	if err != nil {
		panic(err)
	}

}

func put(url string, client *http.Client) {
	//client := accessor()
	request, err := http.NewRequest("PUT", url, strings.NewReader("120"))
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	response.Body.Close()
}

func metadataAcess() {
	fmt.Println("Requesting....")
	db := accessor()
	reqst1, err := db.Get("http://0.0.0.0/orthanc/studies")

	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(reqst1.Body)
	if err != nil {
		log.Fatal(err)
	}
	reqst1.Body.Close()

	// Converting received data to array string for iterability
	fmt.Println("*************** Converting first request ***************")
	var studyIdList []string
	_ = json.Unmarshal([]byte(responseData), &studyIdList)

	hasStudy := len(studyIdList)

	fmt.Println("The number of study is: ", hasStudy)
	if hasStudy > 0 {
		// DO something meaningfull
		for _, element := range studyIdList {
			fmt.Println("This study ID is: ", element)
			time.Sleep(time.Second)

			// Make second request for metadata contained in each study
			reqst2, err := db.Get("http://0.0.0./orthanc/studies/" + element + "/metadata")
			if err != nil {
				log.Fatal(err)
			}
			responseMetadata, err := ioutil.ReadAll(reqst2.Body)
			if err != nil {
				log.Fatal(err)
			}
			reqst2.Body.Close()

			// Converting second request to array string for iterability
			fmt.Println("*************** Converting second request ***************")
			var metadataList []string
			_ = json.Unmarshal([]byte(responseMetadata), &metadataList)

			// Dont forget to change metadata to the HashValue
			hasmetadataState, _ := contains(metadataList, "HashState") // false
			if !hasmetadataState {                                     // false -> true
				// Do something with the metadata value
				fmt.Println("it does not have metadata state")
				fmt.Println("But it does have metadata LastUpdate")

				hasmetadataHash, metadataName := contains(metadataList, "LastUpdate") // true
				if hasmetadataHash {
					fmt.Println("LastUpdate found with name ", metadataName)
					reqst3, err := db.Get("http://0.0.0.0/orthanc/studies/" + element + "/metadata/" + metadataName)
					if err != nil {
						log.Fatal(err)
					}
					metaContent, err := ioutil.ReadAll(reqst3.Body)
					if err != nil {
						log.Fatal(err)
					}
					reqst3.Body.Close()
					hash := string(metaContent)
					fmt.Println("The hash value is: ", hash)
					fmt.Println("Os call execute command")
					// 1-Exec_command to Download file from swarm to specific directory
					//execDownload(hash)
					// 2-From chosen directory, unzip file in
					//execExtract()
					// 3-Send All dicom file to local rohan
					//execTransfer()
					// 4-Delete the previous file already sent
					//execFileRemove()
					// Post
					url := "http://0.0.0.0/orthanc/studies/" + element + "/metadata/HashState"
					put(url, db)
					//reqst4, err := db.PostForm("http://192.168.0.52/orthanc/studies/"+element+"/metadata/HashState", "1234")

				} else {
					// Do something with that
					fmt.Println("It does not have the metadata LastUpdate")
					// Then continue to next/ break to check
				}
			} else {
				// go to the next study and check for the required metadata
				fmt.Println("It has a metadata state")
				//continue
			}

		}
	} else {
		// is an empty list of studies
		fmt.Println("Is an empty list")
		// refresh the list by calling the login perhaps
	}

}

func main() {
	metadataAcess()

}
